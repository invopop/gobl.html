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