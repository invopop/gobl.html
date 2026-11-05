//go:build ignore

// import-docs-examples copies the GOBL examples published in the Invopop docs
// into examples/, so gobl.html renders every document the docs show and can
// keep more of its own on top. It reads the built snippets (the ```json block
// of each snippets/{invoices,payments,deliveries}/<cc>/<name>.mdx, skipping
// .min.mdx inputs and accordions), envelops, calculates and validates them
// with the gobl version pinned here, and writes examples/<cc>-<name>.json.
// A file whose document digest is unchanged is left alone, so re-running is
// a no-op and envelope UUIDs stay put. examples/docs-examples.txt lists what
// came from the docs; entries the docs no longer have are removed. A docs
// example whose name an example of gobl.html's own already uses is skipped,
// and one the current gobl rules reject is reported: both are for the docs
// to resolve.
//
//	go run scripts/import-docs-examples.go ../docs.invopop.com
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/invopop/gobl"
	_ "github.com/invopop/gobl.dev/bundle"
	_ "github.com/invopop/gobl.ro.anaf/addon"
)

const (
	examplesDir = "examples"
	manifest    = "examples/docs-examples.txt"
)

var trees = []string{"invoices", "payments", "deliveries"}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/import-docs-examples.go <docs repo dir>")
		os.Exit(2)
	}
	docs := os.Args[1]

	previous := readManifest()
	current := map[string]string{} // example file -> snippet path
	failed := 0
	for _, tree := range trees {
		files, _ := filepath.Glob(filepath.Join(docs, "snippets", tree, "*", "*.mdx"))
		sort.Strings(files)
		for _, f := range files {
			base := strings.TrimSuffix(filepath.Base(f), ".mdx")
			if strings.HasSuffix(base, ".min") || strings.Contains(base, "accordion") {
				continue
			}
			cc := filepath.Base(filepath.Dir(f))
			name := strings.ToLower(base)
			if !strings.HasPrefix(name, cc+"-") {
				name = cc + "-" + name
			}
			out := name + ".json"
			rel, _ := filepath.Rel(docs, f)
			if _, mine := previous[out]; !mine {
				if _, err := os.Stat(filepath.Join(examplesDir, out)); err == nil {
					fmt.Printf("≠ %s: %s is gobl.html's own, kept\n", rel, out)
					continue
				}
			}
			if err := importOne(f, filepath.Join(examplesDir, out)); err != nil {
				fmt.Printf("✗ %s: %v\n", rel, err)
				failed++
				continue
			}
			current[out] = rel
		}
	}
	for file := range previous {
		if _, ok := current[file]; !ok {
			os.Remove(filepath.Join(examplesDir, file))
			os.Remove(filepath.Join(examplesDir, "out", strings.TrimSuffix(file, ".json")+".html"))
			fmt.Printf("- %s (no longer in the docs)\n", file)
		}
	}
	writeManifest(current)
	fmt.Printf("%d docs examples, %d failed\n", len(current), failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func importOne(src, dst string) error {
	data, err := extractJSON(src)
	if err != nil {
		return err
	}
	old, _ := os.ReadFile(dst)
	if data, err = keepUUID(data, old); err != nil {
		return err
	}
	doc, err := gobl.Parse(data)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}
	env, err := gobl.Envelop(doc)
	if err != nil {
		return fmt.Errorf("enveloping: %w", err)
	}
	if err := env.Validate(); err != nil {
		return fmt.Errorf("validating: %w", err)
	}
	if old != nil {
		prev := new(gobl.Envelope)
		if json.Unmarshal(old, prev) == nil && prev.Head != nil && prev.Head.Digest != nil &&
			env.Head.Digest != nil && prev.Head.Digest.Value == env.Head.Digest.Value {
			return nil // unchanged: keep the file and its UUIDs
		}
	}
	out, err := json.MarshalIndent(env, "", "\t")
	if err != nil {
		return err
	}
	fmt.Printf("✓ %s\n", dst)
	return os.WriteFile(dst, append(out, '\n'), 0o644)
}

// keepUUID gives a document without a uuid the one its previous import has,
// so the digest only changes when the content does. Most docs snippets are
// built without one.
func keepUUID(data, old []byte) ([]byte, error) {
	if old == nil {
		return data, nil
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}
	if _, ok := doc["uuid"]; ok {
		return data, nil
	}
	var prev struct {
		Doc map[string]json.RawMessage `json:"doc"`
	}
	if json.Unmarshal(old, &prev) != nil || prev.Doc["uuid"] == nil {
		return data, nil
	}
	doc["uuid"] = prev.Doc["uuid"]
	return json.Marshal(doc)
}

// extractJSON returns the body of the first ```json block in an MDX file.
func extractJSON(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var b strings.Builder
	in := false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case !in && strings.HasPrefix(strings.TrimSpace(line), "```json"):
			in = true
		case in && strings.TrimSpace(line) == "```":
			return []byte(b.String()), nil
		case in:
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return nil, fmt.Errorf("no ```json block")
}

func readManifest() map[string]string {
	m := map[string]string{}
	f, err := os.Open(manifest)
	if err != nil {
		return m
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "#") {
			continue
		}
		if parts := strings.SplitN(sc.Text(), "\t", 2); len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func writeManifest(m map[string]string) {
	files := make([]string, 0, len(m))
	for f := range m {
		files = append(files, f)
	}
	sort.Strings(files)
	var b strings.Builder
	b.WriteString("# Examples imported from docs.invopop.com by scripts/import-docs-examples.go: example\tsnippet\n")
	for _, f := range files {
		fmt.Fprintf(&b, "%s\t%s\n", f, m[f])
	}
	os.WriteFile(manifest, []byte(b.String()), 0o644)
}
