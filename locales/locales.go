// Package locales provides all the content for translations.
package locales

import "embed"

//go:embed en

// Content is the embedded content for the locales.
var Content embed.FS
