// Package display defines the options used to control which sections of a
// rendered document are shown, and to inject custom HTML snippets at
// well-known anchor points.
//
// Sections are identified by hierarchical, dot-separated keys. Hiding a key
// also hides all of its descendants, so hiding "payment" removes the payment
// instructions, bank data, and terms. Sections are removed at render time,
// so hidden data never reaches the output.
//
// Compliance content, such as regime QR codes and tax notes, has no keys and
// can never be hidden.
package display

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Key identifies a section of the document using a dot-separated path.
type Key string

// Section keys that may be hidden, replaced, or used as anchors for custom
// HTML snippets.
const (
	KeyHeaderLogo             Key = "header.logo"
	KeyHeaderTitleBadges      Key = "header.title.badges"
	KeyHeaderSummaryOrdering  Key = "header.summary.ordering"
	KeyHeaderContactsDelivery Key = "header.contacts.delivery"
	KeyLinesItemNotes         Key = "lines.item-notes"
	KeyTaxes                  Key = "taxes"
	KeyPayment                Key = "payment"
	KeyPaymentInstructions    Key = "payment.instructions"
	KeyPaymentBankData        Key = "payment.instructions.bank-data"
	KeyPaymentOnline          Key = "payment.instructions.online"
	// KeyPaymentQR is reserved for the upcoming payment QR code. It validates,
	// but nothing is rendered at this anchor yet.
	KeyPaymentQR      Key = "payment.qr"
	KeyPaymentTerms   Key = "payment.terms"
	KeyNotesGeneral   Key = "notes.general"
	KeyFooterBranding Key = "footer.branding"
)

var keys = []Key{
	KeyHeaderLogo,
	KeyHeaderTitleBadges,
	KeyHeaderSummaryOrdering,
	KeyHeaderContactsDelivery,
	KeyLinesItemNotes,
	KeyTaxes,
	KeyPayment,
	KeyPaymentInstructions,
	KeyPaymentBankData,
	KeyPaymentOnline,
	KeyPaymentQR,
	KeyPaymentTerms,
	KeyNotesGeneral,
	KeyFooterBranding,
}

// Keys provides the complete list of supported section keys.
func Keys() []Key {
	return slices.Clone(keys)
}

// In returns true if the key is one of the supported section keys.
func (k Key) In(list ...Key) bool {
	return slices.Contains(list, k)
}

// Valid returns true if the key is supported.
func (k Key) Valid() bool {
	return k.In(keys...)
}

// ancestorOf returns true if k is a strict dot-prefix ancestor of other.
func (k Key) ancestorOf(other Key) bool {
	return strings.HasPrefix(string(other), string(k)+".")
}

// Mode determines where a custom HTML snippet is placed relative to the
// section identified by its anchor key.
type Mode string

// Supported HTML snippet modes.
const (
	// ModeBefore places the snippet before the section.
	ModeBefore Mode = "before"
	// ModeAfter places the snippet after the section.
	ModeAfter Mode = "after"
	// ModeReplace hides the section's native content and renders the snippet
	// in its place.
	ModeReplace Mode = "replace"
)

var modes = []Mode{ModeBefore, ModeAfter, ModeReplace}

// Size limits for custom HTML snippets.
const (
	// MaxSnippetSize is the maximum size in bytes of a single snippet.
	MaxSnippetSize = 32 * 1024
	// MaxTotalSize is the maximum combined size in bytes of all snippets.
	MaxTotalSize = 128 * 1024
)

// Options defines which sections of the document should be hidden, and
// the custom HTML snippets to inject.
type Options struct {
	// Hide lists the sections to remove, including all their descendants.
	Hide []Key `json:"hide,omitempty"`
	// HTML lists the custom snippets to inject, rendered in order.
	HTML []*HTML `json:"html,omitempty"`
}

// HTML defines a raw HTML snippet to inject at an anchor key. Content is
// trusted and rendered as is.
type HTML struct {
	// At is the anchor section key.
	At Key `json:"at"`
	// Mode determines the placement relative to the anchor.
	Mode Mode `json:"mode"`
	// Content is the raw HTML to render.
	Content string `json:"content"`
}

