// Package display defines the HTML snippets that can be placed at named
// anchors of a rendered document, and the template variables available
// inside them.
//
// Anchors are the same hierarchical, dot-separated keys that WithHide uses
// to hide sections, so a caller can hide a section and put something else in
// its place with one vocabulary. Compliance content such as regime QR codes,
// the parties, the lines, the totals and the tax notes has no key and can
// neither be hidden nor replaced.
//
// Snippets are trusted: the author is the owner of the configuration, on
// the same level as the logo URL. They are not sanitised, except that
// script, iframe, object and embed elements are stripped, as a PDF can do
// nothing with them. Invoice data reaches a snippet only through the
// variables of Vars, which html/template escapes for the context they land
// in, so a supplier name or a payment URL can never break out of the
// snippet's markup.
package display

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"text/template/parse"
)

// Key identifies a section of the document with a dot-separated path.
type Key string

// Section keys that snippets may be anchored to. Each is also a key
// WithHide accepts.
const (
	KeyHeaderLogo             Key = "header.logo"
	KeyHeaderTitleBadges      Key = "header.title.badges"
	KeyHeaderSummaryOrdering  Key = "header.summary.ordering"
	KeyHeaderContactsDelivery Key = "header.contacts.delivery"
	KeyTaxes                  Key = "taxes"
	KeyPayment                Key = "payment"
	KeyPaymentInstructions    Key = "payment.instructions"
	KeyPaymentBankData        Key = "payment.instructions.bank-data"
	KeyPaymentOnline          Key = "payment.instructions.online"
	KeyPaymentTerms           Key = "payment.terms"
	KeyNotes                  Key = "notes"
	KeyFooterBranding         Key = "footer.branding"
)

// KeyLinesSublines is a WithHide key with no anchor: the line sublines
// repeat for every line of the document, so a snippet there would be
// multiplied by the number of lines and could blow the rendered size
// far past the limits Validate enforces on the configuration.
const KeyLinesSublines Key = "lines.sublines"

var keys = []Key{
	KeyHeaderLogo,
	KeyHeaderTitleBadges,
	KeyHeaderSummaryOrdering,
	KeyHeaderContactsDelivery,
	KeyTaxes,
	KeyPayment,
	KeyPaymentInstructions,
	KeyPaymentBankData,
	KeyPaymentOnline,
	KeyPaymentTerms,
	KeyNotes,
	KeyFooterBranding,
}

// Keys provides the complete list of anchor keys, in document order.
func Keys() []Key {
	return slices.Clone(keys)
}

// Valid returns true if the key is an anchor snippets may be placed at.
func (k Key) Valid() bool {
	return slices.Contains(keys, k)
}

// String provides the key as a string.
func (k Key) String() string {
	return string(k)
}

// Mode determines where a snippet is placed relative to its anchor.
type Mode string

// Supported snippet modes.
const (
	// ModeBefore places the snippet before the section.
	ModeBefore Mode = "before"
	// ModeAfter places the snippet after the section.
	ModeAfter Mode = "after"
	// ModeReplace leaves out the section's own content and renders the
	// snippet in its place. Nothing inside the section is rendered, not
	// even snippets anchored to its descendants.
	ModeReplace Mode = "replace"
)

var modes = []Mode{ModeBefore, ModeAfter, ModeReplace}

// Valid returns true if the mode is supported.
func (m Mode) Valid() bool {
	return slices.Contains(modes, m)
}

// Size limits for snippets, applied to the configured HTML before the
// variables are filled in.
const (
	// MaxSnippetSize is the maximum size in bytes of a single snippet.
	MaxSnippetSize = 32 * 1024
	// MaxTotalSize is the maximum combined size in bytes of all snippets.
	MaxTotalSize = 128 * 1024
)

// Snippet is a block of trusted HTML placed at an anchor. The HTML is a Go
// html/template: write {{ .PaymentLink }} to insert a variable, or
// {{ if .PaymentLink }}…{{ end }} to leave the block out when there is
// nothing to insert. See Vars for the variables.
type Snippet struct {
	// At is the anchor key.
	At Key `json:"at"`
	// Mode is the placement relative to the anchor.
	Mode Mode `json:"mode"`
	// HTML is the trusted snippet template.
	HTML string `json:"html"`
}

// Snippets is the list of snippets to render, in order. It unmarshals from
// a JSON array, or from a JSON string holding one, so that it can travel as
// a single query parameter or form field as well as inside a document.
type Snippets []*Snippet

// UnmarshalJSON accepts either the array itself or a string containing the
// array as JSON.
func (s *Snippets) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) > 0 && data[0] == '"' {
		var src string
		if err := json.Unmarshal(data, &src); err != nil {
			return err
		}
		return s.UnmarshalParam(src)
	}
	type raw Snippets
	return json.Unmarshal(data, (*raw)(s))
}

