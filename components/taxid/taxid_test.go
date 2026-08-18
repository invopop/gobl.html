package taxid

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal"
	"github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

// localized prepares a context with the locales loaded the same way as
// goblhtml does, with "en" merged into every other locale.
func localized(t *testing.T, code string) context.Context {
	t.Helper()
	locs := new(i18n.Locales)
	if err := locs.LoadWithDefault(locales.Content, "en"); err != nil {
		t.Fatal(err)
	}
	l := locs.Match(code)
	if l == nil {
		t.Fatalf("unknown locale %q", code)
	}
	return l.WithContext(context.Background())
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name    string
		id      *tax.Identity
		label   string
		code    string
		country string
	}{
		{
			name:    "national identifier without prefix",
			id:      &tax.Identity{Country: "ES", Code: "B98602642"},
			label:   "NIF",
			code:    "B98602642",
			country: "ES",
		},
		{
			name:  "prefix part of the official identifier",
			id:    &tax.Identity{Country: "DE", Code: "111111125"},
			label: "USt-IdNr.",
			code:  "DE111111125",
		},
		{
			name:  "French VAT number",
			id:    &tax.Identity{Country: "FR", Code: "44732829320"},
			label: "TVA",
			code:  "FR44732829320",
		},
		{
			name: "Swiss UID grouping",
			id:   &tax.Identity{Country: "CH", Code: "E123456789"},
			code: "CHE-123.456.789",
		},
		{
			name: "Norwegian VAT suffix",
			id:   &tax.Identity{Country: "NO", Code: "123456785MVA"},
			code: "NO 123456785 MVA",
		},
		{
			name:    "Colombian NIT grouping",
			id:      &tax.Identity{Country: "CO", Code: "9015852843"},
			label:   "NIT",
			code:    "901.585.284-3",
			country: "CO",
		},
		{
			name:    "Greek identities use the EL tax country code",
			id:      &tax.Identity{Country: "EL", Code: "177472438"},
			code:    "177472438",
			country: "EL",
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			f := Format(localized(t, "en"), ts.id)
			if ts.label != "" && f.Label != ts.label {
				t.Errorf("label: got %q, want %q", f.Label, ts.label)
			}
			if f.Code != ts.code {
				t.Errorf("code: got %q, want %q", f.Code, ts.code)
			}
			if f.Country != ts.country {
				t.Errorf("country: got %q, want %q", f.Country, ts.country)
			}
		})
	}
}

func TestFormatNil(t *testing.T) {
	if f := Format(localized(t, "en"), nil); f != (Formatted{}) {
		t.Errorf("expected empty result, got %+v", f)
	}
}

func TestFormatLabelLocalized(t *testing.T) {
	// Countries without a name of their own take the generic label of the
	// language the document is rendered in.
	generic := &tax.Identity{Country: "SE", Code: "123456789001"}
	if l := Format(localized(t, "en"), generic).Label; l != "TIN" {
		t.Errorf("en: got %q", l)
	}
	if l := Format(localized(t, "es"), generic).Label; l != "Código fiscal" {
		t.Errorf("es: got %q", l)
	}

	// Names defined for a country are used in every language until a locale
	// provides its own.
	named := &tax.Identity{Country: "DE", Code: "111111125"}
	if l := Format(localized(t, "es"), named).Label; l != "USt-IdNr." {
		t.Errorf("es: got %q", l)
	}
}

func TestFormatIntraCommunity(t *testing.T) {
	invoice := func(supplier, customer string) context.Context {
		inv := &bill.Invoice{
			IssueDate: cal.MakeDate(2025, 1, 15),
			Supplier:  &org.Party{TaxID: mustParse(t, supplier)},
			Customer:  &org.Party{TaxID: mustParse(t, customer)},
		}
		return internal.WithDocument(localized(t, "en"), internal.DocumentFor(inv))
	}

	t.Run("supply between member states", func(t *testing.T) {
		ctx := invoice("ESB98602642", "NL000099995B57")
		if f := Format(ctx, mustParse(t, "ESB98602642")); f.Code != "ESB98602642" || f.Country != "" {
			t.Errorf("got %+v", f)
		}
	})

	t.Run("domestic supply keeps the national format", func(t *testing.T) {
		ctx := invoice("ESB98602642", "ESB63272603")
		if f := Format(ctx, mustParse(t, "ESB98602642")); f.Code != "B98602642" || f.Country != "ES" {
			t.Errorf("got %+v", f)
		}
	})

	t.Run("export outside the union keeps the national format", func(t *testing.T) {
		ctx := invoice("ESB98602642", "US123456789")
		if f := Format(ctx, mustParse(t, "ESB98602642")); f.Code != "B98602642" || f.Country != "ES" {
			t.Errorf("got %+v", f)
		}
	})

	t.Run("third party outside the union keeps the national format", func(t *testing.T) {
		ctx := invoice("ESB98602642", "NL000099995B57")
		if f := Format(ctx, &tax.Identity{Country: "SG", Code: "199307558M"}); f.Country != "SG" {
			t.Errorf("got %+v", f)
		}
	})
}

func mustParse(t *testing.T, tin string) *tax.Identity {
	t.Helper()
	id, err := tax.ParseIdentity(tin)
	if err != nil {
		t.Fatalf("parsing %s: %s", tin, err)
	}
	return id
}
