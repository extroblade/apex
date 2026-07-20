package contentsync

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// mediaRoot is where catalog images are rehosted, served by the static nginx
// under /media/catalog (the frontend nginx proxies /media → static). Defaults to
// /catalog so it maps cleanly to the served path.
const (
	defaultCatalogImageDir = "/catalog"
	catalogURLPrefix       = "/media/catalog/"
	mediaSubdir            = "catalog"
	rehostBatchLimit       = 500
	descriptionBatchLimit  = 500
	descriptionPause       = 200 * time.Millisecond
	descriptionMinLen      = 120
	// descriptionMaxLen caps the stored blurb: enough for a paragraph, but keeps
	// a runaway first-<p> fallback from overflowing the column (and the UI).
	descriptionMaxLen = 1000
)

var (
	metaDescriptionRe = regexp.MustCompile(`(?i)<meta\s+name=["']description["']\s+content=["']([^"']+)["']`)
	ogDescriptionRe   = regexp.MustCompile(`(?i)<meta\s+(?:property=["']og:description["']|name=["']og:description["'])\s+content=["']([^"']+)["']`)
	paragraphRe       = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	tagRe             = regexp.MustCompile(`<[^>]+>`)
	spaceRe           = regexp.MustCompile(`\s+`)
)

// rehostImages downloads every catalog image still pointing at a remote URL and
// stores it locally, rewriting image_path to the rehosted /media path. It runs
// as its own step (independent of the content-hash guard) by scanning the DB for
// http(s) image_path values, so it can fill in even when the catalog pages are
// unchanged. Track configs share an image (keyed by base track), so each file is
// downloaded once and every config row is repointed at it.
func (s *Syncer) rehostImages(ctx context.Context) error {
	dir := urlOr("CATALOG_IMAGE_DIR", defaultCatalogImageDir)
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(dir, "cars"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "tracks"), 0o755); err != nil {
		return err
	}

	// Cars: one row per car_id, image is unique.
	if err := s.rehostTable(ctx, dir, "cars", "car_id", "car_name"); err != nil {
		log.Printf("contentsync: rehost cars: %v", err)
	}
	// Tracks: configs share an image (same base track). Download once per distinct
	// remote URL, then repoint every config sharing it.
	if err := s.rehostTable(ctx, dir, "tracks", "track_id", "track_name"); err != nil {
		log.Printf("contentsync: rehost tracks: %v", err)
	}
	return nil
}

// rehostTable rehosts one table's images. idCol uniquely identifies a row for the
// UPDATE; nameCol is only informational for the chosen filename.
func (s *Syncer) rehostTable(ctx context.Context, dir, table, idCol, nameCol string) error {
	rows, err := s.DB.QueryContext(ctx,
		"SELECT "+idCol+", image_path FROM "+table+
			" WHERE image_path LIKE 'http%' ORDER BY "+idCol+
			" LIMIT ?", rehostBatchLimit)
	if err != nil {
		return err
	}
	type pending struct {
		id     int64
		remote string
	}
	var todo []pending
	for rows.Next() {
		var id int64
		var img string
		if err := rows.Scan(&id, &img); err != nil {
			rows.Close()
			return err
		}
		todo = append(todo, pending{id: id, remote: img})
	}
	rows.Close()
	if len(todo) == 0 {
		return nil
	}

	// Cache already-downloaded remote URLs → local path so a track's many configs
	// (which all reference the same base image) download exactly once.
	done := make(map[string]string)
	downloaded := 0
	for _, p := range todo {
		if local, ok := done[p.remote]; ok {
			if _, err := s.DB.ExecContext(ctx,
				"UPDATE "+table+" SET image_path = ? WHERE "+idCol+" = ?", local, p.id); err != nil {
				return err
			}
			continue
		}
		local, err := s.downloadImage(ctx, p.remote, dir, table)
		if err != nil {
			log.Printf("contentsync: download %s: %v", p.remote, err)
			continue // leave the row on its remote URL; try again next run
		}
		done[p.remote] = local
		if _, err := s.DB.ExecContext(ctx,
			"UPDATE "+table+" SET image_path = ? WHERE "+idCol+" = ?", local, p.id); err != nil {
			return err
		}
		downloaded++
	}
	log.Printf("contentsync: rehosted %d %s images (%d distinct)",
		downloaded, table, len(done))
	return nil
}

