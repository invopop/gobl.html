package org

import (
	"context"
	"strings"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/org"
)

func TestIdentityLabels(t *testing.T) {
	// Load locales the same way as goblhtml, with "en" merged into
	// every other locale so untranslated keys fall back to English.
	locs := new(i18n.Locales)
	if err := locs.LoadWithDefault(locales.Content, "en"); err != nil {
		t.Fatal(err)
	}
	ctx := locs.Match("es").WithContext(context.Background())

	var b strings.Builder
	err := Party(&org.Party{
		Name: "Jean Ferran",
		Identities: []*org.Identity{
			{Key: "passport", Code: "25HE94294"},
			{Key: "it-fiscal-code", Code: "RSSGNN60R30H501U"},
			{Key: "internal-ref", Code: "REF-123"},
		},
	}).Render(ctx, &b)
	if err != nil {
		t.Fatal(err)
	}
	out := b.String()
	t.Log(out)
	for _, want := range []string{
		// translated in the "es" locale
		"Pasaporte: 25HE94294",
		// missing in "es", falls back to the "en" locale
		"Codice fiscale: RSSGNN60R30H501U",
		// no translation anywhere, sentence-cased key
		"Internal ref: REF-123",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, "!(") {
		t.Error("missing-translation marker leaked into output")
	}
}
