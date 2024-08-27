# gobl.html

Generate HTML files from GOBL documents.

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2024 [Invopop Ltd.](https://invopop.com).

## Development

### Generate Templates

GOBL HTML uses [templ](https://templ.guide/) to define a set of components in Go. To generate the templates, run:

```bash
templ generate
```

During development, it can help massive to have hot reload to be able to make changes and see them quickly. There are two mechanisms we're currently using:

#### Air

Air is a great tool to auto reload potentially any project, but works great with Go. Install with:

```bash
go install github.com/air-verse/air@latest
```

The `.toml` is already configured and ready in this repository, so simply run:

```bash
air
```

Air is a bit more reliable at detecting file changes, especially for stylesheets. You'll always need to wait a few seconds before page reloads to give the system chance to recompile. A proxy is available with Air, but we didn't find it to be very reliable and was breaking with query parameters, it obvously also wouldn't work for PDF reloads.

#### Templ Watcher

Templ comes with a watch flag that can also be useful. It has the disadvantage however that it uses `.txt` files for comparisons and the generated code should not be uploaded to git directly. Start the process with:

```bash
templ generate --watch --cmd="go run ./cmd/gobl.html serve --pdf prince"
```

Before uploading changes to git, be sure to re-run the regular `templ generate` command, as the live version makes temporary modifications to files that need to be replaced.

### Testing

Tests are currently pretty limited. To ensure the basics are covered, the contents of the `examples` directory are converted to HTML, pretty printed, and output to the `examples/out` directory. The tests will ensure the output is as expected. To update the output test data run:

```bash
go test ./... --update
```
