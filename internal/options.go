// Package internal is used for internal option configuration.
package internal

import (
	"context"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/display"
	"github.com/invopop/gobl.html/layout"
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

	// CurrencyTemplate is used to format currency amounts with their units,
	// and overrides those from the number formatter.
	CurrencyTemplate string
	// ThousandsSeparator overrides the thousands separator from the number formatter.
	ThousandsSeparator string
	// DecimalMark overrides the decimal mark from the number formatter.
	DecimalMark string
	// NegativeTemplate overrides the negative template from the number formatter.
	NegativeTemplate string

	// CalFormatter is used to format calendar dates and times.
	CalFormatter *CalFormatter
	// EmbedAssets when try ensures that all the stylesheet or script files
	// are contained inside the HTML output. This is useful for PDF output
	// or to avoid additional requests.
	EmbedAssets bool
	// Layout indicates the Layout used in te document
	Layout layout.Code
	// Void indicates if the document should be rendered as void
	Void bool
	// Sandbox indicates if the document should be rendered in sandbox mode
	Sandbox bool
	// Watermark indicates the watermark to use in the document
	Watermark string
	// Label indicates the label to use in the document
	Label string
	// AdjustmentMode indicates that invoice types credit-note and debit-note
	// or similar should be rendered as adjustments to the original invoice, as
	// required in certain jurisdictions.
	AdjustmentMode bool
	// Display determines which sections to hide and the custom HTML
	// snippets to inject.
	Display *display.Options
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

// Display provides the display options in the context, which may be nil.
// All display.Options methods are safe to call on a nil value.
func Display(ctx context.Context) *display.Options {
	if opts := Options(ctx); opts != nil {
		return opts.Display
	}
	return nil
}

// Hidden returns true when the section's key or any of its ancestors is
// hidden, or an ancestor has been replaced.
func Hidden(ctx context.Context, key display.Key) bool {
	return Display(ctx).Hidden(key)
}

// Replacement provides the replace snippet defined exactly at the key.
func Replacement(ctx context.Context, key display.Key) (string, bool) {
	return Display(ctx).Replacement(key)
}

// Before provides the snippets to render before the key, in order.
func Before(ctx context.Context, key display.Key) []string {
	return Display(ctx).Before(key)
}

// After provides the snippets to render after the key, in order.
func After(ctx context.Context, key display.Key) []string {
	return Display(ctx).After(key)
}

// CalFormatter defines a simple date and time formatter. Date and time
// formats are configured independently; date-time values are rendered by
// composing the two with a separator.
type CalFormatter struct {
	Date     string // Golang date format for dates, e.g. `02/01/2006`
	Time     string // Golang time format for times, e.g. `15:04`
	Location *time.Location
}

// CalFormatterISO is the default formatter for dates and times based on
// the recommended ISO 8601 formatting.
var CalFormatterISO = CalFormatter{
	Date:     "2006-01-02",
	Time:     "15:04",
	Location: time.UTC,
}

// DateTimeSeparator is used to join the formatted date and time when
// rendering a cal.DateTime value.
const DateTimeSeparator = " · "
