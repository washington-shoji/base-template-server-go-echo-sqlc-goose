package webassets

import "embed"

// Files embeds templates and static assets for single-binary deploys.
//
//go:embed all:templates all:static
var Files embed.FS
