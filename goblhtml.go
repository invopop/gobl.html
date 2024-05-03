package goblhtml

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components"
	"github.com/invopop/gobl.html/components/t"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
)

const (
	defaultLanguage i18n.Code = "en"
)

// Option defines a configuration option to use for rendering.
type Option func(*options)

type options struct {
	locale       i18n.Code
	calFormatter *t.CalFormatter
	numFormatter *num.Formatter
	logo         *org.Image
	notes        string
}

// WithLogo overrides whatever logo was defined in the original envelope,
// if at all, using the provided logo according to the document type.
func WithLogo(logo *org.Image) Option {
	return func(o *options) {
		o.logo = logo
	}
}

// WithNotes adds the provided string to the envelope notes.
func WithNotes(txt string) Option {
	return func(o *options) {
		o.notes = txt
	}
}

// WithLocale sets the locale to use for rendering.
func WithLocale(locale i18n.Code) Option {
	return func(o *options) {
		o.locale = locale
	}
}

// WithCalFormatter prepares simple date and datetime formatting.
func WithCalFormatter(date, dateTime string, loc *time.Location) Option {
	return func(o *options) {
		cf := t.CalFormatterISO
		if date != "" {
			cf.Date = date
		}
		if dateTime != "" {
			cf.DateTime = dateTime
		}
		if loc != nil {
			cf.Location = loc
		}
		o.calFormatter = &cf
	}
}

// WithNumFormatter defines a customer number formatter to use instead of
// that provided by default for the currency.
func WithNumFormatter(nf num.Formatter) Option {
	return func(o *options) {
		o.numFormatter = &nf
	}
}

// Render takes the GOBL envelope and attempts to render an HTML document
// from it.
func Render(ctx context.Context, env *gobl.Envelope, opts ...Option) ([]byte, error) {
	conf := new(options)
	for _, opt := range opts {
		opt(conf)
	}

	// Prepare the Locale
	if conf.locale == "" {
		conf.locale = defaultLanguage
	}
	ctx, err := ctxi18n.WithLocale(ctx, string(conf.locale))
	if err != nil {
		return nil, fmt.Errorf("preparing locale: %w", err)
	}

	// Is there a calendar formatter?
	if conf.calFormatter != nil {
		ctx = t.WithCalFormatter(ctx, *conf.calFormatter)
	}

	// Extract the currency to use for formatting
	if conf.numFormatter == nil {
		cur := currency.EUR
		switch doc := env.Extract().(type) {
		case *bill.Invoice:
			cur = doc.Currency
		}
		nf := cur.Def().Formatter()
		conf.numFormatter = &nf
	}
	ctx = t.WithNumFormatter(ctx, *conf.numFormatter)

	// Does a logo need to be replaced?
	if conf.logo != nil {
		switch doc := env.Extract().(type) {
		case *bill.Invoice:
			doc.Supplier.Logos = []*org.Image{conf.logo}
		}
	}

	// Any additional notes?
	if conf.notes != "" {
		env.Head.Notes = conf.notes
	}

	out := components.Envelope(env)
	buf := new(bytes.Buffer)
	if err := out.Render(ctx, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
