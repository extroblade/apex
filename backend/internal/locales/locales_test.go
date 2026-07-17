package locales

import (
	"encoding/json"
	"sort"
	"testing"
)

// flatten collects the dotted leaf-key paths of a nested translation bundle.
func flatten(prefix string, m map[string]any, out map[string]bool) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if child, ok := v.(map[string]any); ok {
			flatten(key, child, out)
		} else {
			out[key] = true
		}
	}
}

func keysOf(t *testing.T, code string) map[string]bool {
	t.Helper()
	raw, err := seedFS.ReadFile("data/" + code + ".json")
	if err != nil {
		t.Fatalf("read %s: %v", code, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s is not valid JSON: %v", code, err)
	}
	keys := map[string]bool{}
	flatten("", m, keys)
	return keys
}

// TestBuiltinBundlesEmbedded checks every built-in has an embedded bundle.
func TestBuiltinBundlesEmbedded(t *testing.T) {
	for _, b := range builtins {
		if _, err := seedFS.ReadFile("data/" + b.code + ".json"); err != nil {
			t.Errorf("no embedded bundle for built-in %q: %v", b.code, err)
		}
	}
}

// TestLocaleKeyParity ensures the shipped bundles carry exactly the same keys.
// ru.ts is typed against en at compile time; this guards the GENERATED JSON so a
// stale or partial bundle can't ship (missing keys would silently fall back to
// English at runtime).
func TestLocaleKeyParity(t *testing.T) {
	en := keysOf(t, "en")
	ru := keysOf(t, "ru")

	var missing, extra []string
	for k := range en {
		if !ru[k] {
			missing = append(missing, k)
		}
	}
	for k := range ru {
		if !en[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		t.Errorf("ru is missing %d key(s) present in en: %v", len(missing), missing)
	}
	if len(extra) > 0 {
		t.Errorf("ru has %d key(s) not in en: %v", len(extra), extra)
	}
}
