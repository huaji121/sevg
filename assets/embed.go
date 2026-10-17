package assets

import "embed"

// Textures contains editable block textures and flat item sprites. The files are
// embedded so the game remains a single self-contained executable.
//
//go:embed textures/blocks/*.png textures/items/*.png
var Textures embed.FS

// Sounds contains the embedded sound effects.
//
//go:embed sounds/*.wav
var Sounds embed.FS

// Lang contains the UI and content translation tables.
//
//go:embed lang/*.json
var Lang embed.FS
