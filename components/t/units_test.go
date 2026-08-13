package t_test

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	ct "github.com/invopop/gobl.html/components/t"
	srclocales "github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitName(t *testing.T) {
	locales := new(i18n.Locales)
	require.NoError(t, locales.LoadWithDefault(srclocales.Content, "en"))

	ctxFor := func(code string) context.Context {
		l := locales.Get(i18n.Code(code))
		require.NotNil(t, l)
		return l.WithContext(context.Background())
	}

	t.Run("uses symbol when available", func(t *testing.T) {
		ctx := ctxFor("es")
		assert.Equal(t, "kg", ct.UnitName(ctx, org.UnitKilogram))
		assert.Equal(t, "m²", ct.UnitName(ctx, org.UnitSquareMetre))
	})

	t.Run("translates units without symbol", func(t *testing.T) {
		assert.Equal(t, "hours", ct.UnitName(ctxFor("en"), org.UnitHour))
		assert.Equal(t, "horas", ct.UnitName(ctxFor("es"), org.UnitHour))
		// German nouns keep their capitalization.
		assert.Equal(t, "Stunden", ct.UnitName(ctxFor("de"), org.UnitHour))
		assert.Equal(t, "unidades", ct.UnitName(ctxFor("es"), org.UnitUnit))
	})

	t.Run("passes through unknown UN/ECE codes", func(t *testing.T) {
		ctx := ctxFor("en")
		assert.Equal(t, "E48", ct.UnitName(ctx, org.Unit("E48")))
	})

	t.Run("all units resolve to a name in every locale", func(t *testing.T) {
		for _, code := range []string{"ar", "ca", "da", "de", "el", "en", "es", "eu", "fr", "gl", "it", "nl", "no", "pl", "pt"} {
			ctx := ctxFor(code)
			for _, def := range org.UnitDefinitions {
				name := ct.UnitName(ctx, def.Unit)
				assert.NotEmpty(t, name, "unit %q in locale %q", def.Unit, code)
				assert.NotContains(t, name, "!(MISSING", "unit %q in locale %q", def.Unit, code)
			}
		}
	})

	t.Run("every locale translates all symbol-less units", func(t *testing.T) {
		// Load without the default locale merge so that missing keys in a
		// locale's own units.yml are not masked by the English fallback.
		unmerged := new(i18n.Locales)
		require.NoError(t, unmerged.Load(srclocales.Content))

		for _, code := range []string{"ar", "ca", "da", "de", "el", "en", "es", "eu", "fr", "gl", "it", "nl", "no", "pl", "pt"} {
			l := unmerged.Get(i18n.Code(code))
			require.NotNil(t, l)
			for _, def := range org.UnitDefinitions {
				if def.Symbol != "" {
					continue
				}
				assert.True(t, l.Has("units."+string(def.Unit)), "missing translation for unit %q in locale %q", def.Unit, code)
			}
		}
	})
}

func TestItemUnitName(t *testing.T) {
	locales := new(i18n.Locales)
	require.NoError(t, locales.LoadWithDefault(srclocales.Content, "en"))
	ctx := locales.Get("en").WithContext(context.Background())

	tests := []struct {
		name string
		item *org.Item
		want string
	}{
		{name: "nil item", want: ""},
		{name: "empty item", item: &org.Item{}, want: ""},
		{name: "GOBL unit", item: &org.Item{Unit: org.UnitKilogram}, want: "kg"},
		{
			name: "GOBL and UNTDID units",
			item: &org.Item{
				Unit: org.UnitKilogram,
				Ext:  tax.MakeExtensions().Set("untdid-unit", "KGM"),
			},
			want: "kg (KGM)",
		},
		{
			name: "translated GOBL and UNTDID units",
			item: &org.Item{
				Unit: org.UnitHour,
				Ext:  tax.MakeExtensions().Set("untdid-unit", "HUR"),
			},
			want: "hours (HUR)",
		},
		{
			name: "UNTDID unit only",
			item: &org.Item{Ext: tax.MakeExtensions().Set("untdid-unit", "E48")},
			want: "E48",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, ct.ItemUnitName(ctx, test.item))
			assert.Equal(t, test.want != "", ct.ItemHasUnit(test.item))
		})
	}
}
