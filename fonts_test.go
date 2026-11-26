package goblhtml_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// externalURL matches an attribute or CSS url() pointing at a remote host,
// which an embedded document must not contain anywhere in its head.
var externalURL = regexp.MustCompile(`(?i)(href|src)="https?://|url\(\s*["']?https?://`)

func TestFontsEmbedded(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(examplesPath, "es-verifactu-invoice-env.json"))
	require.NoError(t, err)
	env := new(gobl.Envelope)
	require.NoError(t, json.Unmarshal(data, env))

	t.Run("embedded assets carry the fonts inline", func(t *testing.T) {
		out, err := goblhtml.Render(context.Background(), env, goblhtml.WithEmbeddedAssets())
		require.NoError(t, err)
		html := string(out)
		head := html[:strings.Index(html, "<body")]

		assert.Contains(t, head, `@font-face`)
		assert.Contains(t, head, `font-family: "Inter"`)
		assert.Contains(t, head, `font-family: "Geist Mono"`)
		assert.Equal(t, 2, strings.Count(head, "url(data:font/woff2;base64,"), "both fonts inlined")
		assert.NotContains(t, head, "fonts.googleapis.com")
		assert.NotContains(t, head, "<link")
		assert.Empty(t, externalURL.FindString(head), "no external resource in the head")
	})

	t.Run("linked assets reference the font stylesheet", func(t *testing.T) {
		out, err := goblhtml.Render(context.Background(), env)
		require.NoError(t, err)
		html := string(out)

		assert.Contains(t, html, `<link rel="stylesheet" href="fonts/fonts.css">`)
		assert.NotContains(t, html, "fonts.googleapis.com")
		assert.NotContains(t, html, "data:font/woff2")
	})
}
