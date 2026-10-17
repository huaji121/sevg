// Package ui provides a small immediate-mode 2D canvas and a CJK-capable
// system-font atlas with a deterministic fallback bitmap font.
package ui

import (
	"fmt"
	"image"
	"os"
	"sort"

	"sevg/internal/core"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	fontSize   = 16.0
	atlasWidth = 2048
	cellPad    = 1
)

type Glyph struct {
	U0, V0, U1, V1   float32
	W, H             float32
	OffsetX, OffsetY float32
	Advance          float32
}

type IconSide struct {
	AX, AY, BX, BY float32
	U, V           float32
}

type FontAtlas struct {
	Pixels     []byte
	W, H       int
	WhiteUV    [2]float32
	Glyphs     map[rune]Glyph
	Icons      map[core.ItemID]Glyph
	IconSides  map[core.ItemID][]IconSide
	LineHeight float32
	Ascent     float32
	Source     string
}

type placement struct {
	r     rune
	item  core.ItemID
	g     Glyph
	alpha *image.Alpha
	rgba  *image.RGBA
}

// BuildFontAtlas rasterizes all requested runes into one texture. It uses a
// system CJK font when available (Microsoft YaHei, Noto Sans CJK, PingFang,
// etc.) and falls back to the embedded 7x13 bitmap font.
func BuildFontAtlas(runes []rune, blockAtlas []byte) FontAtlas {
	face, source := loadFace()
	if face != nil {
		defer face.Close()
	}

	set := make(map[rune]bool)
	for r := rune(32); r < 127; r++ {
		set[r] = true
	}
	for _, r := range runes {
		set[r] = true
	}
	set['?'] = true
	chars := make([]rune, 0, len(set))
	for r := range set {
		chars = append(chars, r)
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })

	metrics := face.Metrics()
	lineHeight := float32(metrics.Height) / 64
	ascent := float32(metrics.Ascent) / 64
	if lineHeight < 8 {
		lineHeight = 14
	}

	placements := make([]placement, 0, len(chars))
	for _, r := range chars {
		g, alpha, ok := rasterGlyph(face, r)
		if !ok {
			continue
		}
		placements = append(placements, placement{r: r, g: g, alpha: alpha})
	}
	itemImages := buildItemIcons(blockAtlas)
	for item := core.ItemID(1); item < core.ItemCount; item++ {
		if img := itemImages[item]; img != nil {
			placements = append(placements, placement{item: item, rgba: img})
		}
	}

	// Reserve a 2x2 white block in the top-left corner for panels and bars.
	const whiteSize = 2
	cursorX, cursorY, rowH := 4, 0, whiteSize
	for i := range placements {
		p := &placements[i]
		var w, h int
		switch {
		case p.alpha != nil:
			w, h = p.alpha.Rect.Dx(), p.alpha.Rect.Dy()
		case p.rgba != nil:
			w, h = p.rgba.Rect.Dx(), p.rgba.Rect.Dy()
		default:
			continue
		}
		if cursorX+w > atlasWidth {
			cursorX = 0
			cursorY += rowH + 1
			rowH = 0
		}
		p.g.W = float32(w)
		p.g.H = float32(h)
		p.g.OffsetX = float32(cursorX)
		p.g.OffsetY = float32(cursorY)
		cursorX += w + 1
		if h > rowH {
			rowH = h
		}
	}
	height := cursorY + rowH
	if height < 16 {
		height = 16
	}
	for i := range placements {
		p := &placements[i]
		var w, h int
		switch {
		case p.alpha != nil:
			w, h = p.alpha.Rect.Dx(), p.alpha.Rect.Dy()
		case p.rgba != nil:
			w, h = p.rgba.Rect.Dx(), p.rgba.Rect.Dy()
		default:
			continue
		}
		p.g.U0 = (p.g.OffsetX + 0.01) / atlasWidth
		p.g.V0 = (p.g.OffsetY + 0.01) / float32(height)
		p.g.U1 = (p.g.OffsetX + float32(w) - 0.01) / atlasWidth
		p.g.V1 = (p.g.OffsetY + float32(h) - 0.01) / float32(height)
	}

	pixels := make([]byte, atlasWidth*height*4)
	for y := 0; y < whiteSize; y++ {
		for x := 0; x < whiteSize; x++ {
			i := (y*atlasWidth + x) * 4
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = 255, 255, 255, 255
		}
	}
	glyphs := make(map[rune]Glyph, len(placements))
	icons := make(map[core.ItemID]Glyph)
	iconSides := make(map[core.ItemID][]IconSide)
	for i := range placements {
		p := &placements[i]
		px, py := int(p.g.OffsetX), int(p.g.OffsetY)
		if p.alpha != nil {
			w, h := p.alpha.Rect.Dx(), p.alpha.Rect.Dy()
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					a := p.alpha.AlphaAt(x, y).A
					if a == 0 {
						continue
					}
					i := ((py+y)*atlasWidth + (px + x)) * 4
					pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = 255, 255, 255, a
				}
			}
			// Restore the glyph's pixel offsets relative to the pen/line top.
			bounds, _, _ := face.GlyphBounds(p.r)
			p.g.OffsetX = float32(bounds.Min.X)/64 - cellPad
			p.g.OffsetY = ascent + float32(bounds.Min.Y)/64 - cellPad
			glyphs[p.r] = p.g
			continue
		}
		if p.alpha == nil && p.rgba == nil {
			if p.g.Advance > 0 {
				glyphs[p.r] = p.g
			}
			continue
		}
		if p.rgba != nil {
			for y := 0; y < p.rgba.Rect.Dy(); y++ {
				for x := 0; x < p.rgba.Rect.Dx(); x++ {
					src := p.rgba.Pix[y*p.rgba.Stride+x*4:]
					if src[3] == 0 {
						continue
					}
					dst := ((py+y)*atlasWidth + (px + x)) * 4
					copy(pixels[dst:dst+4], src[:4])
				}
			}
			icons[p.item] = p.g
			iconSides[p.item] = buildIconSides(p.rgba, p.g)
		}
	}

	whiteUV := [2]float32{1.0 / atlasWidth, 1.0 / float32(height)}
	return FontAtlas{
		Pixels: pixels, W: atlasWidth, H: height,
		WhiteUV: whiteUV, Glyphs: glyphs, Icons: icons, IconSides: iconSides,
		LineHeight: lineHeight, Ascent: ascent, Source: source,
	}
}

