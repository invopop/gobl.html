// Package internal is used for internal option configuration.
package internal

import (
	"context"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/internal/layout"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
)

type optionsKey string

const (
	optsKey optionsKey = "opts"
)

// Opts defines configuration options used internally with
// the current document. Putting this amount of information inside a
// context is normally an anti-pattern in Go, but saves a huge amount
// of effort when writing components.
type Opts struct {
	// Locale is the language code to use for rendering.
	Locale i18n.Code
	// Logo used instead of the document's logo.
	Logo *org.Image
	// Notes to add to the document footer instead of the notes included
	// in the envelope.
	Notes string
	// NumFormatter is used to format numbers.
	NumFormatter *num.Formatter
	// CalFormatter is used to format calendar dates and times.
	CalFormatter *CalFormatter
	// EmbedStylesheets when try ensures that all the stylesheet files
	// are contained inside the HTML output. This is useful for PDF
	// output or to avoid additional requests.
	EmbedStylesheets bool
	// Layout indicates the Layout used in te document
	Layout layout.Code
}

// WithOptions prepares the context with the options to use.
func WithOptions(ctx context.Context, opts *Opts) context.Context {
	return context.WithValue(ctx, optsKey, opts)
}

// Options provides the current options in the context.
func Options(ctx context.Context) *Opts {
	if opts, ok := ctx.Value(optsKey).(*Opts); ok {
		return opts
	}
	return nil
}

// CalFormatter defines a simple date and datetime formatter
type CalFormatter struct {
	Date     string // Golang date format for dates, e.g. `02/01/2006`
	Location *time.Location
	DateTime string // Date-time format
}

// CalFormatterISO is the default formatter for dates and times based on
// the recommended ISO 8601 formatting.
var CalFormatterISO = CalFormatter{
	Date:     "2006-01-02",
	DateTime: "2006-01-02 15:04",
	Location: time.UTC,
}
