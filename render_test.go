package goblhtml_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadRawExample reads an example without calculating or validating it, the way
// an envelope arriving over the wire is handled.
func loadRawExample(t *testing.T, name string) *gobl.Envelope {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(examplesPath, name))
	require.NoError(t, err)

	env := new(gobl.Envelope)
	require.NoError(t, json.Unmarshal(data, env))

	return env
}

// stripDocField removes a top-level field from the envelope's document, so we
// can simulate a document that was never calculated.
func stripDocField(t *testing.T, name, field string) *gobl.Envelope {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(examplesPath, name))
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	doc, ok := raw["doc"].(map[string]any)
	require.True(t, ok, "example has no doc")
	delete(doc, field)

	out, err := json.Marshal(raw)
	require.NoError(t, err)

	env := new(gobl.Envelope)
	require.NoError(t, json.Unmarshal(out, env))

	return env
}

func TestRenderWithoutTotals(t *testing.T) {
	ctx := context.Background()

	t.Run("invoice is rejected instead of panicking", func(t *testing.T) {
		env := stripDocField(t, "de-invoice.json", "totals")

		var out []byte
		var err error
		require.NotPanics(t, func() {
			out, err = goblhtml.Render(ctx, env)
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invoice totals are missing")
		assert.Nil(t, out)
	})

	t.Run("order renders without a totals section", func(t *testing.T) {
		env := stripDocField(t, "pt-at-order.json", "totals")

		var out []byte
		var err error
		require.NotPanics(t, func() {
			out, err = goblhtml.Render(ctx, env)
		})

		require.NoError(t, err)
		assert.NotContains(t, string(out), `<tr class="sum strong">`)
	})

	t.Run("non-valued delivery renders without a totals section", func(t *testing.T) {
		env := loadRawExample(t, "pt-at-delivery-non-valued.json")

		out, err := goblhtml.Render(ctx, env)
		require.NoError(t, err)
		assert.NotContains(t, string(out), `<tr class="sum strong">`)
	})
}
