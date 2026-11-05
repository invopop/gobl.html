//go:build ignore

// build-examples turns bare GOBL documents into the enveloped, validated
// examples the tests and the previewer use. Sources come from two places:
//
//   - examples/src/*.json: documents of gobl.html's own, hand-authored to
//     exercise the renderer (the per-regime "full" invoices live here).
//   - the GOBL examples published in the Invopop docs, when a checkout of
//     the docs repository is given: the ```json block of each built
//     snippets/{invoices,payments,deliveries}/<cc>/<name>.mdx (skipping
//     .min.mdx inputs and accordions), written as examples/<cc>-<name>.json
//     and listed in examples/docs-examples.txt so a later run updates and
//     adds them in step with the docs, and removes with -prune. A docs example whose name an
//     example of gobl.html's own already uses is skipped, and one the current
//     gobl rules reject is reported: both are for the docs to resolve.
//
// Each document is enveloped, calculated and validated with the gobl pinned
// here. One whose addon has the tax agency stamp it gets sample stamps
// shaped like the real ones, a code if it has none, and a signature from a
// throwaway key, so the previewer shows what the final PDF carries.
// Re-running is a no-op: a document without a uuid keeps the one its
// previous build had, and an unchanged document keeps its file, UUID and
// signature.
//
//	go run scripts/build-examples.go                      # own sources only
//	go run scripts/build-examples.go -docs ../docs.invopop.com          # add and update
//	go run scripts/build-examples.go -docs ../docs.invopop.com -prune   # also remove what the docs dropped
//	go run scripts/build-examples.go examples/src/xx.json # just these
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/invopop/gobl"
	_ "github.com/invopop/gobl.dev/bundle"
	_ "github.com/invopop/gobl.ro.anaf/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/head"
	"github.com/invopop/gobl/uuid"
)

const (
	examplesDir = "examples"
	manifest    = "examples/docs-examples.txt"
)

var trees = []string{"invoices", "payments", "deliveries"}

const srcDir = "examples/src"

func main() {
	docs := flag.String("docs", "", "checkout of the docs repository to import examples from")
	prune := flag.Bool("prune", false, "with -docs, remove examples the docs no longer have (kept and reported otherwise)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/build-examples.go [-docs DIR] [examples/src/NAME.json ...]")
	}
	flag.Parse()

	failed := 0
	if flag.NArg() > 0 {
		for _, src := range flag.Args() {
			out := filepath.Join(examplesDir, filepath.Base(src))
			if err := importOne(src, out, jsonSource); err != nil {
				fmt.Printf("✗ %s: %v\n", src, err)
				failed++
			}
		}
		exit(failed)
		return
	}

	own, _ := filepath.Glob(filepath.Join(srcDir, "*.json"))
	sort.Strings(own)
	for _, src := range own {
		out := filepath.Join(examplesDir, filepath.Base(src))
		if err := importOne(src, out, jsonSource); err != nil {
			fmt.Printf("✗ %s: %v\n", src, err)
			failed++
		}
	}
	fmt.Printf("%d own sources, %d failed\n", len(own), failed)
	if *docs != "" {
		failed += importDocs(*docs, *prune)
	}
	exit(failed)
}

func exit(failed int) {
	if failed > 0 {
		os.Exit(1)
	}
}

// importDocs brings the docs' examples in and returns how many failed.
func importDocs(docs string, prune bool) int {
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
			if err := importOne(f, filepath.Join(examplesDir, out), extractJSON); err != nil {
				fmt.Printf("✗ %s: %v\n", rel, err)
				failed++
				continue
			}
			current[out] = rel
		}
	}
	// An example the docs no longer have may come from a docs branch not
	// yet merged, so it stays, and in the manifest, unless pruning.
	for file, src := range previous {
		if _, ok := current[file]; ok {
			continue
		}
		if prune {
			os.Remove(filepath.Join(examplesDir, file))
			os.Remove(filepath.Join(examplesDir, "out", strings.TrimSuffix(file, ".json")+".html"))
			fmt.Printf("- %s (no longer in the docs)\n", file)
		} else {
			current[file] = src
			fmt.Printf("? %s: not in these docs, kept (pass -prune to remove)\n", file)
		}
	}
	writeManifest(current)
	fmt.Printf("%d docs examples, %d failed\n", len(current), failed)
	return failed
}

