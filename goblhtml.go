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
	"github.com/invopop/gobl.html/internal/doc"
	"github.com/invopop/gobl.html/layout"
	srclocales "github.com/invopop/gobl.html/locales"
	"github.com/invopop/gobl/l10n"
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

// WithEmbeddedAssets indicates that the stylesheets and scripts should be
// embedded inside the HTML document as opposed to links.
func WithEmbeddedAssets() Option {
	return func(o *internal.Opts) {
		o.EmbedAssets = true
	}
}

// WithLayout indicates that the document should be rendered
// acording to the given Layout.
func WithLayout(l layout.Code) Option {
	return func(o *internal.Opts) {
		o.Layout = l
	}
}

// WithVoid indicates that the document is marked as void
// and should be rendered accordingly.
func WithVoid(void bool) Option {
	return func(o *internal.Opts) {
		o.Void = void
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
		if doc := doc.ExtractFrom(env); doc != nil {
			nf := doc.GetCurrency().Def().Formatter()

			if doc.GetRegime().Country == l10n.PT.Tax() {
				// As required by the Portuguese tax law
				nf.ThousandsSeparator = " "
				nf.DecimalMark = ","
				nf.Template = "%n %u"
			}

			o.NumFormatter = &nf
		}
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
