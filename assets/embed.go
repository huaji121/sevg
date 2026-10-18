package assets

import "embed"

// Textures contains categorized block, entity, sky, palette, effect and item sprites. The files are
// embedded so the game remains a single self-contained executable.
//
//go:embed textures/blocks/*.png textures/items/*.png textures/entities/*.png textures/sky/*.png textures/palette/*.png textures/effects/*.png
var Textures embed.FS

// Sounds contains the embedded sound effects.
//
//go:embed sounds/*.wav
var Sounds embed.FS

// Lang contains the UI and content translation tables.
//
//go:embed lang/*.json
var Lang embed.FS