package goblhtml

import (
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/gobl.html/layout"
	"github.com/invopop/gobl/org"
)

// Params is the serializable form of the rendering options, so that the same
// configuration can arrive as JSON (the WASM renderer), as query parameters
// (the development server) or as a task configuration (pdf.go). Field names
// follow pdf.go's task configuration.
//
// Date and time formats accept either Go reference layouts or the subset of
// strftime directives used by Invopop's configuration forms (%Y, %m, %d, %H,
// %I, %M, %S, %p, %P).
type Params struct {
	// Locale is the language code to render the document in.
	Locale i18n.Code `json:"locale,omitempty" query:"locale"`
	// LogoURL replaces the supplier's logo.
	LogoURL string `json:"logo_url,omitempty" query:"logo_url"`
	// LogoHeight sets the height of the replacement logo, in pixels.
	LogoHeight int32 `json:"logo_height,omitempty" query:"logo_height"`
	// DateFormat overrides the ISO date format.
	DateFormat string `json:"date_format,omitempty" query:"date_format"`
	// TimeFormat overrides the ISO time format.
	TimeFormat string `json:"time_format,omitempty" query:"time_format"`
	// Notes are added to the document footer.
	Notes string `json:"notes,omitempty" query:"notes"`
	// Layout selects the page size; invalid or empty values use A4.
	Layout layout.Code `json:"layout,omitempty" query:"layout"`
	// Label is shown next to the document title.
	Label string `json:"label,omitempty" query:"label"`
	// Watermark text is overlaid on every page.
	Watermark string `json:"watermark,omitempty" query:"watermark"`
	// AdjustmentMode renders credit and debit notes as adjustments to the
	// original invoice.
	AdjustmentMode bool `json:"adjustment_mode,omitempty" query:"adjustment_mode"`
	// CurrencyTemplate overrides the position of the currency unit.
	CurrencyTemplate string `json:"currency_template,omitempty" query:"currency_template"`
	// NegativeTemplate overrides how negative amounts are shown.
	NegativeTemplate string `json:"negative_template,omitempty" query:"negative_template"`
	// ThousandsSeparator overrides the currency's thousands separator.
	ThousandsSeparator string `json:"thousands_separator,omitempty" query:"thousands_separator"`
	// DecimalMark overrides the currency's decimal mark.
	DecimalMark string `json:"decimal_mark,omitempty" query:"decimal_mark"`
	// Void marks the document as void.
	Void bool `json:"void,omitempty" query:"void"`
	// Sandbox marks the document as a sandbox test.
	Sandbox bool `json:"sandbox,omitempty" query:"sandbox"`
	// EmbedAssets inlines the stylesheets and scripts instead of linking them,
	// so the HTML is self-contained.
	EmbedAssets bool `json:"embed_assets,omitempty" query:"embed_assets"`
}

// dateReplacer maps strftime-style date directives to Go's reference layout.
var dateReplacer = strings.NewReplacer(
	"%Y", "2006",
	"%m", "01",
	"%d", "02",
)

// timeReplacer maps strftime-style time directives to Go's reference layout.
// Longer and uppercase tokens come first so that none is consumed by a
// shorter prefix.
var timeReplacer = strings.NewReplacer(
	"%H", "15",
	"%I", "03",
	"%M", "04",
	"%S", "05",
	"%p", "PM",
	"%P", "pm",
)

// Options converts the parameters into the functional options accepted by
// Render. Zero values leave the defaults untouched.
func (p *Params) Options() []Option {
	if p == nil {
		return nil
	}
	opts := make([]Option, 0, 16)
	if p.Locale != "" {
		opts = append(opts, WithLocale(p.Locale))
	}
	if p.LogoURL != "" {
		opts = append(opts, WithLogo(&org.Image{
			URL:    p.LogoURL,
			Height: p.LogoHeight,
		}))
	}
	if p.DateFormat != "" || p.TimeFormat != "" {
		opts = append(opts, WithCalFormatter(
			dateReplacer.Replace(p.DateFormat),
			timeReplacer.Replace(p.TimeFormat),
			nil,
		))
	}
	if p.Notes != "" {
		opts = append(opts, WithNotes(p.Notes))
	}
	if p.Layout.IsValid() {
		opts = append(opts, WithLayout(p.Layout))
	} else {
		opts = append(opts, WithLayout(layout.A4))
	}
	if p.Label != "" {
		opts = append(opts, WithLabel(p.Label))
	}
	if p.Watermark != "" {
		opts = append(opts, WithWatermark(p.Watermark))
	}
	if p.AdjustmentMode {
		opts = append(opts, EnableAdjustmentMode)
	}
	if p.CurrencyTemplate != "" {
		opts = append(opts, WithCurrencyTemplate(p.CurrencyTemplate))
	}
	if p.NegativeTemplate != "" {
		opts = append(opts, WithNegativeTemplate(p.NegativeTemplate))
	}
	if p.ThousandsSeparator != "" {
		opts = append(opts, WithThousandsSeparator(p.ThousandsSeparator))
	}
	if p.DecimalMark != "" {
		opts = append(opts, WithDecimalMark(p.DecimalMark))
	}
	if p.Void {
		opts = append(opts, WithVoid(true))
	}
	if p.Sandbox {
		opts = append(opts, WithSandbox(true))
	}
	if p.EmbedAssets {
		opts = append(opts, WithEmbeddedAssets())
	}
	return opts
}