func rasterGlyph(face font.Face, r rune) (Glyph, *image.Alpha, bool) {
	bounds, advance, ok := face.GlyphBounds(r)
	if !ok {
		return Glyph{}, nil, false
	}
	adv := float32(advance) / 64
	if bounds.Max.X <= bounds.Min.X || bounds.Max.Y <= bounds.Min.Y {
		return Glyph{Advance: adv}, nil, true
	}
	w := bounds.Max.X.Ceil() - bounds.Min.X.Floor()
	h := bounds.Max.Y.Ceil() - bounds.Min.Y.Floor()
	if w <= 0 || h <= 0 {
		return Glyph{Advance: adv}, nil, true
	}
	img := image.NewAlpha(image.Rect(0, 0, w+cellPad*2, h+cellPad*2))
	dot := fixed.Point26_6{
		X: fixed.I(cellPad) - bounds.Min.X,
		Y: fixed.I(cellPad) - bounds.Min.Y,
	}
	d := &font.Drawer{Dst: img, Src: image.Opaque, Face: face, Dot: dot}
	d.DrawString(string(r))
	return Glyph{Advance: adv}, img, true
}

func loadFace() (font.Face, string) {
	if path := os.Getenv("SEVG_FONT"); path != "" {
		if face, err := loadFontFile(path); err == nil {
			return face, path
		} else {
			fmt.Printf("UI font %s: %v\n", path, err)
		}
	}
	candidates := []string{
		"C:\\Windows\\Fonts\\msyh.ttc",
		"C:\\Windows\\Fonts\\msyh.ttf",
		"C:\\Windows\\Fonts\\Deng.ttf",
		"C:\\Windows\\Fonts\\simhei.ttf",
		"C:\\Windows\\Fonts\\simsun.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
		"/System/Library/Fonts/PingFang.ttc",
		"/System/Library/Fonts/STHeiti Light.ttc",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		face, err := loadFontFile(path)
		if err == nil {
			return face, path
		}
		fmt.Printf("UI font %s: %v\n", path, err)
	}
	fmt.Println("UI font: fallback bitmap (install a CJK font for Chinese glyphs)")
	return basicfont.Face7x13, "basicfont"
}