// jsonSource reads a bare document from a .json file.
func jsonSource(path string) ([]byte, error) { return os.ReadFile(path) }

func importOne(src, dst string, read func(string) ([]byte, error)) error {
	data, err := read(src)
	if err != nil {
		return err
	}
	old, _ := os.ReadFile(dst)
	if data, err = keepUUID(data, old); err != nil {
		return err
	}
	data, err = sampleCode(data)
	if err != nil {
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
	prev := new(gobl.Envelope)
	if old != nil && json.Unmarshal(old, prev) == nil && prev.Head != nil && prev.Head.UUID != uuid.Empty {
		env.Head.UUID = prev.Head.UUID // keep the envelope's identity
	}
	env.Head.Stamps = sampleStamps(env)
	if old != nil && prev.Head != nil {
		if sameDigest(prev, env) && sameStamps(prev, env) && (len(prev.Signatures) > 0) == (len(env.Head.Stamps) > 0) {
			return nil // unchanged: keep the file and its signature
		}
	}
	// Stamps are only valid on a signed envelope, as they would be once
	// the agency accepted the document; the key is throwaway.
	if len(env.Head.Stamps) > 0 {
		if err := env.Sign(dsig.NewES256Key()); err != nil {
			return fmt.Errorf("signing: %w", err)
		}
	}
	if err := env.Validate(); err != nil {
		return fmt.Errorf("validating: %w", err)
	}
	out, err := json.MarshalIndent(env, "", "\t")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	fmt.Printf("✓ %s\n", dst)
	return os.WriteFile(dst, out, 0o644)
}

func sameDigest(a, b *gobl.Envelope) bool {
	return a.Head.Digest != nil && b.Head.Digest != nil && a.Head.Digest.Value == b.Head.Digest.Value
}

func sameStamps(a, b *gobl.Envelope) bool {
	x, _ := json.Marshal(a.Head.Stamps)
	y, _ := json.Marshal(b.Head.Stamps)
	return string(x) == string(y)
}

// sampleCode numbers a document that an agency would stamp, since a
// signed envelope needs its document to carry a code. Docs snippets often
// leave it out, as numbering happens on issue.
func sampleCode(data []byte) ([]byte, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}
	if _, ok := doc["code"]; ok {
		return data, nil
	}
	var addons []string
	if raw, ok := doc["$addons"]; ok {
		_ = json.Unmarshal(raw, &addons)
	}
	if !stamped(addons) {
		return data, nil
	}
	doc["code"] = json.RawMessage(`"0001"`)
	return json.Marshal(doc)
}

// stamped says whether an addon set gets sample stamps.
func stamped(addons []string) bool {
	for _, a := range addons {
		switch a {
		case "es-verifactu-v1", "es-tbai-v1", "pt-saft-v1", "mx-cfdi-v4", "pl-favat-v3",
			"sa-zatca-v1", "gr-mydata-v1", "co-dian-v2", "ar-arca-v4", "it-ticket-v1":
			return true
		}
	}
	return false
}

