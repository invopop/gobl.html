// Package assets contains the static resources for things like styles.
package assets

import "embed"

//go:embed styles

// Content stores all the asset contents.
var Content embed.FS