// Validate checks the options strictly: unknown keys or modes, multiple
// replacements for the same anchor, and oversized snippets are all errors.
// Rendering itself is lenient and ignores unknown keys.
func (o *Options) Validate() error {
	if o == nil {
		return nil
	}
	var errs []error
	for i, k := range o.Hide {
		if !k.Valid() {
			errs = append(errs, fmt.Errorf("hide: %d: unknown key '%s'", i, k))
		}
	}
	total := 0
	replaced := make(map[Key]bool)
	for i, h := range o.HTML {
		if h == nil {
			errs = append(errs, fmt.Errorf("html: %d: missing", i))
			continue
		}
		if !h.At.Valid() {
			errs = append(errs, fmt.Errorf("html: %d: at: unknown key '%s'", i, h.At))
		}
		if !slices.Contains(modes, h.Mode) {
			errs = append(errs, fmt.Errorf("html: %d: mode: unknown mode '%s'", i, h.Mode))
		}
		if h.Mode == ModeReplace {
			if replaced[h.At] {
				errs = append(errs, fmt.Errorf("html: %d: duplicate replace for '%s'", i, h.At))
			}
			replaced[h.At] = true
		}
		if len(h.Content) > MaxSnippetSize {
			errs = append(errs, fmt.Errorf("html: %d: content: exceeds %d bytes", i, MaxSnippetSize))
		}
		total += len(h.Content)
	}
	if total > MaxTotalSize {
		errs = append(errs, fmt.Errorf("html: total content exceeds %d bytes", MaxTotalSize))
	}
	return errors.Join(errs...)
}

// AncestorHidden returns true when any strict ancestor of the key is hidden
// or replaced, in which case nothing at all should be rendered for the key,
// not even custom snippets.
func (o *Options) AncestorHidden(key Key) bool {
	if o == nil {
		return false
	}
	for _, k := range o.Hide {
		if k.ancestorOf(key) {
			return true
		}
	}
	for _, h := range o.HTML {
		if h != nil && h.Mode == ModeReplace && h.At.ancestorOf(key) {
			return true
		}
	}
	return false
}

// Hidden returns true when the key's native content should not be shown
// because the key, or any of its ancestors, is hidden, or an ancestor
// has been replaced.
func (o *Options) Hidden(key Key) bool {
	if o == nil {
		return false
	}
	return slices.Contains(o.Hide, key) || o.AncestorHidden(key)
}

// Suppressed returns true when the key's native content should not be
// rendered, either because it is hidden or replaced by a snippet.
func (o *Options) Suppressed(key Key) bool {
	if o.Hidden(key) {
		return true
	}
	_, ok := o.Replacement(key)
	return ok
}

// Replacement provides the replace snippet defined exactly at the key, if any.
func (o *Options) Replacement(key Key) (string, bool) {
	if o == nil {
		return "", false
	}
	for _, h := range o.HTML {
		if h != nil && h.Mode == ModeReplace && h.At == key {
			return h.Content, true
		}
	}
	return "", false
}

// Before provides the snippets to render before the key, in order.
func (o *Options) Before(key Key) []string {
	return o.snippets(key, ModeBefore)
}

// After provides the snippets to render after the key, in order.
func (o *Options) After(key Key) []string {
	return o.snippets(key, ModeAfter)
}

// HasSnippets returns true when any snippet would be rendered at the key,
// which is useful to render containers that would otherwise be empty.
func (o *Options) HasSnippets(key Key) bool {
	if o == nil || o.AncestorHidden(key) {
		return false
	}
	for _, h := range o.HTML {
		if h != nil && h.At == key && slices.Contains(modes, h.Mode) {
			return true
		}
	}
	return false
}

func (o *Options) snippets(key Key, mode Mode) []string {
	if o == nil {
		return nil
	}
	var out []string
	for _, h := range o.HTML {
		if h != nil && h.Mode == mode && h.At == key {
			out = append(out, h.Content)
		}
	}
	return out
}
