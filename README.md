# gobl.html

Generate HTML files from GOBL documents.

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2024 [Invopop Ltd.](https://invopop.com).

## Development

### Generate Templates

GOBL HTML uses [templ](https://templ.guide/) to define a set of components in Go. To generate the templates, run:

```bash
templ generate
```

During development, it can help massive to have hot reload to be able to make changes and see them quickly. You can do this using the following example command:

```bash
templ generate --watch --cmd="go run ./cmd/gobl.html serve --pdf prince"
```

### Testing

Tests are currently pretty limited. To ensure the basics are covered, the contents of the `examples` directory are converted to HTML, pretty printed, and output to the `examples/out` directory. The tests will ensure the output is as expected. To update the output test data run:

```bash
go test ./... --update
```
