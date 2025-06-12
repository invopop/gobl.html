// Package assets contains the static resources for things like styles.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed styles
//go:embed scripts

// Content stores all the asset contents.
var Content embed.FS

// ReadData reads the data from the given asset path.
func ReadData(path string) string {
	data, err := fs.ReadFile(Content, path)
	if err != nil {
		return ""
	}
	return string(data)
}
