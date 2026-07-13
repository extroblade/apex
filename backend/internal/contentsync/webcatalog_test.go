package contentsync

import "testing"

// A trimmed-down version of the real iRacing cars-page markup: each card is a
// div with data-name/data-type, an onclick pointing at its detail page, and an
// img (a WordPress thumbnail URL).
const sampleCards = `
<div data-name="Next Gen NASCAR Cup Series Chevrolet Camaro ZL1" data-order="0" data-type="paid" data-timestamp="1620161122" class="grid-item" onclick="javascript: document.location = 'https://www.iracing.com/cars/next-gen-nascar-cup-series-chevrolet-camaro/';">
  <div class="thumbnail"><a href="x"><img src="https://s100.iracing.com/wp-content/uploads/2021/05/nascar-nextgen-camaro-feature-1024x576.jpg" class="wp-post-image" /></a></div>
</div>
<div data-name="Global Mazda MX-5 Cup" data-order="1" data-type="free" data-timestamp="1620161122" class="grid-item" onclick="javascript: document.location = 'https://www.iracing.com/cars/global-mazda-mx-5-cup/';">
  <div class="thumbnail"><a href="x"><img src="https://s100.iracing.com/wp-content/uploads/2021/05/mx5-feature-350x197.jpg" class="wp-post-image" /></a></div>
</div>`

func TestParseCatalogCards(t *testing.T) {
	items := parseCatalogCards(sampleCards)
	if len(items) != 2 {
		t.Fatalf("got %d cards, want 2", len(items))
	}
	if items[0].free || !items[1].free {
		t.Fatalf("free flags wrong: %+v", items)
	}
	// Thumbnail size suffix stripped to the full image.
	if items[0].image != "https://s100.iracing.com/wp-content/uploads/2021/05/nascar-nextgen-camaro-feature.jpg" {
		t.Fatalf("image not full-size: %q", items[0].image)
	}
	if items[1].name != "Global Mazda MX-5 Cup" {
		t.Fatalf("name wrong: %q", items[1].name)
	}
	// Detail URL pulled from the card's onclick.
	if items[0].detailURL != "https://www.iracing.com/cars/next-gen-nascar-cup-series-chevrolet-camaro/" {
		t.Fatalf("detail url wrong: %q", items[0].detailURL)
	}
}

func TestNormalizeName(t *testing.T) {
	cases := map[string]string{
		"Nürburgring Grand-Prix": "nurburgringgrandprix",
		"nurburgring grand prix": "nurburgringgrandprix",
		"Mazda MX-5 Cup":         "mazdamx5cup",
	}
	for in, want := range cases {
		if got := normalizeName(in); got != want {
			t.Errorf("normalizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFullSizeImage(t *testing.T) {
	got := fullSizeImage("https://x/y/foo-feature-768x432.jpg")
	if got != "https://x/y/foo-feature.jpg" {
		t.Fatalf("fullSizeImage = %q", got)
	}
	// No suffix — unchanged.
	if fullSizeImage("https://x/y/foo.png") != "https://x/y/foo.png" {
		t.Fatal("unexpected change to suffix-less url")
	}
}
