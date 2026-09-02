package locales_test

import (
	"path"
	"slices"
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

// The files each locale is made up of.
const (
	appFile        = "app.yml"
	countriesFile  = "countries.yml"
	currenciesFile = "currencies.yml"
	unitsFile      = "units.yml"
)

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
	case strings.HasPrefix(key, "org.party.labels.") && key != "org.party.labels.default":
		// Country tax ID abbreviations such as NIF, RFC and P.IVA are used as
		// they are in every language. Only the generic label is translated.
		return true
	}
	return false
}

// Gaps shared by several locales, named so that the table below stays readable
// and each key is written exactly once.
var (
	// personLabels and websiteLabels were added to English after most of the
	// locales were written.
	personLabels  = []string{"org.party.person", "org.party.person_label"}
	websiteLabels = []string{"org.party.website", "org.party.website_label"}

	// adjustmentTitle is the title used when rendering in adjustment mode.
	adjustmentTitle = []string{"billing.*.title.adjustment"}

	// chequePayment is reachable both through the payment instructions and
	// through the payment methods block, so it goes missing in pairs.
	chequePayment = []string{
		"billing.*.payment.instructions.methods.cheque",
		"billing.*.methods.instructions.methods.cheque",
	}

	// dueDatesPlural is the "other" plural form of the due date label. The
	// locales below define "many" instead, which the templates never ask for.
	dueDatesPlural = []string{"billing.*.payment.terms.due_dates.other"}

	// currencyTail is where these locales' currency lists stop, at UYU.
	currencyTail = []string{
		"currencies.UYW", "currencies.UZS", "currencies.VES", "currencies.VND",
		"currencies.VUV", "currencies.WST", "currencies.XAF", "currencies.XCD",
		"currencies.XOF", "currencies.XPF", "currencies.YER", "currencies.ZAR",
		"currencies.ZMW", "currencies.ZWL",
	}
)

// knownGaps lists the keys a locale does not translate yet and which therefore
// fall back to English when rendered. They all pre-date this test.
//
// New locales must ship without any entry here, so please translate the key
// rather than adding to this list. Removing the last gap from a locale is what
// "complete" looks like: ar, ca, eu, fi and gl have no entry.
//
// Keys under `billing` are normalised to `billing.*.` because the same block is
// merged into the invoice, payment, delivery and order documents.
var knownGaps = map[string][]string{
	"da": slices.Concat(adjustmentTitle, personLabels, websiteLabels),
	"de": append(slices.Concat(adjustmentTitle, personLabels, websiteLabels),
		"country_names.AW"),
	"el": slices.Concat(adjustmentTitle, chequePayment, dueDatesPlural, personLabels),
	"es": slices.Concat(dueDatesPlural, personLabels, websiteLabels),
	// The three `terms.keys` below are spelt with hyphens in fr/app.yml
	// ("due-date", "end-of-month", "advance"), so they never match.
	"fr": append(slices.Concat(adjustmentTitle, dueDatesPlural, personLabels, websiteLabels),
		"billing.*.payment.terms.keys.advanced",
		"billing.*.payment.terms.keys.due_date",
		"billing.*.payment.terms.keys.end_of_month"),
	"it": append(slices.Concat(adjustmentTitle, chequePayment, personLabels, websiteLabels, currencyTail),
		"country_names.KY"),
	"nl": slices.Concat(adjustmentTitle, personLabels, websiteLabels),
	"no": slices.Concat(adjustmentTitle, personLabels, websiteLabels),
	// `prices_include` is spelt `prices_include_tax` in pl/app.yml.
	"pl": append(slices.Concat(chequePayment, dueDatesPlural, personLabels, websiteLabels, currencyTail),
		"billing.*.totals.prices_include",
		"country_names.KY",
		"org.party.ext_map.co-dian-municipality"),
	"pt": append(slices.Concat(adjustmentTitle, chequePayment, dueDatesPlural, personLabels, websiteLabels),
		"org.party.ext_map.co-dian-municipality"),
}

// localeFiles are the files every locale is expected to provide, except for
// English which has no currency names of its own.
var localeFiles = []string{appFile, countriesFile, currenciesFile, unitsFile}

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
			if code == referenceCode && name == currenciesFile {
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
	required := keysOf(t, referenceCode, appFile)
	for k, v := range keysOf(t, referenceCode, countriesFile) {
		required[k] = v
	}
	// English has no currencies.yml, so take every currency any locale names.
	for _, code := range locales.Codes() {
		for k, v := range keysOf(t, code, currenciesFile) {
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

			have := keysOf(t, code, appFile)
			for _, name := range []string{countriesFile, currenciesFile} {
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