// sampleStamps gives a document the stamps its tax agency would add once
// it is accepted, so the previewer shows the QR codes, codes and chains
// the final PDF carries. The values are samples shaped like the real ones
// (the same ones gobl.html's own fixtures use), with the document's own
// identifiers and totals where the format carries them.
func sampleStamps(env *gobl.Envelope) []*head.Stamp {
	var (
		addons []cbc.Key
		tax    string // supplier tax code
		series string
		code   string
		date   string // YYYY-MM-DD
		total  string
	)
	switch doc := env.Extract().(type) {
	case *bill.Invoice:
		addons = doc.GetAddons()
		if doc.Supplier != nil && doc.Supplier.TaxID != nil {
			tax = doc.Supplier.TaxID.Code.String()
		}
		series, code, date = doc.Series.String(), doc.Code.String(), doc.IssueDate.String()
		if doc.Totals != nil {
			total = doc.Totals.TotalWithTax.String()
		}
	case *bill.Delivery:
		addons = doc.GetAddons()
		if doc.Supplier != nil && doc.Supplier.TaxID != nil {
			tax = doc.Supplier.TaxID.Code.String()
		}
		series, code, date = doc.Series.String(), doc.Code.String(), doc.IssueDate.String()
		if doc.Totals != nil {
			total = doc.Totals.TotalWithTax.String()
		}
	default:
		return nil
	}
	if code == "" {
		code = "001"
	}
	number := code
	if series != "" {
		number = series + "-" + code
	}
	compact := strings.ReplaceAll(date, "-", "") // YYYYMMDD
	var ddmmyyyy, ddmmyy string
	if len(date) == 10 {
		ddmmyyyy = date[8:] + "-" + date[5:7] + "-" + date[:4]
		ddmmyy = date[8:] + date[5:7] + date[2:4]
	}
	stamp := func(prv, val string) *head.Stamp { return &head.Stamp{Provider: cbc.Key(prv), Value: val} }
	var out []*head.Stamp
	for _, a := range addons {
		switch a {
		case "es-verifactu-v1":
			out = append(out, stamp("verifactu-qr", "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?nif="+tax+"&numserie="+number+"&fecha="+ddmmyyyy+"&importe="+total))
		case "es-tbai-v1":
			id := "TBAI-" + tax + "-" + ddmmyy + "-LrNMp23wKPLjx-112"
			out = append(out,
				stamp("tbai-code", id),
				stamp("tbai-qr", "https://batuz.eus/QRTBAI/?id="+id+"&s="+series+"&nf="+code+"&i="+total+"&cr=027"))
		case "pt-saft-v1":
			out = append(out,
				stamp("at-atcud", "AAJFJR2KW4-2"),
				stamp("at-hash", "0vot"),
				stamp("at-hash-full", "0WMEHADswWv5aavY+2VboOclMpPyoQtl7neqfA/7PPvhayQR1OQMok5fFr8N4vi6TGm8ymmHSpfX4QG28jeDJqNuO92v5TbRH78/4lPbcryEmVIgG0x9jXtYPX6RehwYYuUZg1q1XSEYrEGg+zQoPN8VZetI5PxGRWqrkIe8G5w="),
				stamp("at-qr", "A:"+tax+"*B:999999990*C:PT*D:FT*E:N*F:"+compact+"*G:"+series+" "+code+"*H:AAJFJR2KW4-2*I1:PT*I7:"+total+"*N:0.00*O:"+total+"*Q:0vot*R:0"),
				stamp("at-app-id", "0"))
		case "mx-cfdi-v4":
			const sig = "Qk+hftgbpSYMqcc7BoBzGOctcNbDypP4IiHfzgf4mJuraxGEFbo8JEfS6BOzfqXojVLbRxPkXE83fqzWdnowOhOmi0Mvk3CdlTZNZ9i2i7hJv6cwfHXMTPpeOxfptuv2AzhFX7O14gS2szeai5yJIcT+d8/W6H0zN6b/2S12GkMdik0tyF8XRtQQat4Q1jxQzslOp46brROc0Xuobc7/jogq2ib7oaJEu7qDKhjYAkH7BHMaLMJgeC6eVeHUkkrr7RVTkH75H4wteKZNuP85E/F12nCkbMSQq7uplzRuLyO6YEkyIWxpT+0Gqtcb6pziYk8wqwP1pVsVZZ4Q0PPOsw=="
			const satSig = "MQWzFEvDEWag8sM8OWjlCS1DmrZWe+/1VtGsxDz8ycSH1yh8aoUGV3lKezQy190TnB2mncusGvFoUpbDhrk0tlRpTGT2+68/43yb0RZ+yWTVCrd09sCzRk/Dct+oyUKWr2AYYxA0Jt7ECyR4DQfPQYbPF4shsmcerMWxsbMUKqR43shvsb5QTxk/yA6CiWx3z/+uSxvJitUoUihAtFeBBdiNnKVBcG7ifWNJaoaBL03UOm1Ro77rGvUfYuxqqI02Fav98iv9eS//YmGzUTpz0Uvc9Lut+MRqkBvs3G/PBiJvO5Yi6I2PhzNo0O7z9SD6P9WnmufmVI9JFCbHVk+PNg=="
			id := env.Head.UUID.String()
			ts := date + "T07:16:10"
			out = append(out,
				stamp("sat-uuid", id),
				stamp("cfdi-sig", sig),
				stamp("cfdi-serial", "30001000000500003416"),
				stamp("sat-sig", satSig),
				stamp("sat-serial", "30001000000500003456"),
				stamp("sat-timestamp", ts),
				stamp("sat-chain", "||1.1|"+id+"|"+ts+"|SPR190613I52|"+sig+"|30001000000500003456||"),
				stamp("sat-url", "https://verificacfdi.facturaelectronica.sat.gob.mx/default.aspx?id="+id+"&re="+tax+"&tt="+total+"&fe="+sig[len(sig)-8:]))
		case "pl-favat-v3":
			id := tax + "-" + compact + "-D1984449A6BC-02"
			out = append(out,
				stamp("favat-ksef-number", id),
				stamp("favat-qr", "https://ksef-test.mf.gov.pl/web/verify/"+id+"/8KCfB1Sa5Jc4Gg6ctK4HKdwnRwK%2BdxT1ScymF6m5d18%3D"))
		case "sa-zatca-v1":
			out = append(out, stamp("zatca-qr", "AS3Yp9mF2YrZitmF2YjYsyDYp9mE2LnYsdio2YrYqSDZhNmE2YHYudin2YTZitin2Kog2KfZhNiq2LHZgdmK2YfZitipAg8zMTE1ODgxMDcyMDAwMDMDEzIwMjUtMDktMjVUMDA6MDA6MDBaBAY5NDguNzUFBjEyMy43NQ=="))
		case "gr-mydata-v1":
			out = append(out,
				stamp("iapr-qr", "https://test.vs.gr/iv/invoice/10cc0a7d90a1c9470190a634ec19073e"),
				stamp("iapr-hash", "792F62ADE1B5E4CCBC8FFA787258CC68E90B817B"),
				stamp("iapr-mark", "400001929816893"),
				stamp("iapr-provider", "ΙΛΥΔΑ ΠΛΗΡΟΦΟΡΙΚΗ Α.Ε. — https://test.vs.gr"),
				stamp("iapr-uid", "EB1CF2137764B0D534885F361365C39A32C0205A"))
		case "co-dian-v2":
			const cude = "50d06ed88fb3f4f2d4a903c613db4d7b8bc550df7087d35091860dcf78a4a920846b2d16dd15ef4b8438e66f4857b887"
			out = append(out,
				stamp("dian-cude", cude),
				stamp("dian-qr", "NumFac: "+code+"\nFecFac: "+date+"\nNitFac: "+tax+"\nValTotal: "+total+"\nCUFE: "+cude+"\nhttps://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey="+cude),
				stamp("dian-provider", "EMPRESA DE DIVULGACIONES Y ASESORIAS ECA SAS"),
				stamp("dian-provider-nit", "860517022"),
				stamp("dian-software", "PLEMSI"))
		case "ar-arca-v4":
			out = append(out,
				stamp("arca-cae", "1234567891"),
				stamp("arca-cae-expiry", date),
				stamp("arca-qr", "https://www.afip.gob.ar/fe/qr/conceptos-generales.asp"))
		case "it-ticket-v1":
			out = append(out, stamp("ade-ref", "0000654321"))
		}
	}
	return out
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
	b.WriteString("# Examples imported from docs.invopop.com by scripts/build-examples.go: example\tsnippet\n")
	for _, f := range files {
		fmt.Fprintf(&b, "%s\t%s\n", f, m[f])
	}
	os.WriteFile(manifest, []byte(b.String()), 0o644)
}
