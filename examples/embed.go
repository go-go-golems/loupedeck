package examples

import "embed"

// ScriptsFS contains the built-in example JavaScript scene repository shipped
// with loupedeck.
//
//go:embed js/*.js
var ScriptsFS embed.FS
