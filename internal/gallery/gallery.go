// Package gallery builds examples/index.html: scan examples/*.json, locale hints, embed catalog.html.
package gallery

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	_ "embed"
)

//go:embed catalog.html
var catalogHTML string

var catalogTmpl = template.Must(template.New("catalog").Parse(catalogHTML))

// Entry is one tile in the examples gallery (HTML + PDF links for the dev server).
type Entry struct {
	Label    string
	HtmlHref string
	PdfHref  string
	// FlagURL is https://assets.invopop.com/flags/{cc}.svg when the example basename has a country prefix.
	FlagURL string
}

// ExamplesDir is the path to the examples directory relative to the module root.
const ExamplesDir = "examples"

// IndexHTMLPath returns the path to the generated gallery page under ExamplesDir.
func IndexHTMLPath() string { return filepath.Join(ExamplesDir, "index.html") }

// LocaleForExample maps an example basename prefix to an ISO 639-1 language code for ?locale=.
func LocaleForExample(base string) string {
	i := strings.IndexByte(base, '-')
	if i <= 0 {
		return "en"
	}
	switch strings.ToLower(base[:i]) {
	case "mx", "es", "ar":
		return "es"
	case "pt":
		return "pt"
	case "de":
		return "de"
	case "pl":
		return "pl"
	case "it":
		return "it"
	case "gr":
		return "el"
	case "sa":
		return "ar"
	case "us", "zw":
		return "en"
	default:
		return "en"
	}
}

// ListEntries scans dir for *.json (non-recursive), sorted by name.
func ListEntries(dir string) ([]Entry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading examples directory: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".json") {
			names = append(names, e.Name())
		}
	}
	slices.Sort(names)

	out := make([]Entry, 0, len(names))
	for _, name := range names {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		q := "?locale=" + LocaleForExample(base)
		cc := countryCodeFromBase(base)
		out = append(out, Entry{
			Label:    base,
			HtmlHref: "/" + base + ".html" + q,
			PdfHref:  "/" + base + ".pdf" + q + "#view=Fit&toolbar=0",
			FlagURL:  flagAssetURL(cc),
		})
	}
	return out, nil
}

func countryCodeFromBase(base string) string {
	i := strings.IndexByte(base, '-')
	if i <= 0 {
		return ""
	}
	return strings.ToLower(base[:i])
}

func flagAssetURL(cc string) string {
	if cc == "" {
		return ""
	}
	return "https://assets.invopop.com/flags/" + cc + ".svg"
}

// WriteIndexHTML renders the gallery to outPath from examplesDir/*.json.
func WriteIndexHTML(examplesDir, outPath string) error {
	items, err := ListEntries(examplesDir)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := catalogTmpl.Execute(&buf, items); err != nil {
		return fmt.Errorf("rendering gallery: %w", err)
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}
	return nil
}
