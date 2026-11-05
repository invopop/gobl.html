# gobl.html

Generate HTML files from GOBL documents.

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2024 [Invopop Ltd.](https://invopop.com).

## Development

### Generate Templates

GOBL HTML uses [templ](https://templ.guide/) to define a set of components in Go. To generate the templates, run:

```bash
templ generate
```

Hot reload is very convenient to make changes and preview them quickly. There are two mechanisms we're currently using:

#### Air

Air is a great tool to auto reload potentially any project, but works great with Go. Install with:

```bash
go install github.com/air-verse/air@latest
```

The `.toml` is already configured and ready in this repository, so simply run:

```bash
air
```

Air is a bit more reliable at detecting file changes, especially for stylesheets. You'll always need to wait a few seconds before page reloads to give the system chance to recompile. A proxy is available with Air, but we didn't find it to be very reliable and was breaking with query parameters, it obviously also wouldn't work for PDF reloads.

#### Templ Watcher

Templ comes with a watch flag that can also be useful. However, it has the disadvantage of using `.txt` files for comparisons and the generated code should not be uploaded to git directly. Start the process with:

```bash
templ generate --watch --cmd="go run ./cmd/gobl.html serve --pdf prince"
```

Before uploading changes to git, be sure to re-run the regular `templ generate` command, as the live version makes temporary modifications to files that need to be replaced.

### WebAssembly

The renderer also compiles to WebAssembly so a browser can render a GOBL
envelope without a server round trip. The `wasm` package exposes:

```js
goblRender(envelopeJSON, { locale: "es", watermark: "DRAFT" }) // → { html } or { error }
goblVersion()                                                  // → { version, gobl }
```

Parameters use the same names as pdf.go's task configuration (see `Params` in
`params.go`). Documents are rendered as given, without calculation, and assets
are embedded unless `embed_assets` is false.

Build it with:

```bash
scripts/cdn-dist.sh dist
```

The development server (`go run ./cmd/gobl.html serve`) opens a previewer
at `/`, built with [PopUI](https://github.com/invopop/popui.go), that lists
the examples with filters for country, document kind, addon and content (QR
codes, payment details, logos, …) and renders the selected one in the browser
through the WebAssembly build on a true-ratio sheet, with page breaks, an
optional DIN 5008 envelope window and links to the server HTML and PDF. A
benchmark button renders the listed examples both ways and compares the
times. After editing `internal/previewer/previewer.templ`, regenerate with
the templ version from `go.mod`:

```bash
go run github.com/a-h/templ/cmd/templ generate -path internal/previewer
```

The `wasm` package test builds the module and checks,
through Node, that every example renders byte-identically to the native
renderer.

Each release publishes the build to
`https://assets.invopop.com/gobl.html/<version>/`.

### Hiding sections and placing snippets

`WithHide` removes named sections from the document with CSS, and
`WithSnippets` places trusted blocks of HTML at the same named places.
Both use the keys below, so a caller can hide a section and put something
else in its place with one vocabulary. As task configuration (`Params`)
they are the `hide` list and the `snippets` array:

```json
{
  "hide": "payment.instructions.bank-data,footer.branding",
  "snippets": [
    {
      "at": "payment.instructions",
      "mode": "after",
      "html": "{{ if .PaymentLink }}<a href=\"{{ .PaymentLink }}\">Pay {{ .TotalDue }} now</a>{{ end }}"
    }
  ]
}
```

| Key | Section |
| --- | --- |
| `header.logo` | Supplier logo, or the alias when there is no logo |
| `header.title.badges` | Paid, void, self-billed and label badges next to the code |
| `header.summary.ordering` | Ordering rows of the summary (order code, period, identities) |
| `header.contacts.delivery` | Delivery receiver contact |
| `lines.sublines` | Line references, identities and notes (`hide` only, see below) |
| `taxes` | Tax breakdown table |
| `payment` | The whole payment section |
| `payment.instructions` | Payment instructions list |
| `payment.instructions.bank-data` | Credit transfer details |
| `payment.instructions.online` | Online payment links, listed and as QR codes |
| `payment.terms` | Payment terms and due dates |
| `notes` | The document's notes; a replacement keeps the regime's tax notes |
| `footer.branding` | GOBL logo in the footer |

The title, the supplier and customer, the lines, the totals, the regime
extensions, the tax notes, the page numbers and every regime QR code have
no key: they are compliance content and cannot be hidden or replaced.

Hiding cascades by prefix, so `payment` also hides `payment.terms`.
Unknown keys are ignored when rendering; `Snippets.Validate` rejects them,
which is what a service taking the configuration in should call.

#### Snippets

A snippet names an anchor key, a mode and the HTML to place:

- `before` and `after` render the snippet on either side of the section,
  even when the section is hidden or has no content for this document, in
  the order given. `payment.instructions` and `payment.terms` get their
  payment section created for them when the document has none.
- `replace` renders the snippet instead of the section's own content.
  Nothing inside the section is rendered, not even snippets anchored to keys
  within it. At most one replacement per anchor.

Each snippet is wrapped as `<div class="snippet snippet-<mode>"
data-at="<key>">`, or `<li>` at the anchors inside a list (ordering,
bank data, online). The title repeats in the running header of later pages
and the footer has a print and a screen variant, so a snippet anchored
there renders once per copy. `lines.sublines` repeats per line and has no
anchor.

Snippet HTML is trusted, on the same level as the logo URL: the author is
the owner of the configuration and the markup is not sanitised, except that
`script`, `iframe`, `object` and `embed` elements are stripped, as a PDF
can do nothing with them. A single snippet is limited to 32 KiB and a set
to 128 KiB; the snippets of a document render once, whatever its length.

The HTML is a Go `html/template` over a fixed set of variables, filled from
the document and formatted the way the document shows them, with the
locale, date and number formats of the render. Every value is escaped for
the place it lands in, as text, inside an attribute or as a URL, so
document data can never break out of the snippet's markup. A variable that
does not apply is empty, which `{{ if .DueDate }}…{{ end }}` guards for;
a variable that does not exist fails validation and rendering.

| Variable | Value |
| --- | --- |
| `.Code` | Document code |
| `.Series` | Document series, or empty |
| `.Number` | Series and code joined as the title shows them |
| `.IssueDate` | Issue date |
| `.DueDate` | First due date of the invoice's payment terms, or empty |
| `.TotalPayable` | Total payable with its currency, before advances |
| `.TotalDue` | Amount left to pay with its currency, after advances |
| `.Currency` | Currency code |
| `.SupplierName` | Supplier's name |
| `.CustomerName` | Customer's name, or empty |
| `.PaymentLink` | First online payment URL in the instructions, or empty |

Only an invoice fills `.DueDate`, `.TotalPayable`, `.TotalDue` and
`.PaymentLink`. `examples/es-verifactu-invoice-full.snippets.json` renders
with the set in `internal/catalog/snippets.json`, a "Pay now" button after
the payment instructions and a contact line after the notes, on every
path that renders the examples; the previewer's options also take a set of
snippets to try.

### Examples

Every file in `examples/` is named `<country>-<format>-<kind>[-<variant>]`,
with states and locales as dotted infixes: the regime's country, the agency
or format the addon targets (`verifactu`, `sat`, `at`, `sdi`, `peppol`…),
the document kind (`invoice`, `credit-note`, `debit-note`, `corrective`,
`proforma`, `order`, `delivery`, `receipt`, `party`…) and what else sets
the example apart, as in `es-verifactu-invoice-full` or
`mx-sat-invoice-b2c.hidden`. `scripts/build-examples.go` derives the name
from the document, and `-rename` moves every example, with its output,
source and manifest entry, to its canonical name.

An example can say what it shows: the envelope head's `meta` carries a
`title` and a `description`, which the previewer lists and which the
digest does not cover. The docs import takes them from the accordion
each snippet appears in; a source under `examples/src/` gives them by
wrapping the document as `{"head": {"meta": {"title": …, "description":
…}}, "doc": {…}}`.

### Docs examples

Most of `examples/` comes from the GOBL examples published on
[docs.invopop.com](https://docs.invopop.com), so the previewer renders every
document the docs show; gobl.html keeps more of its own on top. The docs are
the source of those files: edit the `.min.mdx` snippet there and re-import
rather than editing the copy here. With a checkout of the docs repository
next to this one:

```bash
go run scripts/build-examples.go -docs ../docs.invopop.com
go test . -run Example -update
```

The script envelops and validates each built snippet with the gobl pinned
here, writes `examples/<country>-<name>.json`, records what it imported in
`examples/docs-examples.txt`; with `-prune` it also removes imports the
docs dropped (without it they are kept and reported, since they may come
from a docs branch not merged yet). It skips a
docs example whose name an example of gobl.html's own already uses, and
reports one the current gobl rules reject, so both show up as work for the
docs. A document whose addon has the tax agency stamp it (VERI\*FACTU,
TicketBAI, AT, SAT, KSeF, ZATCA, myDATA, DIAN, ARCA, Smart Receipts) gets
sample stamps shaped like the real ones, a code if it has none, and a
signature from a throwaway key, so the previewer shows the QR codes and
references the final PDF carries.

### Testing

Tests are currently pretty limited. To ensure the basics are covered, the contents of the `examples` directory are converted to HTML, pretty printed, and output to the `examples/out` directory. The tests will ensure the output matches what's expected. To update the output test data run:

```bash
go test ./... --update
```

When adding new examples make sure to run:

```bash
go test examples_test.go --update
```

This will ensure the test file for that example is updated.