// downloadImage fetches a remote image and writes it under dir/<table>/<base>,
// returning the /media-served path. The filename is the URL basename so a shared
// base image dedupes naturally across track configs.
func (s *Syncer) downloadImage(ctx context.Context, remote, dir, table string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remote, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ContentPilot/1.0 (content sync)")
	req.Header.Set("Referer", "https://www.iracing.com/")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", resp.Body.Close()
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20)) // 8 MiB cap
	if err != nil {
		return "", err
	}

	base := path.Base(remote)
	base = strings.TrimPrefix(base, "/")
	if base == "" || base == "." || base == "/" {
		return "", errors.New("empty image basename")
	}
	dest := filepath.Join(dir, table, base)
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", err
	}
	// Served path: /media/catalog/<table>/<base>. The static container mounts the
	// volume at /usr/share/nginx/html/catalog, and the frontend nginx proxies
	// /media/* → static/*.
	return catalogURLPrefix + table + "/" + base, nil
}

// fillDescriptions fetches detail pages for catalog rows that have a detail_url
// but no description, extracting the blurb from the page. Like rehostImages it is
// independent of the content-hash guard — it scans the DB for the gap and fills
// it. It pauses between requests to be a polite citizen and caps the batch.
func (s *Syncer) fillDescriptions(ctx context.Context) error {
	for _, table := range []string{"cars", "tracks"} {
		if err := s.fillTable(ctx, table); err != nil {
			log.Printf("contentsync: descriptions %s: %v", table, err)
		}
	}
	return nil
}

func (s *Syncer) fillTable(ctx context.Context, table string) error {
	rows, err := s.DB.QueryContext(ctx,
		"SELECT detail_url FROM "+table+
			" WHERE detail_url <> '' AND description = '' LIMIT ?", descriptionBatchLimit)
	if err != nil {
		return err
	}
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return err
		}
		urls = append(urls, u)
	}
	rows.Close()
	if len(urls) == 0 {
		return nil
	}

	filled := 0
	for _, url := range urls {
		desc, err := s.fetchDescription(ctx, url)
		if err != nil {
			log.Printf("contentsync: description %s: %v", url, err)
			continue
		}
		if desc == "" {
			continue // leave the row for next run; no usable text on the page
		}
		if _, err := s.DB.ExecContext(ctx,
			"UPDATE "+table+" SET description = ? WHERE detail_url = ?", desc, url); err != nil {
			return err
		}
		filled++
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(descriptionPause):
		}
	}
	log.Printf("contentsync: filled %d %s descriptions (%d attempted)", filled, table, len(urls))
	return nil
}

// fetchDescription pulls a detail page and extracts its blurb, preferring the
// meta description tag, then og:description, finally the first long paragraph.
func (s *Syncer) fetchDescription(ctx context.Context, url string) (string, error) {
	data, err := s.fetchRaw(ctx, url)
	if err != nil {
		return "", err
	}
	html := string(data)
	if m := metaDescriptionRe.FindStringSubmatch(html); len(m) > 1 {
		if d := cleanText(htmlUnescape(m[1])); len(d) >= descriptionMinLen {
			return capDescription(d), nil
		}
	}
	if m := ogDescriptionRe.FindStringSubmatch(html); len(m) > 1 {
		if d := cleanText(htmlUnescape(m[1])); len(d) >= descriptionMinLen {
			return capDescription(d), nil
		}
	}
	// Fallback: first paragraph longer than the threshold.
	for _, m := range paragraphRe.FindAllStringSubmatch(html, -1) {
		if d := cleanText(htmlUnescape(m[1])); len(d) >= descriptionMinLen {
			return capDescription(d), nil
		}
	}
	return "", nil
}

// fetchRaw is a plain GET (no hash-guard) for detail pages. Bounded to 8 MiB.
func (s *Syncer) fetchRaw(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ContentPilot/1.0 (content sync)")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errStatus(resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

// cleanText strips tags, collapses whitespace, and trims the result.
func cleanText(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = htmlUnescape(s)
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// capDescription trims an over-long blurb to descriptionMaxLen runes, cutting at
// the last word boundary and appending an ellipsis. It counts runes (not bytes)
// so multi-byte text stays within the column's character limit.
func capDescription(s string) string {
	r := []rune(s)
	if len(r) <= descriptionMaxLen {
		return s
	}
	r = r[:descriptionMaxLen]
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] == ' ' {
			r = r[:i]
			break
		}
	}
	return strings.TrimSpace(string(r)) + "…"
}

// errStatus returns a typed error carrying an HTTP status code.
type statusError int

func (e statusError) Error() string {
	return "status " + strconv.Itoa(int(e))
}

func errStatus(code int) error { return statusError(code) }
