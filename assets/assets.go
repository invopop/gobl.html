// Package assets contains the static resources for things like styles.
package assets

import (
	"embed"
	"encoding/base64"
	"io/fs"
	"path"
	"regexp"
)

//go:embed styles
//go:embed scripts
//go:embed fonts

// Content stores all the asset contents.
var Content embed.FS

// FontStylesheet is the stylesheet declaring the document typefaces, relative
// to the fonts directory.
const FontStylesheet = "fonts/fonts.css"

// fontURL matches the url() of a font file in the font stylesheet, which is
// relative to the stylesheet itself.
var fontURL = regexp.MustCompile(`url\(([^)]+\.woff2)\)`)

// ReadData reads the data from the given asset path.
func ReadData(path string) string {
	data, err := fs.ReadFile(Content, path)
	if err != nil {
		return ""
	}
	return string(data)
}

// FontStyles returns the font stylesheet with every font file inlined as a
// data URI, so a document carrying it depends on no external resource.
func FontStyles() string {
	css := ReadData(FontStylesheet)
	dir := path.Dir(FontStylesheet)
	return fontURL.ReplaceAllStringFunc(css, func(m string) string {
		file := fontURL.FindStringSubmatch(m)[1]
		data, err := fs.ReadFile(Content, path.Join(dir, file))
		if err != nil {
			return m
		}
		return "url(data:font/woff2;base64," + base64.StdEncoding.EncodeToString(data) + ")"
	})
}
