# gobl.html

Generate HTML files from GOBL documents.

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl/blob/main/LICENSE), Copyright 2024 [Invopop Ltd.](https://invopop.com).

## Development

### Generate Templates

GOBL HTML uses [templ](https://templ.guide/) to define a set of components in Go. To genera the templates, run:

```bash
templ generate
```

During development, it can help massive to have hot reload to be able to make changes and see them quickly. You can do this using the following example command:

```bash
templ generate --watch --cmd="go run ./cmd/gobl.html serve --pdf prince"
```
