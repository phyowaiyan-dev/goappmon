package web

import "embed"

// PostmanFS embeds sample API collections into the binary.
//
//go:embed postman/*.json
var PostmanFS embed.FS
