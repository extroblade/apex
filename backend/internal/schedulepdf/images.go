package schedulepdf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
)

// bannerName is the season_assets key + on-disk filename stem for the banner.
const bannerName = "season-banner"

// saveBanner extracts the largest embedded JPEG (the schedule header banner) and
// writes it to ImageDir, recording its path in season_assets. It's idempotent:
// an unchanged banner (same hash) is left in place. A no-op when ImageDir is unset.
func (in *Ingestor) saveBanner(ctx context.Context, pdf []byte) error {
	if in.ImageDir == "" {
		return nil
	}
	imgs := extractJPEGs(pdf)
	if len(imgs) == 0 {
		return nil
	}
	banner := imgs[0]
	sum := sha256.Sum256(banner)
	hash := hex.EncodeToString(sum[:])

	var have string
	_ = in.DB.QueryRowContext(ctx,
		`SELECT sha256 FROM season_assets WHERE name = ?`, bannerName).Scan(&have)
	if have == hash {
		return nil // unchanged
	}

	if err := os.MkdirAll(in.ImageDir, 0o755); err != nil {
		return err
	}
	file := bannerName + ".jpg"
	if err := os.WriteFile(filepath.Join(in.ImageDir, file), banner, 0o644); err != nil {
		return err
	}
	if _, err := in.DB.ExecContext(ctx, `
		INSERT INTO season_assets (name, path, sha256) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE path = VALUES(path), sha256 = VALUES(sha256)`,
		bannerName, file, hash); err != nil {
		return err
	}
	log.Printf("schedulepdf: saved banner %s (%d bytes)", file, len(banner))
	return nil
}

// JPEG start-of-image / end-of-image markers. A DCTDecode image stream in a PDF
// *is* a JPEG verbatim (the filter name just declares the encoding), so we can
// recover embedded JPEGs by scanning for these markers. We do this rather than
// go through the pdf library's stream reader, which panics on DCTDecode.
var (
	jpegSOI = []byte{0xFF, 0xD8, 0xFF}
	jpegEOI = []byte{0xFF, 0xD9}
)

// minJPEGBytes filters out tiny fragments that are almost never real artwork.
const minJPEGBytes = 1024

// extractJPEGs scans raw PDF bytes for embedded JPEG images and returns each
// one's bytes, largest first. iRacing's season PDF carries a single header
// banner this way; there are no per-series raster logos to pull.
func extractJPEGs(data []byte) [][]byte {
	var out [][]byte
	for i := 0; i+len(jpegSOI) <= len(data); {
		start := bytes.Index(data[i:], jpegSOI)
		if start < 0 {
			break
		}
		start += i
		end := bytes.Index(data[start:], jpegEOI)
		if end < 0 {
			break
		}
		end += start + len(jpegEOI)
		if end-start >= minJPEGBytes {
			img := make([]byte, end-start)
			copy(img, data[start:end])
			out = append(out, img)
		}
		i = end
	}
	// Largest first — the banner dominates, fragments (if any) trail.
	for a := 0; a < len(out); a++ {
		for b := a + 1; b < len(out); b++ {
			if len(out[b]) > len(out[a]) {
				out[a], out[b] = out[b], out[a]
			}
		}
	}
	return out
}
