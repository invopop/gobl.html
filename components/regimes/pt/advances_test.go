package pt

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n/i18n"
	srclocales "github.com/invopop/gobl.html/locales"
	saft "github.com/invopop/gobl.pt.saft/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/pay"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvanceMeans(t *testing.T) {
	locales := new(i18n.Locales)
	require.NoError(t, locales.LoadWithDefault(srclocales.Content, "en"))
	ctxFor := func(code string) context.Context {
		l := locales.Get(i18n.Code(code))
		require.NotNil(t, l)
		return l.WithContext(context.Background())
	}
	inv := &bill.Invoice{}
	inv.SetAddons(saft.V1)

	t.Run("names the SAF-T payment means in the document language", func(t *testing.T) {
		adv := &pay.Record{Key: pay.MeansKeyCash, Ext: tax.Extensions{}.Set(saft.ExtKeyPaymentMeans, "NU")}
		assert.Equal(t, "cash", AdvanceMeans(ctxFor("en"), inv, adv))
		assert.Equal(t, "numerário", AdvanceMeans(ctxFor("pt"), inv, adv))

		adv = &pay.Record{Key: pay.MeansKeyCard, Ext: tax.Extensions{}.Set(saft.ExtKeyPaymentMeans, "CC")}
		assert.Equal(t, "credit card", AdvanceMeans(ctxFor("en"), inv, adv))
		assert.Equal(t, "cartão crédito", AdvanceMeans(ctxFor("pt"), inv, adv))
	})

	t.Run("falls back to the code for unknown values", func(t *testing.T) {
		adv := &pay.Record{Ext: tax.Extensions{}.Set(saft.ExtKeyPaymentMeans, "XX")}
		assert.Equal(t, "XX", AdvanceMeans(ctxFor("en"), inv, adv))
	})

	t.Run("is empty without the extension", func(t *testing.T) {
		assert.Empty(t, AdvanceMeans(ctxFor("en"), inv, &pay.Record{Key: pay.MeansKeyCash}))
		assert.Empty(t, AdvanceMeans(ctxFor("en"), inv, nil))
	})

	t.Run("is empty without the pt-saft addon", func(t *testing.T) {
		adv := &pay.Record{Key: pay.MeansKeyCash, Ext: tax.Extensions{}.Set(saft.ExtKeyPaymentMeans, "NU")}
		assert.Empty(t, AdvanceMeans(ctxFor("en"), &bill.Invoice{}, adv))
		assert.Empty(t, AdvanceMeans(ctxFor("en"), nil, adv))
	})
}
