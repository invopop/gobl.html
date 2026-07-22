package t_test

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	ct "github.com/invopop/gobl.html/components/t"
	srclocales "github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/org"
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
		assert.Equal(t, "Hours", ct.UnitName(ctxFor("en"), org.UnitHour))
		assert.Equal(t, "Horas", ct.UnitName(ctxFor("es"), org.UnitHour))
		assert.Equal(t, "Stunden", ct.UnitName(ctxFor("de"), org.UnitHour))
		assert.Equal(t, "Unidades", ct.UnitName(ctxFor("es"), org.UnitUnit))
	})

	t.Run("passes through unknown UN/ECE codes", func(t *testing.T) {
		ctx := ctxFor("en")
		assert.Equal(t, "E48", ct.UnitName(ctx, org.Unit("E48")))
	})

	t.Run("all units resolve to a name in every locale", func(t *testing.T) {
		for _, code := range []string{"ar", "ca", "de", "el", "en", "es", "eu", "fr", "gl", "it", "pl", "pt"} {
			ctx := ctxFor(code)
			for _, def := range org.UnitDefinitions {
				name := ct.UnitName(ctx, def.Unit)
				assert.NotEmpty(t, name, "unit %q in locale %q", def.Unit, code)
				assert.NotContains(t, name, "!(MISSING", "unit %q in locale %q", def.Unit, code)
			}
		}
	})
}
