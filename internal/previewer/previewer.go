// Package previewer holds the page the development server serves at /: a
// filterable catalog of the examples rendered in the browser through the
// WebAssembly build, with the server render a click away.
package previewer

import _ "embed"

// IndexHTML is the previewer page.
//
//go:embed index.html
var IndexHTML []byte
