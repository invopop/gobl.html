// Package goblhtml provides a simple way to render HTML documents from GOBL envelopes.
package goblhtml

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl"
	"github.com/invopop/gobl.html/components"
	"github.com/invopop/gobl.html/internal"
	srclocales "github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
)

const (
	defaultLocale i18n.Code = "en"
)

// locales contains the database of locales defined for this package.
var locales *i18n.Locales

func init() {
	// Locales are loaded into a global variable for this package so that
	// users can continue to use their own locales outside of this one.
	locales = new(i18n.Locales)
	if err := locales.LoadWithDefault(srclocales.Content, defaultLocale); err != nil {
		panic(fmt.Errorf("loading locales: %w", err))
	}
}

// Option defines a configuration option to use for rendering.
type Option func(*internal.Opts)

// WithLogo overrides whatever logo was defined in the original envelope,
// if at all, using the provided logo according to the document type.
func WithLogo(logo *org.Image) Option {
	return func(o *internal.Opts) {
		o.Logo = logo
	}
}

// WithNotes adds the provided string to the envelope notes.
func WithNotes(txt string) Option {
	return func(o *internal.Opts) {
		o.Notes = txt
	}
}

// WithLocale sets the locale to use for rendering.
func WithLocale(locale i18n.Code) Option {
	return func(o *internal.Opts) {
		o.Locale = locale
	}
}

// WithCalFormatter prepares simple date and datetime formatting.
func WithCalFormatter(date, dateTime string, loc *time.Location) Option {
	return func(o *internal.Opts) {
		cf := internal.CalFormatterISO
		if date != "" {
			cf.Date = date
		}
		if dateTime != "" {
			cf.DateTime = dateTime
		}
		if loc != nil {
			cf.Location = loc
		}
		o.CalFormatter = &cf
	}
}

// WithNumFormatter defines a customer number formatter to use instead of
// that provided by default for the currency.
func WithNumFormatter(nf num.Formatter) Option {
	return func(o *internal.Opts) {
		o.NumFormatter = &nf
	}
}

// WithEmbeddedStylesheets indicates that the stylesheets should be embedded
// inside the HTML document as opposed to links.
func WithEmbeddedStylesheets() Option {
	return func(o *internal.Opts) {
		o.EmbedStylesheets = true
	}
}

// WithZugferd adds the Zugferd XMP metadata to the HTML document.
func WithZugferd() Option {
	return func(o *internal.Opts) {
		o.Zugferd = true
	}
}

// Render takes the GOBL envelope and attempts to render an HTML document
// from it.
func Render(ctx context.Context, env *gobl.Envelope, opts ...Option) ([]byte, error) {
	o := new(internal.Opts)
	for _, opt := range opts {
		opt(o)
	}

	// Prepare the Locale
	if o.Locale == "" {
		o.Locale = defaultLocale
	}
	l := locales.Match(string(o.Locale))
	if l == nil {
		l = locales.Get(defaultLocale)
	}
	ctx = l.WithContext(ctx)

	// Extract the currency to use for formatting
	if o.NumFormatter == nil {
		cur := currency.EUR
		switch doc := env.Extract().(type) {
		case *bill.Invoice:
			cur = doc.Currency
		}
		nf := cur.Def().Formatter()
		o.NumFormatter = &nf
	}

	if o.CalFormatter == nil {
		o.CalFormatter = &internal.CalFormatterISO
	}

	ctx = internal.WithOptions(ctx, o)

	out := components.Envelope(env)
	buf := new(bytes.Buffer)
	if err := out.Render(ctx, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
