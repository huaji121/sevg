package world

import "sevg/internal/core"

const (
	padSizeX = ChunkSize + 2
	padArea  = padSizeX * padSizeX * ChunkHeight
)

type Vertex struct {
	X, Y, Z float32
	U, V    float32
	Light   float32
	Sky     float32
	Block   float32
}

type MeshData struct {
	Vertices []Vertex
	Indices  []uint32
}

type ChunkMesh struct {
	Opaque      MeshData
	Transparent MeshData
}

type faceDef struct {
	normal  [3]int
	corners [4][3]float32
	uv      [4][2]float32
	axisA   int
	axisB   int
	light   float32
}

var faces = [6]faceDef{
	{ // +X
		normal:  [3]int{1, 0, 0},
		corners: [4][3]float32{{1, 0, 1}, {1, 0, 0}, {1, 1, 0}, {1, 1, 1}},
		uv:      [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		axisA:   1, axisB: 2, light: 0.76,
	},
	{ // -X
		normal:  [3]int{-1, 0, 0},
		corners: [4][3]float32{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}, {0, 1, 0}},
		uv:      [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		axisA:   1, axisB: 2, light: 0.72,
	},
	{ // +Y
		normal:  [3]int{0, 1, 0},
		corners: [4][3]float32{{0, 1, 1}, {1, 1, 1}, {1, 1, 0}, {0, 1, 0}},
		uv:      [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		axisA:   0, axisB: 2, light: 1.0,
	},
	{ // -Y
		normal:  [3]int{0, -1, 0},
		corners: [4][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		uv:      [4][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
		axisA:   0, axisB: 2, light: 0.42,
	},
	{ // +Z
		normal:  [3]int{0, 0, 1},
		corners: [4][3]float32{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
		uv:      [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		axisA:   0, axisB: 1, light: 0.86,
	},
	{ // -Z
		normal:  [3]int{0, 0, -1},
		corners: [4][3]float32{{1, 0, 0}, {0, 0, 0}, {0, 1, 0}, {1, 1, 0}},
		uv:      [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}},
		axisA:   0, axisB: 1, light: 0.82,
	},
}

func TileUV(tile uint8) (u0, v0, u1, v1 float32) {
	return tileUV(tile)
}

func tileUV(tile uint8) (u0, v0, u1, v1 float32) {
	const half = 0.01
	cx := int(tile) % core.AtlasCols
	cy := int(tile) / core.AtlasCols
	px := float32(cx*core.AtlasCell + core.AtlasPad)
	py := float32(cy*core.AtlasCell + core.AtlasPad)
	w := float32(core.AtlasW)
	h := float32(core.AtlasH)
	u0 = (px + half) / w
	v0 = (py + half) / h
	u1 = (px + core.AtlasTilePx - half) / w
	v1 = (py + core.AtlasTilePx - half) / h
	return
}

// Mesh builds CPU geometry for one chunk. The padded 18x18 column snapshot
// makes neighbour and ambient-occlusion lookups cheap and consistent.
func (w *World) Mesh(pos ChunkPos) (ChunkMesh, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	c := w.chunks[pos.Key()]
	if c == nil {
		return ChunkMesh{}, false
	}
	pad := make([]core.BlockID, padArea)
	padSky := make([]uint8, padArea)
	padBlock := make([]uint8, padArea)
	for pz := 0; pz < padSizeX; pz++ {
		wz := pos.Z*ChunkSize + pz - 1
		for px := 0; px < padSizeX; px++ {
			wx := pos.X*ChunkSize + px - 1
			cp := ChunkOf(wx, wz)
			nc := w.chunks[cp.Key()]
			if nc == nil {
				continue
			}
			lx, lz := Mod(wx, ChunkSize), Mod(wz, ChunkSize)
			base := (px + pz*padSizeX) * ChunkHeight
			src := lx + lz*ChunkSize
			for y := 0; y < ChunkHeight; y++ {
				si := src + y*ChunkSize*ChunkSize
				pad[base+y] = nc.Blocks[si]
				padSky[base+y] = nc.SkyLight[si]
				padBlock[base+y] = nc.BlockLight[si]
			}
		}
	}
	at := func(x, y, z int) core.BlockID {
		if y < 0 || y >= ChunkHeight {
			return core.Air
		}
		x++
		z++
		if x < 0 || x >= padSizeX || z < 0 || z >= padSizeX {
			return core.Air
		}
		return pad[(x+z*padSizeX)*ChunkHeight+y]
	}
	solidAO := func(x, y, z int) bool { return at(x, y, z).Opaque() }
	atSky := func(x, y, z int) uint8 {
		if y < 0 || y >= ChunkHeight {
			return 0
		}
		x++
		z++
		if x < 0 || x >= padSizeX || z < 0 || z >= padSizeX {
			return 15
		}
		return padSky[(x+z*padSizeX)*ChunkHeight+y]
	}
	atBlock := func(x, y, z int) uint8 {
		if y < 0 || y >= ChunkHeight {
			return 0
		}
		x++
		z++
		if x < 0 || x >= padSizeX || z < 0 || z >= padSizeX {
			return 0
		}
		return padBlock[(x+z*padSizeX)*ChunkHeight+y]
	}

	var out ChunkMesh
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			for y := 0; y < ChunkHeight; y++ {
				b := at(x, y, z)
				if b == core.Air {
					continue
				}
				if b.Cross() {
					addCross(&out.Opaque, x, y, z, pos.X*ChunkSize, pos.Z*ChunkSize, b, atSky, atBlock)
					continue
				}
				transparent := b == core.Water || b == core.Glass || b == core.Ice
				target := &out.Opaque
				if transparent {
					target = &out.Transparent
				}
				for fi := 0; fi < 6; fi++ {
					f := &faces[fi]
					nx, ny, nz := x+f.normal[0], y+f.normal[1], z+f.normal[2]
					nb := at(nx, ny, nz)
					if !faceVisible(b, nb) {
						continue
					}
					addFace(target, at, solidAO, atSky, atBlock, x, y, z, pos.X*ChunkSize, pos.Z*ChunkSize, b, fi)
				}
			}
		}
	}
	return out, true
}

func faceVisible(cur, nb core.BlockID) bool {
	if nb == core.Air {
		return true
	}
	if nb.Opaque() {
		return false
	}
	if nb == cur {
		return false
	}
	if cur == core.Water {
		return !nb.Opaque()
	}
	if cur == core.Glass || cur == core.Ice {
		return nb != cur && !nb.Opaque()
	}
	return true
}

func addFace(m *MeshData, at func(int, int, int) core.BlockID, solid func(int, int, int) bool, skyAt, blockAt func(int, int, int) uint8, bx, by, bz, ox, oz int, b core.BlockID, fi int) {
	f := &faces[fi]
	u0, v0, u1, v1 := tileUV(b.Tile(fi))
	base := uint32(len(m.Vertices))
	for i := 0; i < 4; i++ {
		c := f.corners[i]
		var ao int32 = 3
		oa := int(c[f.axisA]*2) - 1
		ob := int(c[f.axisB]*2) - 1
		sx, sy, sz := bx+f.normal[0], by+f.normal[1], bz+f.normal[2]
		p1 := [3]int{sx, sy, sz}
		p1[f.axisA] += oa
		p2 := [3]int{sx, sy, sz}
		p2[f.axisB] += ob
		pc := [3]int{sx, sy, sz}
		pc[f.axisA] += oa
		pc[f.axisB] += ob
		s1 := solid(p1[0], p1[1], p1[2])
		s2 := solid(p2[0], p2[1], p2[2])
		cc := solid(pc[0], pc[1], pc[2])
		if s1 && s2 {
			ao = 0
		} else {
			if s1 {
				ao--
			}
			if s2 {
				ao--
			}
			if cc {
				ao--
			}
		}
		light := f.light * (0.58 + float32(ao)*0.14)
		sky := float32(skyAt(sx, sy, sz)) / 15
		blk := float32(blockAt(sx, sy, sz)) / 15
		uu := u0 + (u1-u0)*f.uv[i][0]
		vv := v0 + (v1-v0)*f.uv[i][1]
		m.Vertices = append(m.Vertices, Vertex{
			X: float32(ox+bx) + c[0], Y: float32(by) + c[1], Z: float32(oz+bz) + c[2],
			U: uu, V: vv, Light: light, Sky: sky, Block: blk,
		})
	}
	m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
}

func addCross(m *MeshData, bx, by, bz, ox, oz int, b core.BlockID, skyAt, blockAt func(int, int, int) uint8) {
	u0, v0, u1, v1 := tileUV(b.Tile(2))
	lo, hi, top := float32(0.12), float32(0.88), float32(1)
	if b == core.Torch {
		lo, hi, top = 0.40, 0.60, 0.70
	}
	quads := [2][4][3]float32{
		{{lo, 0, lo}, {hi, 0, hi}, {hi, top, hi}, {lo, top, lo}},
		{{lo, 0, hi}, {hi, 0, lo}, {hi, top, lo}, {lo, top, hi}},
	}
	uvs := [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	sky := float32(skyAt(bx, by, bz)) / 15
	blk := float32(blockAt(bx, by, bz)) / 15
	for _, q := range quads {
		base := uint32(len(m.Vertices))
		for i := 0; i < 4; i++ {
			m.Vertices = append(m.Vertices, Vertex{
				X: float32(ox+bx) + q[i][0], Y: float32(by) + q[i][1], Z: float32(oz+bz) + q[i][2],
				U: u0 + (u1-u0)*uvs[i][0], V: v0 + (v1-v0)*uvs[i][1],
				Light: 0.92, Sky: sky, Block: blk,
			})
		}
		m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}
