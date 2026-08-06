package org

import (
	"context"
	"strings"
	"testing"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/org"
)

func TestIdentityLabelSanity(t *testing.T) {
	if err := ctxi18n.Load(locales.Content); err != nil {
		t.Fatal(err)
	}
	ctx, err := ctxi18n.WithLocale(context.Background(), "es")
	if err != nil {
		t.Fatal(err)
	}

	var b strings.Builder
	err = Party(&org.Party{
		Name: "Jean Ferran",
		Identities: []*org.Identity{
			{Key: "passport", Code: "25HE94294"},
			{Key: "internal-ref", Code: "REF-123"},
		},
	}).Render(ctx, &b)
	if err != nil {
		t.Fatal(err)
	}
	out := b.String()
	t.Log(out)
	for _, want := range []string{"Pasaporte: 25HE94294", "Internal ref: REF-123"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, "!(") {
		t.Error("missing-translation marker leaked into output")
	}
}