func loadFontFile(path string) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	collection, err := opentype.ParseCollection(data)
	if err != nil {
		return nil, err
	}
	f, err := collection.Font(0)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size: fontSize, DPI: 72, Hinting: font.HintingFull,
	})
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
	W, H       int
	Verts      []Vertex
	White      [2]float32
	Glyphs     map[rune]Glyph
	Icons      map[core.ItemID]Glyph
	IconSides  map[core.ItemID][]IconSide
	LineHeight float32
}

func NewCanvas(w, h int, atlas FontAtlas) *Canvas {
	return &Canvas{W: w, H: h, White: atlas.WhiteUV, Glyphs: atlas.Glyphs, Icons: atlas.Icons, IconSides: atlas.IconSides, LineHeight: atlas.LineHeight}
}

func (c *Canvas) Resize(w, h int) { c.W, c.H = w, h }
func (c *Canvas) Reset()          { c.Verts = c.Verts[:0] }

func (c *Canvas) quad(x, y, w, h float32, u0, v0, u1, v1 float32, col [4]float32) {
	if w <= 0 || h <= 0 || col[3] <= 0 {
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

// Icon draws an item preview from the UI atlas (font glyphs + item icons).
func (c *Canvas) Icon(item core.ItemID, x, y, w, h float32, alpha float32) {
	g, ok := c.Icons[item]
	if !ok || alpha <= 0 {
		return
	}
	c.quad(x, y, w, h, g.U0, g.V0, g.U1, g.V1, [4]float32{1, 1, 1, alpha})
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

func (c *Canvas) glyph(r rune) Glyph {
	if g, ok := c.Glyphs[r]; ok {
		return g
	}
	return c.Glyphs['?']
}

func (c *Canvas) Text(x, y float32, s string, scale float32, col [4]float32) float32 {
	cx, cy := x, y
	startX := x
	for _, r := range s {
		if r == 10 {
			cx = startX
			cy += c.LineHeight * scale
			continue
		}
		g := c.glyph(r)
		if g.W > 0 && g.H > 0 {
			c.quad(cx+g.OffsetX*scale, cy+g.OffsetY*scale, g.W*scale, g.H*scale, g.U0, g.V0, g.U1, g.V1, col)
		}
		cx += g.Advance * scale
	}
	return cx
}

func (c *Canvas) TextShadow(x, y float32, s string, scale float32, col [4]float32) {
	o := scale * 1.2
	c.Text(x+o, y+o, s, scale, [4]float32{0, 0, 0, col[3] * 0.75})
	c.Text(x, y, s, scale, col)
}

// TextInkBounds reports the vertical ink bounds of a single-line string
// relative to the text origin. It is used to center actual glyphs rather than
// the font line box, which includes uneven ascent/descent padding.
func (c *Canvas) TextInkBounds(s string, scale float32) (float32, float32) {
	minY, maxY := float32(1e9), float32(-1e9)
	for _, r := range s {
		g := c.glyph(r)
		if g.W <= 0 || g.H <= 0 {
			continue
		}
		top := g.OffsetY * scale
		bottom := top + g.H*scale
		if top < minY {
			minY = top
		}
		if bottom > maxY {
			maxY = bottom
		}
	}
	if minY > maxY {
		return 0, c.LineHeight * scale
	}
	return minY, maxY
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
		line += c.glyph(r).Advance * scale
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
