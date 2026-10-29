package bill

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	srclocales "github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentMethodName(t *testing.T) {
	locales := new(i18n.Locales)
	require.NoError(t, locales.LoadWithDefault(srclocales.Content, "en"))

	ctxFor := func(code string) context.Context {
		l := locales.Get(i18n.Code(code))
		require.NotNil(t, l)
		return i18n.WithScope(l.WithContext(context.Background()), "billing.payment.methods.instructions")
	}

	t.Run("uses the translated label of the key", func(t *testing.T) {
		assert.Equal(t, "Bank card", paymentMethodName(ctxFor("en"), &pay.Record{Key: pay.MeansKeyCard}))
		assert.Equal(t, "Cash", paymentMethodName(ctxFor("en"), &pay.Record{Key: pay.MeansKeyCash}))
		assert.Equal(t, "Transferência bancária", paymentMethodName(ctxFor("pt"), &pay.Record{Key: pay.MeansKeyCreditTransfer}))
	})

	t.Run("falls back to the GOBL definition, then the key", func(t *testing.T) {
		assert.Equal(t, "Promissory Note", paymentMethodName(ctxFor("en"), &pay.Record{Key: pay.MeansKeyPromissoryNote}))
		assert.Equal(t, "crypto", paymentMethodName(ctxFor("en"), &pay.Record{Key: "crypto"}))
	})

	t.Run("handles records without a key", func(t *testing.T) {
		assert.Empty(t, paymentMethodName(ctxFor("en"), &pay.Record{Description: "Paid in full"}))
		assert.Equal(t, "Advance", paymentMethodName(ctxFor("en"), &pay.Record{}))
		assert.Equal(t, "Anticipo", paymentMethodName(ctxFor("es"), &pay.Record{}))
	})

	t.Run("common labels exist in every locale", func(t *testing.T) {
		for _, code := range srclocales.Codes() {
			for _, rec := range []*pay.Record{{Key: pay.MeansKeyCard}, {Key: pay.MeansKeyCash}, {Key: pay.MeansKeyCreditTransfer}, {}} {
				txt := paymentMethodName(ctxFor(code), rec)
				assert.NotEmpty(t, txt, "locale %q key %q", code, rec.Key)
				assert.NotContains(t, txt, "!(MISSING", "locale %q key %q", code, rec.Key)
			}
		}
	})
}
