package entity

import (
	"math"
	"math/rand"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/world"
)

type Kind uint8

const (
	Pig Kind = iota
	Sheep
)

func (k Kind) Name() string {
	if k == Sheep {
		return "Sheep"
	}
	return "Pig"
}

type Box struct {
	Center mathx.Vec3
	Size   mathx.Vec3
	Color  [4]float32
}

type Mob struct {
	ID          int
	Kind        Kind
	Body        physics.Body
	Health      float32
	MaxHealth   float32
	Yaw         float32
	Dead        bool
	HitFlash    float32
	WanderTimer float32
	WanderYaw   float32
	JumpTimer   float32
	OnGround    bool
}

func (m *Mob) Forward() mathx.Vec3 {
	return mathx.Vec3{X: float32(math.Sin(float64(m.Yaw))), Z: -float32(math.Cos(float64(m.Yaw)))}
}

func (m *Mob) Right() mathx.Vec3 {
	return mathx.Vec3{X: float32(math.Cos(float64(m.Yaw))), Z: float32(math.Sin(float64(m.Yaw)))}
}

func (m *Mob) Bounds() (mathx.Vec3, mathx.Vec3) { return m.Body.Min(), m.Body.Max() }

func (m *Mob) Boxes() []Box {
	p := m.Body.Pos
	f, r := m.Forward(), m.Right()
	dark := [4]float32{0.18, 0.14, 0.13, 1}
	if m.Kind == Pig {
		body := [4]float32{0.94, 0.52, 0.56, 1}
		snout := [4]float32{0.86, 0.38, 0.44, 1}
		head := p.Add(f.Mul(0.52)).Add(mathx.Vec3{Y: 0.72})
		boxes := []Box{
			{Center: p.Add(mathx.Vec3{Y: 0.55}), Size: mathx.Vec3{X: 0.9, Y: 0.55, Z: 0.66}, Color: body},
			{Center: head, Size: mathx.Vec3{X: 0.5, Y: 0.48, Z: 0.5}, Color: body},
			{Center: head.Add(f.Mul(0.3)).Add(mathx.Vec3{Y: -0.04}), Size: mathx.Vec3{X: 0.22, Y: 0.18, Z: 0.28}, Color: snout},
		}
		for _, sx := range []float32{-1, 1} {
			for _, sz := range []float32{-1, 1} {
				c := p.Add(r.Mul(sx * 0.26)).Add(f.Mul(sz * 0.21)).Add(mathx.Vec3{Y: 0.18})
				boxes = append(boxes, Box{Center: c, Size: mathx.Vec3{X: 0.18, Y: 0.38, Z: 0.18}, Color: dark})
			}
		}
		return boxes
	}
	body := [4]float32{0.92, 0.92, 0.88, 1}
	face := [4]float32{0.78, 0.74, 0.67, 1}
	head := p.Add(f.Mul(0.52)).Add(mathx.Vec3{Y: 0.74})
	boxes := []Box{
		{Center: p.Add(mathx.Vec3{Y: 0.57}), Size: mathx.Vec3{X: 0.92, Y: 0.62, Z: 0.68}, Color: body},
		{Center: head, Size: mathx.Vec3{X: 0.44, Y: 0.46, Z: 0.44}, Color: face},
	}
	for _, sx := range []float32{-1, 1} {
		for _, sz := range []float32{-1, 1} {
			c := p.Add(r.Mul(sx * 0.26)).Add(f.Mul(sz * 0.22)).Add(mathx.Vec3{Y: 0.17})
			boxes = append(boxes, Box{Center: c, Size: mathx.Vec3{X: 0.16, Y: 0.36, Z: 0.16}, Color: [4]float32{0.55, 0.53, 0.5, 1}})
		}
	}
	return boxes
}

type ItemDrop struct {
	ID    int
	Item  core.ItemID
	Count int
	Body  physics.Body
	Age   float32
	Bob   float32
	Dead  bool
}

func (d *ItemDrop) Boxes() []Box {
	d.Bob += 0 // animation is driven by Age in the caller
	c := d.Body.Pos.Add(mathx.Vec3{Y: 0.18})
	col := [4]float32{0.8, 0.8, 0.8, 1}
	if d.Item.Valid() {
		c3 := d.Item.Def().Color
		col = [4]float32{c3[0], c3[1], c3[2], 1}
	}
	return []Box{{Center: c, Size: mathx.Vec3{X: 0.28, Y: 0.28, Z: 0.28}, Color: col}}
}

