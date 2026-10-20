package bill

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	srclocales "github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvanceText(t *testing.T) {
	locales := new(i18n.Locales)
	require.NoError(t, locales.LoadWithDefault(srclocales.Content, "en"))

	ctxFor := func(code string) context.Context {
		l := locales.Get(i18n.Code(code))
		require.NotNil(t, l)
		return i18n.WithScope(l.WithContext(context.Background()), "billing.invoice.totals")
	}

	date := cal.MakeDate(2026, 1, 2)

	t.Run("uses the description with ref and date", func(t *testing.T) {
		adv := &pay.Record{Description: "Deposit", Ref: "DEP-1", Date: &date}
		assert.Equal(t, "2026-01-02 Deposit (DEP-1)", advanceText(ctxFor("en"), adv))
	})

	t.Run("falls back to a translated label without description", func(t *testing.T) {
		assert.Equal(t, "Advance", advanceText(ctxFor("en"), &pay.Record{Key: pay.MeansKeyCard}))
		assert.Equal(t, "Anticipo", advanceText(ctxFor("es"), &pay.Record{}))
		assert.Equal(t, "2026-01-02 Advance (DEP-1)", advanceText(ctxFor("en"), &pay.Record{Ref: "DEP-1", Date: &date}))
	})

	t.Run("fallback label exists in every locale", func(t *testing.T) {
		for _, code := range srclocales.Codes() {
			txt := advanceText(ctxFor(code), &pay.Record{})
			assert.NotEmpty(t, txt, "locale %q", code)
			assert.NotContains(t, txt, "!(MISSING", "locale %q", code)
		}
	})
}
