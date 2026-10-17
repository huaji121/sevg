package assets

import "embed"

// Textures contains the editable block and item texture assets. The files are
// embedded so the game remains a single self-contained executable.
//
//go:embed textures/blocks/*.png textures/items/*.png
var Textures embed.FS
