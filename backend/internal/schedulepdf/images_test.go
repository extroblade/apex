package schedulepdf

import (
	"bytes"
	"testing"
)

func TestExtractJPEGs(t *testing.T) {
	// A valid-looking JPEG (SOI ... EOI) padded past the min-size guard.
	big := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x00}, 2000)...)
	big = append(big, 0xFF, 0xD9)
	small := append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x01}, 0xFF, 0xD9) // below min

	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"none", []byte("no images here, just %PDF text"), 0},
		{"one banner", append([]byte("%PDF-1.7\n"), big...), 1},
		{"tiny fragment ignored", append([]byte("junk"), small...), 0},
		{"two, largest first", append(append([]byte{}, small...), append(big, big...)...), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJPEGs(tt.data)
			if len(got) != tt.want {
				t.Fatalf("extractJPEGs() = %d images, want %d", len(got), tt.want)
			}
			for i := 1; i < len(got); i++ {
				if len(got[i-1]) < len(got[i]) {
					t.Fatalf("images not sorted largest-first: %d < %d", len(got[i-1]), len(got[i]))
				}
			}
			for _, img := range got {
				if !bytes.HasPrefix(img, jpegSOI) || !bytes.HasSuffix(img, jpegEOI) {
					t.Fatalf("extracted image is not a full SOI..EOI JPEG")
				}
			}
		})
	}
}