type Manager struct {
	Mobs       []*Mob
	Items      []*ItemDrop
	rng        *rand.Rand
	nextID     int
	spawnTimer float32
}

func NewManager(seed int64) *Manager {
	return &Manager{rng: rand.New(rand.NewSource(seed + 909))}
}

func (m *Manager) SpawnInitial(w *world.World, center mathx.Vec3) {
	for i := 0; i < 7; i++ {
		m.trySpawn(w, center, 8, 30)
	}
}

// SpawnAt creates a mob at an exact position (used by tests and the world
// population code).
func (m *Manager) SpawnAt(kind Kind, pos mathx.Vec3) *Mob {
	mob := &Mob{
		ID:        m.nextID,
		Kind:      kind,
		Health:    mobHealth(kind),
		MaxHealth: mobHealth(kind),
		Yaw:       m.rng.Float32() * mathx.Radians(360),
		WanderYaw: m.rng.Float32() * mathx.Radians(360),
	}
	m.nextID++
	mob.Body = physics.Body{Pos: pos, Width: 0.7, Height: 0.95}
	m.Mobs = append(m.Mobs, mob)
	return mob
}

func (m *Manager) trySpawn(w *world.World, center mathx.Vec3, minR, maxR float32) bool {
	for attempt := 0; attempt < 14; attempt++ {
		ang := m.rng.Float64() * math.Pi * 2
		dist := minR + m.rng.Float32()*(maxR-minR)
		x := int(math.Floor(float64(center.X + float32(math.Cos(ang))*dist)))
		z := int(math.Floor(float64(center.Z + float32(math.Sin(ang))*dist)))
		if !w.HasChunk(world.ChunkOf(x, z)) {
			continue
		}
		h := w.SurfaceHeight(x, z)
		above := w.GetBlock(x, h+1, z)
		above2 := w.GetBlock(x, h+2, z)
		if above != core.Air || above2 != core.Air {
			continue
		}
		if !w.GetBlock(x, h, z).Solid() {
			continue
		}
		kind := Pig
		if m.rng.Intn(2) == 0 {
			kind = Sheep
		}
		mob := &Mob{
			ID:        m.nextID,
			Kind:      kind,
			Health:    mobHealth(kind),
			MaxHealth: mobHealth(kind),
			Yaw:       m.rng.Float32() * mathx.Radians(360),
			WanderYaw: m.rng.Float32() * mathx.Radians(360),
		}
		m.nextID++
		mob.Body = physics.Body{
			Pos:    mathx.Vec3{X: float32(x) + 0.5, Y: float32(h) + 1.02, Z: float32(z) + 0.5},
			Width:  0.7,
			Height: 0.95,
		}
		m.Mobs = append(m.Mobs, mob)
		_ = minR
		return true
	}
	return false
}

func mobHealth(k Kind) float32 {
	if k == Sheep {
		return 8
	}
	return 10
}

