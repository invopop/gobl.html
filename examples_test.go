package goblhtml_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
	"github.com/yosssi/gohtml"
)

const (
	examplesPath    = "./examples"
	examplesOutPath = "./examples/out"
)

var updateOut = flag.Bool("update", false, "Update the HTML files in the examples/out directory")

func TestGOBLRenderExamples(t *testing.T) {
	examples, err := findExamples()
	require.NoError(t, err)

	for _, example := range examples {
		name := fmt.Sprintf("should convert %s example file successfully", example)

		t.Run(name, func(t *testing.T) {
			data, err := convertExample(example)
			require.NoError(t, err)

			outPath := filepath.Join(examplesOutPath, strings.TrimSuffix(example, ".json")+".html")

			// Make the output pretty!
			out := gohtml.Format(string(data))

			if *updateOut {
				err = os.WriteFile(outPath, []byte(out), 0644)
				require.NoError(t, err)
				return
			}

			expected, err := os.ReadFile(outPath)

			require.False(t, os.IsNotExist(err), "output file %s missing, run tests with `--update` flag to create", filepath.Base(outPath))
			require.NoError(t, err)
			if out != string(expected) {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(expected)),
					B:        difflib.SplitLines(out),
					FromFile: "Expected",
					ToFile:   "Got",
					Context:  2,
				}
				text, _ := difflib.GetUnifiedDiffString(diff)
				t.Errorf("output file %s does not match:\n%s", filepath.Base(outPath), text)
			}
			// assert.Equal(t, string(expected), out, "output file %s does not match, run tests with `--update` flag to update", filepath.Base(outPath))
		})
	}
}

func findExamples() ([]string, error) {
	examples, err := filepath.Glob(filepath.Join(examplesPath, "*.json"))
	if err != nil {
		return nil, err
	}

	for i, example := range examples {
		examples[i] = filepath.Base(example)
	}

	return examples, nil
}

func convertExample(name string) ([]byte, error) {
	env, err := loadExample(name)
	if err != nil {
		return nil, err
	}

	// Extract the state from the name
	var opts []goblhtml.Option
	if strings.HasSuffix(name, ".void.json") {
		opts = append(opts, goblhtml.WithVoid(true))
	}

	ctx := context.Background()
	return goblhtml.Render(ctx, env, opts...)
}

func loadExample(name string) (*gobl.Envelope, error) {
	src, _ := os.Open(filepath.Join(examplesPath, name))

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, err
	}

	env := new(gobl.Envelope)
	if err := json.Unmarshal(buf.Bytes(), env); err != nil {
		return nil, err
	}

	if err := env.Calculate(); err != nil {
		return nil, err
	}

	if err := env.Validate(); err != nil {
		return nil, err
	}

	if *updateOut {
		data, err := json.MarshalIndent(env, "", "\t")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(examplesPath, name), data, 0644); err != nil {
			return nil, err
		}
	}

	return env, nil
}
