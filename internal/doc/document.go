// Package doc provides an interface to handle GOBL documents generically.
package doc

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

// Document is the interface to handle GOBL documents generically.
type Document interface {
	Extract() any
	GetType() cbc.Key
	GetSupplier() *org.Party
	GetIssueDate() cal.Date
	GetSeries() cbc.Code
	GetCode() cbc.Code
	GetRegime() tax.Regime
	GetCurrency() currency.Code
	GetExt() tax.Extensions
}

// For returns a Document interface for the given document.
func For(doc any) Document {
	switch doc := doc.(type) {
	case *bill.Invoice:
		return (*invoiceDoc)(doc)
	case *bill.Delivery:
		return (*deliveryDoc)(doc)
	case *bill.Order:
		return (*orderDoc)(doc)
	case *bill.Payment:
		return (*paymentDoc)(doc)
	default:
		return nil
	}
}

// Alias types to implement the Document interface.
type invoiceDoc bill.Invoice
type deliveryDoc bill.Delivery
type orderDoc bill.Order
type paymentDoc bill.Payment

// Document interface implementation for the invoice wrapper
func (i *invoiceDoc) Extract() any               { return (*bill.Invoice)(i) }
func (i *invoiceDoc) GetType() cbc.Key           { return i.Type }
func (i *invoiceDoc) GetSupplier() *org.Party    { return i.Supplier }
func (i *invoiceDoc) GetIssueDate() cal.Date     { return i.IssueDate }
func (i *invoiceDoc) GetSeries() cbc.Code        { return i.Series }
func (i *invoiceDoc) GetCode() cbc.Code          { return i.Code }
func (i *invoiceDoc) GetRegime() tax.Regime      { return i.Regime }
func (i *invoiceDoc) GetCurrency() currency.Code { return i.Currency }
func (i *invoiceDoc) GetExt() tax.Extensions {
	if i.Tax == nil {
		return nil
	}
	return i.Tax.Ext
}

// Document interface implementation for the delivery wrapper
func (d *deliveryDoc) Extract() any               { return (*bill.Delivery)(d) }
func (d *deliveryDoc) GetType() cbc.Key           { return d.Type }
func (d *deliveryDoc) GetSupplier() *org.Party    { return d.Supplier }
func (d *deliveryDoc) GetIssueDate() cal.Date     { return d.IssueDate }
func (d *deliveryDoc) GetSeries() cbc.Code        { return d.Series }
func (d *deliveryDoc) GetCode() cbc.Code          { return d.Code }
func (d *deliveryDoc) GetRegime() tax.Regime      { return d.Regime }
func (d *deliveryDoc) GetCurrency() currency.Code { return d.Currency }
func (d *deliveryDoc) GetExt() tax.Extensions {
	if d.Tax == nil {
		return nil
	}
	return d.Tax.Ext
}

// Document interface implementation for the order wrapper
func (o *orderDoc) Extract() any               { return (*bill.Order)(o) }
func (o *orderDoc) GetType() cbc.Key           { return o.Type }
func (o *orderDoc) GetSupplier() *org.Party    { return o.Supplier }
func (o *orderDoc) GetIssueDate() cal.Date     { return o.IssueDate }
func (o *orderDoc) GetSeries() cbc.Code        { return o.Series }
func (o *orderDoc) GetCode() cbc.Code          { return o.Code }
func (o *orderDoc) GetRegime() tax.Regime      { return o.Regime }
func (o *orderDoc) GetCurrency() currency.Code { return o.Currency }
func (o *orderDoc) GetExt() tax.Extensions {
	if o.Tax == nil {
		return nil
	}
	return o.Tax.Ext
}

// Document interface implementation for the payment wrapper
func (p *paymentDoc) Extract() any               { return (*bill.Payment)(p) }
func (p *paymentDoc) GetType() cbc.Key           { return p.Type }
func (p *paymentDoc) GetSupplier() *org.Party    { return p.Supplier }
func (p *paymentDoc) GetIssueDate() cal.Date     { return p.IssueDate }
func (p *paymentDoc) GetSeries() cbc.Code        { return p.Series }
func (p *paymentDoc) GetCode() cbc.Code          { return p.Code }
func (p *paymentDoc) GetRegime() tax.Regime      { return p.Regime }
func (p *paymentDoc) GetCurrency() currency.Code { return p.Currency }
func (p *paymentDoc) GetExt() tax.Extensions     { return p.Ext }