func (m *Manager) Update(dt float32, w *world.World, playerPos mathx.Vec3, inv *core.Inventory) {
	if dt > 0.05 {
		dt = 0.05
	}
	for _, mob := range m.Mobs {
		if mob.Dead {
			continue
		}
		if mob.HitFlash > 0 {
			mob.HitFlash -= dt
		}
		mob.WanderTimer -= dt
		if mob.WanderTimer <= 0 {
			mob.WanderTimer = 2 + m.rng.Float32()*4
			if m.rng.Float32() < 0.3 {
				mob.WanderYaw = mob.Yaw + mathx.Radians(180)
			} else {
				mob.WanderYaw = m.rng.Float32() * mathx.Radians(360)
			}
		}
		mob.Yaw += angleDelta(mob.Yaw, mob.WanderYaw) * mathx.Clamp(dt*5, 0, 1)
		fwd := mob.Forward()
		walk := float32(1.6)
		if mob.Body.Pos.Sub(playerPos).Length() > 48 {
			walk = 0
		}
		mob.Body.Vel.X += (fwd.X*walk - mob.Body.Vel.X) * mathx.Clamp(dt*8, 0, 1)
		mob.Body.Vel.Z += (fwd.Z*walk - mob.Body.Vel.Z) * mathx.Clamp(dt*8, 0, 1)

		inWater := w.GetBlock(int(math.Floor(float64(mob.Body.Pos.X))), int(math.Floor(float64(mob.Body.Pos.Y))), int(math.Floor(float64(mob.Body.Pos.Z)))) == core.Water
		if inWater {
			mob.Body.Vel.Y -= 5 * dt
			mob.Body.Vel.Y *= 0.92
			mob.Body.Vel.Y += 2.8 * dt
		} else {
			mob.Body.Vel.Y -= 25 * dt
			if mob.Body.Vel.Y < -40 {
				mob.Body.Vel.Y = -40
			}
		}
		mob.JumpTimer -= dt
		hx, _, hz := mob.Body.Move(w, mob.Body.Vel.Mul(dt))
		if (hx || hz) && mob.Body.Grounded(w) && mob.JumpTimer <= 0 {
			mob.Body.Vel.Y = 7.6
			mob.JumpTimer = 0.8
		}
		if mob.Body.Pos.Y < -5 {
			mob.Dead = true
		}
	}

	// Collect pickups and expire old drops.
	for _, d := range m.Items {
		if d.Dead {
			continue
		}
		d.Age += dt
		d.Body.Vel.Y -= 18 * dt
		if d.Body.Vel.Y < -30 {
			d.Body.Vel.Y = -30
		}
		d.Body.Move(w, d.Body.Vel.Mul(dt))
		if d.Body.OnGround {
			d.Body.Vel.X *= 0.7
			d.Body.Vel.Z *= 0.7
		}
		if d.Body.Pos.Y > 300 {
			d.Dead = true
		}
		if d.Age > 0.45 {
			delta := d.Body.Pos.Add(mathx.Vec3{Y: 0.2}).Sub(playerPos)
			if delta.Length() < 2.2 {
				left := inv.Add(d.Item, d.Count)
				if left <= 0 {
					d.Dead = true
				} else {
					d.Count = left
				}
			}
		}
		if d.Age > 300 {
			d.Dead = true
		}
	}

	// Compact slices and spawn new animals while the herd is small.
	live := m.Mobs[:0]
	for _, mob := range m.Mobs {
		if !mob.Dead {
			live = append(live, mob)
		}
	}
	m.Mobs = live
	li := m.Items[:0]
	for _, d := range m.Items {
		if !d.Dead {
			li = append(li, d)
		}
	}
	m.Items = li
	m.spawnTimer -= dt
	if m.spawnTimer <= 0 {
		m.spawnTimer = 6
		if len(m.Mobs) < 9 {
			m.trySpawn(w, playerPos, 22, 44)
		}
	}
}

func angleDelta(a, b float32) float32 {
	d := b - a
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

func (m *Manager) Drop(item core.ItemID, count int, pos mathx.Vec3, vel mathx.Vec3) {
	if item == core.ItemNone || count <= 0 {
		return
	}
	d := &ItemDrop{ID: m.nextID, Item: item, Count: count, Body: physics.Body{Pos: pos, Vel: vel, Width: 0.25, Height: 0.25}}
	m.nextID++
	m.Items = append(m.Items, d)
}

// PickMob finds the nearest mob along a ray without damaging it.
func (m *Manager) PickMob(origin, dir mathx.Vec3, maxDist float32) (*Mob, float32, bool) {
	dir = dir.Normalize()
	bestDist := maxDist
	var best *Mob
	for _, mob := range m.Mobs {
		if mob.Dead {
			continue
		}
		min, max := mob.Bounds()
		if d, ok := world.RayAABB(origin, dir, min, max); ok && d < bestDist {
			bestDist = d
			best = mob
		}
	}
	if best == nil {
		return nil, 0, false
	}
	return best, bestDist, true
}

// DamageMob applies damage, knockback and death drops.
func (m *Manager) DamageMob(mob *Mob, damage float32, playerPos mathx.Vec3) {
	if mob == nil || mob.Dead {
		return
	}
	mob.Health -= damage
	mob.HitFlash = 0.25
	kb := mob.Body.Pos.Sub(playerPos).Normalize().Mul(6)
	mob.Body.Vel.X = kb.X
	mob.Body.Vel.Z = kb.Z
	mob.Body.Vel.Y = 4
	if mob.Health <= 0 {
		mob.Dead = true
		drop := core.ItemWool
		n := 2 + m.rng.Intn(2)
		if mob.Kind == Pig {
			drop = core.ItemApple
			n = 1 + m.rng.Intn(2)
		}
		m.Drop(drop, n, mob.Body.Pos.Add(mathx.Vec3{Y: 0.5}), mathx.Vec3{X: m.rng.Float32()*2 - 1, Y: 4, Z: m.rng.Float32()*2 - 1})
	}
}

func (m *Manager) MobCount() int  { return len(m.Mobs) }
func (m *Manager) ItemCount() int { return len(m.Items) }
