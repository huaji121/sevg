// Package ui provides a tiny immediate-mode 2D canvas and bitmap font.
package ui

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	CellW     = 8
	CellH     = 14
	Cols      = 16
	Rows      = 8
	First     = 32
	Last      = 126
	WhiteCell = 127
)

type FontAtlas struct {
	Pixels  []byte
	W, H    int
	WhiteUV [2]float32
}

func BuildFontAtlas() FontAtlas {
	w, h := Cols*CellW, Rows*CellH
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	face := basicfont.Face7x13
	for r := rune(First); r <= Last; r++ {
		cell := int(r - First)
		cx, cy := cell%Cols, cell/Cols
		d := &font.Drawer{
			Dst:  img,
			Src:  image.White,
			Face: face,
			Dot:  fixed.P(cx*CellW, cy*CellH+11),
		}
		d.DrawString(string(r))
	}
	// A solid white texel for panels and bars.
	cell := WhiteCell
	cx, cy := cell%Cols, cell/Cols
	for y := 0; y < CellH; y++ {
		for x := 0; x < CellW; x++ {
			img.SetRGBA(cx*CellW+x, cy*CellH+y, color.RGBA{255, 255, 255, 255})
		}
	}
	white := [2]float32{
		(float32(cx*CellW) + CellW/2) / float32(w),
		(float32(cy*CellH) + CellH/2) / float32(h),
	}
	return FontAtlas{Pixels: img.Pix, W: w, H: h, WhiteUV: white}
}

func cellUV(r rune) (u0, v0, u1, v1 float32) {
	if r < First || r > Last {
		r = '?'
	}
	cell := int(r - First)
	if cell >= WhiteCell {
		cell = '?' - First
	}
	cx, cy := cell%Cols, cell/Cols
	const eps = 0.01
	u0 = (float32(cx*CellW) + eps) / float32(Cols*CellW)
	v0 = (float32(cy*CellH) + eps) / float32(Rows*CellH)
	u1 = (float32((cx+1)*CellW) - eps) / float32(Cols*CellW)
	v1 = (float32((cy+1)*CellH) - eps) / float32(Rows*CellH)
	return
}

type Vertex struct {
	X, Y       float32
	U, V       float32
	R, G, B, A float32
}

type Rect struct{ X, Y, W, H float32 }

func (r Rect) Contains(x, y float32) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

type Canvas struct {
	W, H  int
	Verts []Vertex
	White [2]float32
}

func NewCanvas(w, h int, atlas FontAtlas) *Canvas {
	return &Canvas{W: w, H: h, White: atlas.WhiteUV}
}

func (c *Canvas) Resize(w, h int) {
	c.W, c.H = w, h
}

func (c *Canvas) Reset() {
	c.Verts = c.Verts[:0]
}

func (c *Canvas) quad(x, y, w, h float32, u0, v0, u1, v1 float32, col [4]float32) {
	if w <= 0 || h <= 0 {
		return
	}
	x1, y1 := x+w, y+h
	c.Verts = append(c.Verts,
		Vertex{x, y, u0, v0, col[0], col[1], col[2], col[3]},
		Vertex{x1, y, u1, v0, col[0], col[1], col[2], col[3]},
		Vertex{x1, y1, u1, v1, col[0], col[1], col[2], col[3]},
		Vertex{x, y, u0, v0, col[0], col[1], col[2], col[3]},
		Vertex{x1, y1, u1, v1, col[0], col[1], col[2], col[3]},
		Vertex{x, y1, u0, v1, col[0], col[1], col[2], col[3]},
	)
}

func (c *Canvas) Rect(x, y, w, h float32, col [4]float32) {
	c.quad(x, y, w, h, c.White[0], c.White[1], c.White[0], c.White[1], col)
}

func (c *Canvas) Border(x, y, w, h, t float32, col [4]float32) {
	c.Rect(x, y, w, t, col)
	c.Rect(x, y+h-t, w, t, col)
	c.Rect(x, y, t, h, col)
	c.Rect(x+w-t, y, t, h, col)
}

func (c *Canvas) Panel(x, y, w, h float32, bg, border [4]float32) {
	c.Rect(x+3, y+3, w, h, [4]float32{0, 0, 0, 0.35})
	c.Rect(x, y, w, h, bg)
	c.Border(x, y, w, h, 2, border)
}

func (c *Canvas) Text(x, y float32, s string, scale float32, col [4]float32) float32 {
	cx, cy := x, y
	startX := x
	for _, r := range s {
		if r == 10 {
			cx = startX
			cy += float32(CellH) * scale
			continue
		}
		u0, v0, u1, v1 := cellUV(r)
		c.quad(cx, cy, float32(CellW)*scale, float32(CellH)*scale, u0, v0, u1, v1, col)
		cx += float32(CellW) * scale * 0.875
	}
	return cx
}

func (c *Canvas) TextShadow(x, y float32, s string, scale float32, col [4]float32) {
	c.Text(x+scale, y+scale, s, scale, [4]float32{0, 0, 0, col[3] * 0.75})
	c.Text(x, y, s, scale, col)
}

func (c *Canvas) TextCentered(cx, y float32, s string, scale float32, col [4]float32) {
	w := c.TextWidth(s, scale)
	c.Text(cx-w/2, y, s, scale, col)
}

func (c *Canvas) TextShadowCentered(cx, y float32, s string, scale float32, col [4]float32) {
	w := c.TextWidth(s, scale)
	c.TextShadow(cx-w/2, y, s, scale, col)
}

func (c *Canvas) TextWidth(s string, scale float32) float32 {
	max, line := float32(0), float32(0)
	for _, r := range s {
		if r == 10 {
			if line > max {
				max = line
			}
			line = 0
			continue
		}
		line += float32(CellW) * scale * 0.875
	}
	if line > max {
		max = line
	}
	return max
}

func RGBA(r, g, b, a uint8) [4]float32 {
	return [4]float32{float32(r) / 255, float32(g) / 255, float32(b) / 255, float32(a) / 255}
}

var (
	TextColor  = RGBA(238, 244, 248, 255)
	MutedColor = RGBA(168, 180, 190, 255)
	Accent     = RGBA(255, 196, 82, 255)
	Good       = RGBA(115, 220, 130, 255)
	Bad        = RGBA(238, 92, 92, 255)
)