// UnmarshalParam parses the JSON array of snippets held in a query or form
// parameter, as used by Echo's binder. An empty value clears the list.
func (s *Snippets) UnmarshalParam(src string) error {
	src = strings.TrimSpace(src)
	if src == "" {
		*s = nil
		return nil
	}
	var list []*Snippet
	if err := json.Unmarshal([]byte(src), &list); err != nil {
		return fmt.Errorf("snippets: %w", err)
	}
	*s = list
	return nil
}

// Validate checks the snippets strictly: an unknown key or mode, a key
// with no anchor, more than one replacement at the same anchor, a snippet
// over MaxSnippetSize or a set over MaxTotalSize, a template that does not
// parse or whose markup html/template cannot escape, and a variable that
// Vars does not define are all errors.
func (s Snippets) Validate() error {
	var errs []error
	total := 0
	replaced := make(map[Key]bool)
	for i, sn := range s {
		if sn == nil {
			errs = append(errs, fmt.Errorf("%d: missing", i))
			continue
		}
		if !sn.At.Valid() {
			if sn.At == KeyLinesSublines {
				errs = append(errs, fmt.Errorf("%d: at: '%s' repeats per line and cannot anchor a snippet", i, sn.At))
			} else {
				errs = append(errs, fmt.Errorf("%d: at: unknown key '%s'", i, sn.At))
			}
		}
		if !sn.Mode.Valid() {
			errs = append(errs, fmt.Errorf("%d: mode: unknown mode '%s'", i, sn.Mode))
		}
		if sn.Mode == ModeReplace {
			if replaced[sn.At] {
				errs = append(errs, fmt.Errorf("%d: duplicate replace at '%s'", i, sn.At))
			}
			replaced[sn.At] = true
		}
		if len(sn.HTML) > MaxSnippetSize {
			errs = append(errs, fmt.Errorf("%d: html: exceeds %d bytes", i, MaxSnippetSize))
		}
		total += len(sn.HTML)
		if _, err := sn.execute(sampleVars()); err != nil {
			errs = append(errs, fmt.Errorf("%d: html: %w", i, err))
		}
	}
	if total > MaxTotalSize {
		errs = append(errs, fmt.Errorf("total html exceeds %d bytes", MaxTotalSize))
	}
	return errors.Join(errs...)
}

// Render fills the variables into every snippet and groups the results by
// anchor. Snippets at unknown keys or in unknown modes are left out before
// any other check, so a configuration written for a newer renderer still
// works; a snippet that would render but is over the size limits, does not
// parse, or uses a variable Vars does not define is an error, as nothing
// should render with a hole in it.
func (s Snippets) Render(vars *Vars) (*Rendered, error) {
	if len(s) == 0 {
		return nil, nil
	}
	if vars == nil {
		vars = new(Vars)
	}
	r := &Rendered{anchors: make(map[Key]*anchor)}
	total := 0
	for i, sn := range s {
		if sn == nil || !sn.At.Valid() || !sn.Mode.Valid() {
			continue
		}
		if len(sn.HTML) > MaxSnippetSize {
			return nil, fmt.Errorf("snippets: %d: html: exceeds %d bytes", i, MaxSnippetSize)
		}
		total += len(sn.HTML)
		if total > MaxTotalSize {
			return nil, fmt.Errorf("snippets: total html exceeds %d bytes", MaxTotalSize)
		}
		out, err := sn.execute(vars)
		if err != nil {
			return nil, fmt.Errorf("snippets: %d: html: %w", i, err)
		}
		a := r.anchors[sn.At]
		if a == nil {
			a = new(anchor)
			r.anchors[sn.At] = a
		}
		switch sn.Mode {
		case ModeBefore:
			a.before = append(a.before, out)
		case ModeAfter:
			a.after = append(a.after, out)
		case ModeReplace:
			if a.replace == nil {
				a.replace = &out
			}
		}
	}
	return r, nil
}

