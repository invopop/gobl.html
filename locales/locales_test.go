package locales_test

import (
	"path"
	"strings"
	"testing"

	"github.com/invopop/gobl.html/locales"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// referenceCode is the locale every other locale is compared against. English
// is the source of truth for `app.yml` and `countries.yml`; it deliberately has
// no `currencies.yml`, so currency codes are compared against the union of what
// all the locales translate.
const referenceCode = "en"

// exemptKeys are English strings that are meant to stay as they are in every
// locale, either because they belong to a specific tax regime or because they
// are proper nouns in another language.
func exemptKey(key string) bool {
	switch {
	case strings.HasPrefix(key, "regimes."):
		// Regime titles are legal document names and stay in the language of
		// the regime that defines them.
		return true
	case strings.Contains(key, ".ext_map.mx-cfdi-"):
		// Mexican CFDI labels are kept in Spanish, as English does.
		return true
	case strings.HasSuffix(key, "identity_labels.it-fiscal-code"),
		strings.HasSuffix(key, "identity_labels.de-tax-number"):
		// "Codice fiscale" and "St. -Nr." are not translated.
		return true
	}
	return false
}

// knownGaps lists the keys a locale does not translate yet and which therefore
// fall back to English when rendered. They all pre-date this test.
//
// New locales must ship without any entry here, so please translate the key
// rather than adding to this list.
//
// Keys under `billing` are normalised to `billing.*.` because the same block is
// merged into the invoice, payment, delivery and order documents.
var knownGaps = map[string][]string{
	"da": {
		"billing.*.title.adjustment",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
	"de": {
		"billing.*.title.adjustment",
		"country_names.AW",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
	"el": {
		"billing.*.methods.instructions.methods.cheque",
		"billing.*.payment.instructions.methods.cheque",
		"billing.*.payment.terms.due_dates.other",
		"billing.*.title.adjustment",
		"org.party.person",
		"org.party.person_label",
	},
	"es": {
		"billing.*.payment.terms.due_dates.other",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
	"fr": {
		// The four `terms.keys` below are spelt with hyphens in fr/app.yml
		// ("due-date", "end-of-month", "advance"), so they never match.
		"billing.*.payment.terms.due_dates.other",
		"billing.*.payment.terms.keys.advanced",
		"billing.*.payment.terms.keys.due_date",
		"billing.*.payment.terms.keys.end_of_month",
		"billing.*.title.adjustment",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
	"it": {
		"billing.*.methods.instructions.methods.cheque",
		"billing.*.payment.instructions.methods.cheque",
		"billing.*.title.adjustment",
		"country_names.KY",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
		// Currency list stops at UYU.
		"currencies.UYW", "currencies.UZS", "currencies.VES", "currencies.VND",
		"currencies.VUV", "currencies.WST", "currencies.XAF", "currencies.XCD",
		"currencies.XOF", "currencies.XPF", "currencies.YER", "currencies.ZAR",
		"currencies.ZMW", "currencies.ZWL",
	},
	"nl": {
		"billing.*.title.adjustment",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
	"no": {
		"billing.*.title.adjustment",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
	"pl": {
		// `prices_include` is spelt `prices_include_tax` in pl/app.yml.
		"billing.*.methods.instructions.methods.cheque",
		"billing.*.payment.instructions.methods.cheque",
		"billing.*.payment.terms.due_dates.other",
		"billing.*.title.adjustment",
		"billing.*.totals.prices_include",
		"country_names.KY",
		"org.party.ext_map.co-dian-municipality",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
		// Currency list stops at UYU.
		"currencies.UYW", "currencies.UZS", "currencies.VES", "currencies.VND",
		"currencies.VUV", "currencies.WST", "currencies.XAF", "currencies.XCD",
		"currencies.XOF", "currencies.XPF", "currencies.YER", "currencies.ZAR",
		"currencies.ZMW", "currencies.ZWL",
	},
	"pt": {
		"billing.*.methods.instructions.methods.cheque",
		"billing.*.payment.instructions.methods.cheque",
		"billing.*.payment.terms.due_dates.other",
		"billing.*.title.adjustment",
		"org.party.ext_map.co-dian-municipality",
		"org.party.person",
		"org.party.person_label",
		"org.party.website",
		"org.party.website_label",
	},
}

// localeFiles are the files every locale is expected to provide, except for
// English which has no currency names of its own.
var localeFiles = []string{"app.yml", "countries.yml", "currencies.yml", "units.yml"}

func TestLocaleCodes(t *testing.T) {
	codes := locales.Codes()
	assert.Contains(t, codes, referenceCode)
	assert.Greater(t, len(codes), 1)
}

// TestLocaleFilesUseOwnCode ensures each file is keyed by the language code of
// the directory it lives in. A file copied from another locale and left with
// the original code would otherwise load silently under the wrong language.
func TestLocaleFilesUseOwnCode(t *testing.T) {
	for _, code := range locales.Codes() {
		for _, name := range localeFiles {
			if code == referenceCode && name == "currencies.yml" {
				continue // English intentionally has no currency names.
			}
			data, err := locales.Content.ReadFile(path.Join(code, name))
			require.NoError(t, err, "locale %q is missing %s", code, name)

			var root map[string]any
			require.NoError(t, yaml.Unmarshal(data, &root), "locale %q, file %s", code, name)
			require.Len(t, root, 1, "locale %q, file %s: expected a single root key", code, name)
			_, ok := root[code]
			assert.True(t, ok, "locale %q, file %s: root key is not %q", code, name, code)
		}
	}
}

// TestLocaleKeyCoverage checks that every locale translates the same keys as
// the reference locale, so that a new language cannot ship with keys that
// silently render in English.
func TestLocaleKeyCoverage(t *testing.T) {
	required := keysOf(t, referenceCode, "app.yml")
	for k, v := range keysOf(t, referenceCode, "countries.yml") {
		required[k] = v
	}
	// English has no currencies.yml, so take every currency any locale names.
	for _, code := range locales.Codes() {
		for k, v := range keysOf(t, code, "currencies.yml") {
			required[k] = v
		}
	}

	for _, code := range locales.Codes() {
		if code == referenceCode {
			continue
		}
		t.Run(code, func(t *testing.T) {
			gaps := make(map[string]bool, len(knownGaps[code]))
			for _, k := range knownGaps[code] {
				gaps[k] = true
			}

			have := keysOf(t, code, "app.yml")
			for _, name := range []string{"countries.yml", "currencies.yml"} {
				for k, v := range keysOf(t, code, name) {
					have[k] = v
				}
			}

			used := make(map[string]bool, len(gaps))
			for key := range required {
				if have[key] || exemptKey(key) {
					continue
				}
				norm := normaliseKey(key)
				if gaps[norm] {
					used[norm] = true
					continue
				}
				assert.Fail(t, "missing translation",
					"locale %q does not translate %q, so it renders in English", code, key)
			}

			for _, k := range knownGaps[code] {
				assert.True(t, used[k],
					"locale %q now translates %q: please remove it from knownGaps", code, k)
			}
		})
	}
}

// normaliseKey collapses the document type out of billing keys, which are
// defined once and merged into the invoice, payment, delivery and order
// documents.
func normaliseKey(key string) string {
	if !strings.HasPrefix(key, "billing.") {
		return key
	}
	parts := strings.SplitN(key, ".", 3)
	if len(parts) < 3 {
		return key
	}
	return "billing.*." + parts[2]
}

// keysOf returns the full set of leaf keys defined by a locale's file, with
// YAML anchors and merge keys resolved, and without the language code prefix.
func keysOf(t *testing.T, code, name string) map[string]bool {
	t.Helper()
	data, err := locales.Content.ReadFile(path.Join(code, name))
	if err != nil {
		return nil // handled by TestLocaleFilesUseOwnCode
	}
	var root map[string]any
	require.NoError(t, yaml.Unmarshal(data, &root), "locale %q, file %s", code, name)

	keys := make(map[string]bool)
	for _, content := range root {
		values, ok := content.(map[string]any)
		require.True(t, ok, "locale %q, file %s: expected a mapping", code, name)
		addKeys(keys, values, "")
	}
	return keys
}

func addKeys(keys map[string]bool, values map[string]any, prefix string) {
	for k, v := range values {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if sub, ok := v.(map[string]any); ok {
			addKeys(keys, sub, key)
			continue
		}
		keys[key] = true
	}
}
