// Package previewer holds the page the development server serves at /: a
// filterable catalog of the examples rendered in the browser through the
// WebAssembly build, built with PopUI and driven by Alpine.
package previewer

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/invopop/gobl.html/display"
	"github.com/invopop/gobl.html/internal/catalog"
	"github.com/invopop/gobl.html/locales"
	"github.com/invopop/popui.go/props"
)

// Script is the Alpine component behind the page, served at /previewer.js.
//
//go:embed previewer.js
var Script []byte

// Group is one facet group in the sidebar.
type Group struct {
	Key    string
	Title  string
	All    string // label of the entry that clears the choice
	Facets []Facet
}

// Facet is one selectable value within a group.
type Facet struct {
	Value string
	Label string
}

var kindLabels = map[string]string{
	"invoice":     "Invoice",
	"credit-note": "Credit note",
	"debit-note":  "Debit note",
	"proforma":    "Proforma",
	"corrective":  "Corrective",
	"order":       "Order",
	"delivery":    "Delivery",
	"payment":     "Payment",
	"party":       "Party",
	"message":     "Message",
}

// FlagURL returns the CDN flag image for a lower-case tax country code.
// Tax codes that differ from ISO 3166-1 (Greece uses EL) are mapped to the
// ISO code the flag set is keyed by.
func FlagURL(country string) string {
	if country == "el" {
		country = "gr"
	}
	return "https://assets.invopop.com/flags/" + country + ".svg"
}

// KindLabel returns the readable name of a document kind.
func KindLabel(kind string) string {
	if l, ok := kindLabels[kind]; ok {
		return l
	}
	return kind
}

// KindStatus maps a document kind to a TagStatus color.
func KindStatus(kind string) string {
	switch kind {
	case "invoice":
		return "blue"
	case "credit-note", "debit-note", "corrective":
		return "orange"
	case "order", "proforma":
		return "purple"
	case "delivery":
		return "teal"
	case "payment":
		return "green"
	default:
		return "grey"
	}
}

// Groups derives the facet groups from the catalog entries.
func Groups(entries []catalog.Entry) []Group {
	collect := func(key, title, all string, of func(catalog.Entry) []string, label func(string) string) Group {
		seen := map[string]bool{}
		g := Group{Key: key, Title: title, All: all}
		for _, e := range entries {
			for _, v := range of(e) {
				if v == "" || seen[v] {
					continue
				}
				seen[v] = true
				g.Facets = append(g.Facets, Facet{Value: v, Label: label(v)})
			}
		}
		sort.Slice(g.Facets, func(i, j int) bool { return g.Facets[i].Label < g.Facets[j].Label })
		return g
	}
	return []Group{
		collect("country", "Country", "All countries", func(e catalog.Entry) []string { return []string{e.Country} }, strings.ToUpper),
		collect("kind", "Document", "All documents", func(e catalog.Entry) []string { return []string{e.Kind} }, KindLabel),
		collect("addon", "Addon", "All addons", func(e catalog.Entry) []string { return e.Addons }, func(s string) string { return s }),
		collect("tag", "Content", "All content", func(e catalog.Entry) []string { return e.Tags }, func(s string) string { return s }),
	}
}

// localeNames gives each shipped locale its own name for the selector.
var localeNames = map[string]string{
	"ar": "العربية",
	"ca": "Català",
	"da": "Dansk",
	"de": "Deutsch",
	"el": "Ελληνικά",
	"en": "English",
	"es": "Español",
	"eu": "Euskara",
	"fi": "Suomi",
	"fr": "Français",
	"gl": "Galego",
	"it": "Italiano",
	"nl": "Nederlands",
	"no": "Norsk",
	"pl": "Polski",
	"pt": "Português",
	"ro": "Română",
}

// SnippetKeyOptions lists the anchors a snippet can be placed at, in
// document order.
func SnippetKeyOptions() []props.SelectOption {
	keys := display.Keys()
	opts := make([]props.SelectOption, 0, len(keys))
	for _, k := range keys {
		opts = append(opts, props.SelectOption{Value: k.String(), Label: k.String()})
	}
	return opts
}

// SnippetModeOptions lists the placements of a snippet relative to its
// anchor.
func SnippetModeOptions() []props.SelectOption {
	return []props.SelectOption{
		{Value: string(display.ModeBefore), Label: "Before"},
		{Value: string(display.ModeAfter), Label: "After"},
		{Value: string(display.ModeReplace), Label: "Replace"},
	}
}

// LocaleOptions lists the locales embedded in the renderer for the header
// selector, after an entry that keeps each example's own locale.
func LocaleOptions() []props.SelectOption {
	codes := locales.Codes()
	opts := make([]props.SelectOption, 0, len(codes)+1)
	opts = append(opts, props.SelectOption{Value: "", Label: "Locale: default"})
	for _, c := range codes {
		name := localeNames[c]
		if name == "" {
			name = strings.ToUpper(c)
		}
		opts = append(opts, props.SelectOption{Value: c, Label: name + " (" + c + ")"})
	}
	return opts
}
