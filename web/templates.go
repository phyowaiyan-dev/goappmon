package web

import "embed"

// TemplatesFS embeds the admin UI templates into the binary.
//
//go:embed templates/**/*.html
var TemplatesFS embed.FS
