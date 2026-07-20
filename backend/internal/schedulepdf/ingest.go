package schedulepdf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// Report summarizes one import.
type Report struct {
	Source        string
	SeriesFound   int
	SeriesMatched int
	WeeksUpdated  int
}

// Ingestor checks sources for new schedule PDFs and imports them.
type Ingestor struct {
	DB   *sql.DB
	HTTP *http.Client
	// URLs are explicit candidates (env-provided); generated wp-content
	// candidates for the current season are probed as well.
	URLs []string
	// Dir is a drop folder scanned for *.pdf files (manual downloads).
	Dir string
	// ImageDir, when set, receives artwork extracted from the PDF (the header
	// banner). Point it at the static service's media root to serve it.
	ImageDir string
}

// Run checks every source once. Errors on individual sources are logged, not
// fatal — the next scheduled run retries.
func (in *Ingestor) Run(ctx context.Context) {
	for _, u := range append(in.URLs, candidateURLs(time.Now())...) {
		if u == "" {
			continue
		}
		if err := in.ingestURL(ctx, u); err != nil {
			log.Printf("schedulepdf: %s: %v", u, err)
		}
	}
	if in.Dir != "" {
		files, _ := filepath.Glob(filepath.Join(in.Dir, "*.pdf"))
		for _, f := range files {
			if err := in.ingestFile(ctx, f); err != nil {
				log.Printf("schedulepdf: %s: %v", f, err)
			}
		}
	}
}

// candidateURLs guesses the official wp-content locations for the current
// season (quarterly: S1 Dec–Feb, S2 Mar–May, S3 Jun–Aug, S4 Sep–Nov).
func candidateURLs(now time.Time) []string {
	year := now.Year()
	season := int(now.Month()-1)/3 + 1 // 1..4 by quarter
	var urls []string
	for m := 1; m <= 12; m++ {
		urls = append(urls,
			fmt.Sprintf("https://www.iracing.com/wp-content/uploads/%d/%02d/%d-Season-%d-Full-Schedule.pdf",
				year, m, year, season))
	}
	return urls
}

func (in *Ingestor) ingestURL(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ContentPilot/1.0 (schedule sync)")
	resp, err := in.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return fmt.Errorf("not a PDF (%s)", resp.Header.Get("Content-Type"))
	}
	return in.ingest(ctx, url, data)
}

func (in *Ingestor) ingestFile(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return fmt.Errorf("not a PDF")
	}
	return in.ingest(ctx, "file:"+filepath.Base(path), data)
}

// ingest parses and applies a PDF unless its hash was already imported.
func (in *Ingestor) ingest(ctx context.Context, source string, data []byte) error {
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	var exists int
	if err := in.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schedule_imports WHERE sha256 = ?`, hash).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil // already imported
	}

	lines, err := extractLines(data)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	schedules := ParseText(lines)
	report, err := in.apply(ctx, source, schedules)
	if err != nil {
		return err
	}

	// Best-effort: pull the header banner. The PDF has no per-series logos, so
	// series/car/track artwork comes from the web catalog (see internal/contentsync).
	if err := in.saveBanner(ctx, data); err != nil {
		log.Printf("schedulepdf: banner: %v", err)
	}

	_, err = in.DB.ExecContext(ctx, `
		INSERT INTO schedule_imports (sha256, source, series_found, series_matched, weeks_updated)
		VALUES (?, ?, ?, ?, ?)`,
		hash, source, report.SeriesFound, report.SeriesMatched, report.WeeksUpdated)
	if err != nil {
		return err
	}
	log.Printf("schedulepdf: imported %s — %d series found, %d matched, %d weeks updated",
		source, report.SeriesFound, report.SeriesMatched, report.WeeksUpdated)
	return nil
}

// apply matches parsed series/track names to the catalog by normalized name and
// upserts season_schedule rows for the matches.
func (in *Ingestor) apply(ctx context.Context, source string, schedules []SeriesSchedule) (Report, error) {
	report := Report{Source: source, SeriesFound: len(schedules)}

	seriesByName, err := in.nameIndex(ctx, `SELECT series_id, series_name FROM series`)
	if err != nil {
		return report, err
	}
	tracksByName, err := in.nameIndex(ctx,
		`SELECT track_id, CONCAT(track_name, ' ', COALESCE(config_name,'')) FROM tracks`)
	if err != nil {
		return report, err
	}
	tracksShort, err := in.nameIndex(ctx, `SELECT track_id, track_name FROM tracks`)
	if err != nil {
		return report, err
	}

	tx, err := in.DB.BeginTx(ctx, nil)
	if err != nil {
		return report, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	for _, s := range schedules {
		seriesID, ok := seriesByName[Normalize(s.SeriesName)]
		if !ok {
			continue
		}
		report.SeriesMatched++
		for week, trackName := range s.Weeks {
			norm := Normalize(trackName)
			trackID, ok := tracksByName[norm]
			if !ok {
				trackID, ok = tracksShort[norm]
			}
			if !ok {
				trackID, ok = prefixMatch(tracksShort, norm)
			}
			if !ok {
				continue
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO season_schedule (series_id, week, track_id) VALUES (?, ?, ?)
				ON DUPLICATE KEY UPDATE track_id = VALUES(track_id)`,
				seriesID, week, trackID); err != nil {
				return report, err
			}
			report.WeeksUpdated++
		}
	}
	return report, tx.Commit()
}

func (in *Ingestor) nameIndex(ctx context.Context, query string) (map[string]int, error) {
	rows, err := in.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	idx := make(map[string]int)
	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		idx[Normalize(name)] = id
	}
	return idx, rows.Err()
}

// prefixMatch finds a unique catalog entry whose normalized name prefixes (or
// is prefixed by) the PDF name — tolerates config suffixes like "- Grand Prix".
func prefixMatch(idx map[string]int, name string) (int, bool) {
	found, id := 0, 0
	for k, v := range idx {
		if strings.HasPrefix(name, k) || strings.HasPrefix(k, name) {
			found++
			id = v
		}
	}
	return id, found == 1
}

// extractLines pulls plain-text rows out of the PDF.
func extractLines(data []byte) ([]string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var lines []string
	for p := 1; p <= reader.NumPage(); p++ {
		page := reader.Page(p)
		if page.V.IsNull() {
			continue
		}
		rows, err := page.GetTextByRow()
		if err != nil {
			continue // skip unreadable pages, keep the rest
		}
		for _, row := range rows {
			var b strings.Builder
			for _, word := range row.Content {
				if b.Len() > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(word.S)
			}
			lines = append(lines, b.String())
		}
	}
	return lines, nil
}