// execute renders the snippet with the variables.
func (sn *Snippet) execute(vars *Vars) (string, error) {
	t, err := sn.parse()
	if err != nil {
		return "", err
	}
	buf := new(strings.Builder)
	if err := t.Execute(buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// parse strips the elements a PDF cannot use, parses the snippet as an
// html/template and checks that every field it refers to is a variable.
func (sn *Snippet) parse() (*template.Template, error) {
	t, err := template.New(string(sn.At)).Option("missingkey=error").Parse(Strip(sn.HTML))
	if err != nil {
		return nil, err
	}
	if t.Tree != nil {
		if err := checkFields(t.Tree.Root); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// stripped matches the elements removed from every snippet: a PDF runs no
// scripts and embeds no frames or plugins, and the renderer should never be
// asked to fetch them. An unterminated script, iframe or object drops the
// rest of the snippet; embed is a void element, so only its tag goes.
var stripped = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<embed\b[^>]*>`),
	regexp.MustCompile(`(?is)<(script|iframe|object)\b[^>]*>.*?</(script|iframe|object)\s*>`),
	regexp.MustCompile(`(?is)<(script|iframe|object)\b[^>]*>.*`),
	regexp.MustCompile(`(?is)</(script|iframe|object|embed)\s*>`),
}

// Strip removes the script, iframe, object and embed elements from the
// HTML, with their content. Render applies it to every snippet.
func Strip(html string) string {
	for _, re := range stripped {
		html = re.ReplaceAllString(html, "")
	}
	return html
}

// varFields are the names of the variables, taken from Vars.
var varFields = func() map[string]bool {
	fields := make(map[string]bool)
	t := reflect.TypeOf(Vars{})
	for i := 0; i < t.NumField(); i++ {
		fields[t.Field(i).Name] = true
	}
	return fields
}()

// sampleVars fills every variable, so that validation executes a snippet
// the way a complete invoice would.
func sampleVars() *Vars {
	v := new(Vars)
	rv := reflect.ValueOf(v).Elem()
	for i := 0; i < rv.NumField(); i++ {
		rv.Field(i).SetString("sample")
	}
	return v
}

// VarNames lists the variables available inside snippets, in the order
// Vars declares them.
func VarNames() []string {
	t := reflect.TypeOf(Vars{})
	names := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		names = append(names, t.Field(i).Name)
	}
	return names
}

// checkFields walks the parsed template and rejects any field access that
// is not a variable of Vars, so that a typo fails validation instead of
// executing against a value that does not exist.
func checkFields(node parse.Node) error {
	switch n := node.(type) {
	case nil:
		return nil
	case *parse.ListNode:
		if n == nil {
			return nil
		}
		for _, c := range n.Nodes {
			if err := checkFields(c); err != nil {
				return err
			}
		}
	case *parse.ActionNode:
		return checkFields(n.Pipe)
	case *parse.PipeNode:
		if n == nil {
			return nil
		}
		for _, c := range n.Cmds {
			for _, a := range c.Args {
				if err := checkFields(a); err != nil {
					return err
				}
			}
		}
	case *parse.IfNode:
		return checkBranch(&n.BranchNode)
	case *parse.RangeNode:
		return checkBranch(&n.BranchNode)
	case *parse.WithNode:
		return checkBranch(&n.BranchNode)
	case *parse.TemplateNode:
		return checkFields(n.Pipe)
	case *parse.ChainNode:
		if err := checkFields(n.Node); err != nil {
			return err
		}
		return checkIdents(n.Field)
	case *parse.FieldNode:
		return checkIdents(n.Ident)
	case *parse.VariableNode:
		// $ is the data itself, so $.Foo is a field access too.
		if len(n.Ident) > 1 && n.Ident[0] == "$" {
			return checkIdents(n.Ident[1:])
		}
	}
	return nil
}

func checkBranch(b *parse.BranchNode) error {
	if err := checkFields(b.Pipe); err != nil {
		return err
	}
	if err := checkFields(b.List); err != nil {
		return err
	}
	return checkFields(b.ElseList)
}

func checkIdents(idents []string) error {
	if len(idents) == 0 {
		return nil
	}
	if !varFields[idents[0]] {
		return fmt.Errorf("unknown variable '%s'", strings.Join(idents, "."))
	}
	if len(idents) > 1 {
		return fmt.Errorf("variable '%s' has no field '%s'", idents[0], idents[1])
	}
	return nil
}

// Rendered holds the snippets with their variables filled in, grouped by
// anchor, ready for the templates to place. A nil Rendered has nothing to
// place anywhere.
type Rendered struct {
	anchors map[Key]*anchor
}

type anchor struct {
	before  []string
	replace *string
	after   []string
}

func (r *Rendered) at(key Key) *anchor {
	if r == nil {
		return nil
	}
	return r.anchors[key]
}

// Before provides the HTML to render before the key, in order.
func (r *Rendered) Before(key Key) []string {
	if a := r.at(key); a != nil {
		return a.before
	}
	return nil
}

// After provides the HTML to render after the key, in order.
func (r *Rendered) After(key Key) []string {
	if a := r.at(key); a != nil {
		return a.after
	}
	return nil
}

// Replacement provides the HTML that takes the place of the key's own
// content, if any.
func (r *Rendered) Replacement(key Key) (string, bool) {
	if a := r.at(key); a != nil && a.replace != nil {
		return *a.replace, true
	}
	return "", false
}

// Has returns true when anything would be rendered at the key, which the
// templates use to render a container that would otherwise be left out.
func (r *Rendered) Has(key Key) bool {
	a := r.at(key)
	return a != nil && (len(a.before) > 0 || len(a.after) > 0 || a.replace != nil)
}
