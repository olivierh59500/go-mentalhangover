// Package assets embeds extracted production data without the original references.
package assets

import "embed"

// Files contains decoded disk transfers used by the native renderer.
//
//go:embed raw/*.bin raw/madness.mod
var Files embed.FS
