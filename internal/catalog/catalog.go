// Package catalog describes the example documents so the previewer can filter
// them by country, document kind, addon and content.
package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/note"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
)

// Entry describes one example file.
type Entry struct {
	// File is the example's file name, such as "es-verifactu-invoice.json".
	File string `json:"file"`
	// Title is a readable name derived from the file name.
	Title string `json:"title"`
	// Country is the lower-case tax country code of the document, if any.
	Country string `json:"country,omitempty"`
	// Kind is the document kind: invoice, credit-note, debit-note, proforma,
	// corrective, order, delivery, payment, party or message.
	Kind string `json:"kind"`
	// Addons lists the tax addons the document declares.
	Addons []string `json:"addons,omitempty"`
	// Tags describe notable content: qr, payment, logo, discounts, charges,
	// delivery, notes, attachments, preceding, prices-include, multipage, and
	// the states void, sandbox and adjustment.
	Tags []string `json:"tags,omitempty"`
	// Currency is the document currency, if any.
	Currency string `json:"currency,omitempty"`
	// Supplier is the supplier's name, if any.
	Supplier string `json:"supplier,omitempty"`
	// Lines is the number of lines in the document.
	Lines int `json:"lines,omitempty"`
	// Locale is the rendering locale the example name implies, if any.
	Locale string `json:"locale,omitempty"`
	// Void, Sandbox and AdjustmentMode are the states the example name implies.
	Void           bool `json:"void,omitempty"`
	Sandbox        bool `json:"sandbox,omitempty"`
	AdjustmentMode bool `json:"adjustment_mode,omitempty"`
}

// Renderer produces the HTML for an envelope; it lets the scan detect content
// that only shows up when rendered, such as QR codes.
type Renderer func(ctx context.Context, env *gobl.Envelope) ([]byte, error)

// qrMarker matches the class names the regime components use for QR codes.
var qrMarker = regexp.MustCompile(`class="[^"]*\bqr\b`)

// Scan reads every *.json file in dir and describes it. Files that fail to
// parse are skipped with their error reported in errs.
func Scan(ctx context.Context, dir string, render Renderer) (entries []Entry, errs []error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, []error{err}
	}
	slices.Sort(matches)
	for _, path := range matches {
		e, err := describe(ctx, path, render)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(path), err))
			continue
		}
		entries = append(entries, e)
	}
	return entries, errs
}

func describe(ctx context.Context, path string, render Renderer) (Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, err
	}
	env := new(gobl.Envelope)
	if err := json.Unmarshal(data, env); err != nil {
		return Entry{}, err
	}

	name := filepath.Base(path)
	e := Entry{
		File:           name,
		Title:          title(name),
		Void:           strings.Contains(name, ".void."),
		Sandbox:        strings.Contains(name, ".sandbox."),
		AdjustmentMode: strings.Contains(name, ".adjustment."),
	}
	for _, l := range []string{"es", "fr", "pl"} {
		if strings.Contains(name, "."+l+".") {
			e.Locale = l
		}
	}

	var tags []string
	add := func(cond bool, tag string) {
		if cond {
			tags = append(tags, tag)
		}
	}

	switch doc := env.Extract().(type) {
	case *bill.Invoice:
		e.Kind = kindForInvoice(doc.Type)
		e.Country = country(doc.Regime)
		e.Addons = addons(doc.Addons)
		e.Currency = doc.Currency.String()
		e.Supplier = partyName(doc.Supplier)
		e.Lines = len(doc.Lines)
		add(doc.Payment != nil && (doc.Payment.Instructions != nil || doc.Payment.Terms != nil || len(doc.Payment.Advances) > 0), "payment")
		add(hasLogo(doc.Supplier), "logo")
		add(len(doc.Discounts) > 0, "discounts")
		add(len(doc.Charges) > 0, "charges")
		add(doc.Delivery != nil, "delivery")
		add(len(doc.Notes) > 0, "notes")
		add(len(doc.Attachments) > 0, "attachments")
		add(len(doc.Preceding) > 0, "preceding")
		add(doc.Tax != nil && doc.Tax.PricesInclude != "", "prices-include")
		add(len(doc.ExchangeRates) > 0, "exchange-rates")
	case *bill.Order:
		e.Kind = "order"
		e.Country = country(doc.Regime)
		e.Addons = addons(doc.Addons)
		e.Currency = doc.Currency.String()
		e.Supplier = partyName(doc.Supplier)
		e.Lines = len(doc.Lines)
		add(hasLogo(doc.Supplier), "logo")
		add(doc.Payment != nil, "payment")
		add(doc.Delivery != nil, "delivery")
		add(len(doc.Notes) > 0, "notes")
	case *bill.Delivery:
		e.Kind = "delivery"
		e.Country = country(doc.Regime)
		e.Addons = addons(doc.Addons)
		e.Currency = doc.Currency.String()
		e.Supplier = partyName(doc.Supplier)
		e.Lines = len(doc.Lines)
		add(hasLogo(doc.Supplier), "logo")
		add(len(doc.Notes) > 0, "notes")
	case *bill.Payment:
		e.Kind = "payment"
		e.Country = country(doc.Regime)
		e.Addons = addons(doc.Addons)
		e.Currency = doc.Currency.String()
		e.Supplier = partyName(doc.Supplier)
		e.Lines = len(doc.Lines)
		add(hasLogo(doc.Supplier), "logo")
		add(len(doc.Notes) > 0, "notes")
		add(len(doc.Methods) > 0, "payment")
	case *org.Party:
		e.Kind = "party"
		if doc.TaxID != nil {
			e.Country = strings.ToLower(doc.TaxID.Country.String())
		}
		e.Supplier = doc.Name
		add(hasLogo(doc), "logo")
	case *note.Message:
		e.Kind = "message"
	default:
		e.Kind = "unknown"
	}

	add(strings.Contains(name, "multipage"), "multipage")
	add(e.Void, "void")
	add(e.Sandbox, "sandbox")
	add(e.AdjustmentMode, "adjustment")

	if render != nil {
		out, err := render(ctx, env)
		if err != nil {
			return Entry{}, fmt.Errorf("rendering: %w", err)
		}
		add(qrMarker.Match(out), "qr")
	}

	slices.Sort(tags)
	e.Tags = tags
	return e, nil
}

func kindForInvoice(t cbc.Key) string {
	switch t {
	case bill.InvoiceTypeCreditNote, bill.InvoiceTypeDebitNote, bill.InvoiceTypeProforma, bill.InvoiceTypeCorrective:
		return t.String()
	default:
		return "invoice"
	}
}

func country(r tax.Regime) string {
	return strings.ToLower(r.Country.String())
}

func addons(a tax.Addons) []string {
	keys := a.GetAddons()
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k.String())
	}
	return out
}

func partyName(p *org.Party) string {
	if p == nil {
		return ""
	}
	return p.Name
}

func hasLogo(p *org.Party) bool {
	return p != nil && len(p.Logos) > 0
}

// title turns "es-verifactu-invoice-tax-included.adjustment.es.json" into
// "verifactu invoice tax included", dropping the country prefix and the
// state suffixes, which are shown separately.
func title(name string) string {
	base := strings.TrimSuffix(name, ".json")
	if i := strings.IndexByte(base, '.'); i > 0 {
		base = base[:i]
	}
	parts := strings.Split(base, "-")
	if len(parts) > 1 && len(parts[0]) == 2 {
		parts = parts[1:]
	}
	return strings.Join(parts, " ")
}
