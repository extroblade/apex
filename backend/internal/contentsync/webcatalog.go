package contentsync

import (
	"context"
	"database/sql"
	"log"
	"regexp"
	"strings"
)

const (
	defaultCarsHTMLURL   = "https://www.iracing.com/cars/"
	defaultTracksHTMLURL = "https://www.iracing.com/tracks/"
)

// The iRacing web catalog renders one card per car / base track as:
//   <div data-name="…" data-order="N" data-type="free|paid"
//        onclick="document.location='https://www.iracing.com/cars/<slug>/';">
//     … <img src="…"> …
// We enrich our catalog from it: `data-type` is the authoritative free/included
// flag, the card image is the artwork the app displays, and the onclick target
// is the detail page (used later to fill missing descriptions). Cards are keyed
// by name (there are no ids in the markup), so we match to the catalog by name.
var (
	cardRe = regexp.MustCompile(`data-name="([^"]+)"[^>]*?data-type="(free|paid)"`)
	imgRe  = regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	// onclick="javascript: document.location = 'https://…/cars/<slug>/';"
	detailRe = regexp.MustCompile(`document\.location\s*=\s*'([^']+)'`)
	// A WordPress resize suffix like -1024x576 before the extension.
	sizeSuffixRe = regexp.MustCompile(`-\d+x\d+(\.(?:jpg|jpeg|png|webp))$`)
)

type webItem struct {
	name      string
	free      bool
	image     string
	detailURL string
}

// parseCatalogCards pulls {name, free, image, detailURL} from an iRacing
// cars/tracks page. The slice is bounded by each card to its next sibling so an
// image/detail URL can't leak across cards.
func parseCatalogCards(htmlDoc string) []webItem {
	locs := cardRe.FindAllStringSubmatchIndex(htmlDoc, -1)
	items := make([]webItem, 0, len(locs))
	for i, loc := range locs {
		name := htmlDoc[loc[2]:loc[3]]
		free := htmlDoc[loc[4]:loc[5]] == "free"
		end := len(htmlDoc)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		body := htmlDoc[loc[1]:end]
		image := ""
		if m := imgRe.FindStringSubmatch(body); m != nil {
			image = fullSizeImage(m[1])
		}
		detail := ""
		if m := detailRe.FindStringSubmatch(body); m != nil {
			detail = m[1]
		}
		items = append(items, webItem{name: htmlUnescape(name), free: free, image: image, detailURL: detail})
	}
	return items
}

// fullSizeImage drops the WordPress thumbnail suffix so we store the full image.
func fullSizeImage(url string) string {
	return sizeSuffixRe.ReplaceAllString(url, "$1")
}

func htmlUnescape(s string) string {
	r := strings.NewReplacer("&amp;", "&", "&#039;", "'", "&#39;", "'", "&quot;", `"`)
	return r.Replace(s)
}

// syncWebCatalog fetches an iRacing catalog page and updates matching rows'
// image_path and is_free by normalized name. table is "cars" or "tracks";
// nameCol / idCol name the columns to match/return.
func (s *Syncer) syncWebCatalog(ctx context.Context, url, table, nameCol string) error {
	data, hash, fresh, err := s.fetch(ctx, url)
	if err != nil || !fresh {
		return err
	}
	items := parseCatalogCards(string(data))
	if len(items) == 0 {
		return nil
	}

	// Index catalog names once, normalized, so matching is a map lookup.
	byName, err := s.nameIndex(ctx, "SELECT "+nameCol+" FROM "+table)
	if err != nil {
		return err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	matched := 0
	for _, it := range items {
		name, ok := byName[normalizeName(it.name)]
		if !ok {
			continue
		}
		// Update every row sharing the name (a track's configs share track_name).
		// Each field is set only when the card supplied it, so an empty card image
		// never clobbers an existing (or rehosted) image_path.
		if _, err := tx.ExecContext(ctx,
			"UPDATE "+table+
				" SET is_free = ?, image_path = IF(? = '', image_path, ?),"+
				" detail_url = IF(? = '', detail_url, ?) WHERE "+nameCol+" = ?",
			it.free, it.image, it.image, it.detailURL, it.detailURL, name); err != nil {
			return err
		}
		matched++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("contentsync: web catalog %s — %d cards, %d matched", table, len(items), matched)
	return s.record(ctx, hash, "web:"+table+":"+url, len(items), matched)
}

// nameIndex maps normalized name -> exact DB name for a single-column query.
func (s *Syncer) nameIndex(ctx context.Context, query string) (map[string]string, error) {
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	idx := make(map[string]string)
	for rows.Next() {
		var name sql.NullString
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if name.Valid {
			idx[normalizeName(name.String)] = name.String
		}
	}
	return idx, rows.Err()
}

// normalizeName lowercases and strips everything but letters/digits so
// "Nürburgring Grand-Prix" and "nurburgring grand prix" match. (ASCII-fold of a
// few common accents keeps track names lined up with the web catalog.)
func normalizeName(s string) string {
	s = strings.ToLower(s)
	fold := strings.NewReplacer(
		"ü", "u", "ö", "o", "ä", "a", "é", "e", "è", "e", "á", "a", "ó", "o", "ñ", "n", "ç", "c",
	)
	s = fold.Replace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
