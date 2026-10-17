package game

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/player"
	"sevg/internal/world"
)

const saveVersion = 2

type playerSave struct {
	X, Y, Z    float32
	Yaw, Pitch float32
	Health     float32
	Flying     bool
}

type SaveGame struct {
	Version   int
	Seed      int64
	Language  string
	Mode      int
	Time      float64
	Player    playerSave
	Inventory core.Inventory
	Edits     []world.EditRecord
}

func (g *Game) saveExists() bool {
	_, err := os.Stat(g.saveAt)
	return err == nil
}

func loadSave(path string) (*SaveGame, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var s SaveGame
	if err := gob.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	if s.Version != saveVersion {
		return nil, fmt.Errorf("save version %d is not supported", s.Version)
	}
	return &s, nil
}

// startWorld creates a fresh world or resumes a saved one.
func (g *Game) startWorld(load bool) error {
	var sv *SaveGame
	var err error
	if load {
		sv, err = loadSave(g.saveAt)
		if err != nil {
			g.notify(i18n.T("toast.could_not_load_save", err.Error()), 4)
			return err
		}
	}
	g.stopWorkers()
	g.gfx.ReleaseAllChunks()
	g.meshed = make(map[[2]int]bool)
	g.pending = make(map[[2]int]bool)

	seed := g.seed
	inv := core.Inventory{}
	inv.FillStarter()
	timeOfDay := 0.31
	var p *player.Player
	var w *world.World

	if sv != nil {
		seed = sv.Seed
		g.seed = seed
		if sv.Language != "" {
			i18n.Set(sv.Language)
		}
		g.mode = GameMode(sv.Mode)
		if g.mode != ModeCreative {
			g.mode = ModeSurvival
		}
		inv = sv.Inventory
		inv.Normalize()
		timeOfDay = sv.Time
		w = world.New(seed)
		w.ImportEdits(sv.Edits)
		spawn := mathx.Vec3{X: sv.Player.X, Y: sv.Player.Y, Z: sv.Player.Z}
		p = player.New(spawn)
		p.Yaw = sv.Player.Yaw
		p.Pitch = sv.Player.Pitch
		p.Health = sv.Player.Health
		if p.Health <= 0 {
			p.Health = p.MaxHP
		}
		p.Flying = sv.Player.Flying
	} else {
		g.mode = ModeSurvival
		w = world.New(seed)
		spawn := w.FindSpawn()
		p = player.New(spawn)
	}
	if p.Body.Pos.Y < 1 {
		p.Body.Pos.Y = 64
	}
	p.Creative = g.mode == ModeCreative
	if p.Creative {
		p.Health = p.MaxHP
	} else {
		p.Flying = false
	}

	center := world.ChunkOf(int(p.Body.Pos.X), int(p.Body.Pos.Z))
	// Generate a safe initial island synchronously, then stream the rest.
	for dz := -2; dz <= 2; dz++ {
		for dx := -2; dx <= 2; dx++ {
			w.EnsureChunk(world.ChunkPos{X: center.X + dx, Z: center.Z + dz})
		}
	}
	if sv == nil {
		p.Body.Pos = w.SafeSpawn(p.Body.Pos, 24)
		w.ClearSpawnArea(p.Body.Pos, 7, 10)
		p.Spawning = p.Body.Pos
		p.FallingReset()
	} else if physics.Collides(w, p.Body.Pos, p.Body.Width, p.Body.Height) ||
		!w.GetBlock(int(math.Floor(float64(p.Body.Pos.X))), int(math.Floor(float64(p.Body.Pos.Y))), int(math.Floor(float64(p.Body.Pos.Z)))).IsAir() {
		p.Body.Pos = w.SafeSpawn(p.Body.Pos, 24)
		p.Spawning = p.Body.Pos
		p.FallingReset()
	}
	if p.Body.Pos.Y < 0 {
		p.Body.Pos = w.SafeSpawn(w.FindSpawn(), 24)
	}

	// Upload the initial island so the first frame already has terrain.
	for dz := -2; dz <= 2; dz++ {
		for dx := -2; dx <= 2; dx++ {
			cp := world.ChunkPos{X: center.X + dx, Z: center.Z + dz}
			if mesh, ok := w.Mesh(cp); ok {
				if err := g.gfx.UploadChunk(cp, mesh); err != nil {
					fmt.Println("initial chunk upload:", err)
				}
				w.ClearDirty(cp)
				g.meshed[cp.Key()] = true
			}
		}
	}

	g.world = w
	g.player = p
	g.inv = inv
	g.mobs = entity.NewManager(seed)
	g.time = timeOfDay
	g.helpTime = 18
	g.toastTime = 0
	g.saveTimer = 0
	g.breakProgress = 0
	g.breakActive = false
	g.mobs.SpawnInitial(w, p.Body.Pos)
	g.startWorkers()
	fmt.Printf("WORLD seed=%d spawn=%.2f,%.2f,%.2f loaded=%d meshed=%d mobs=%d\n", seed, p.Body.Pos.X, p.Body.Pos.Y, p.Body.Pos.Z, w.Loaded(), len(g.meshed), g.mobs.MobCount())
	_ = g.Win.SetTitle(i18n.T("window.title"))
	g.addChat(i18n.T("chat.welcome"))
	g.setState(StatePlaying)
	g.requestChunks(center)
	if sv != nil {
		g.notify(i18n.T("toast.loaded"), 2)
	} else {
		g.notify(i18n.T("toast.new_world", seed), 3)
	}
	return nil
}

func (g *Game) startWorkers() {
	if g.world == nil || g.genJobs != nil {
		return
	}
	g.genDone = make(chan struct{})
	g.genJobs = make(chan world.ChunkPos, 128)
	g.genResult = make(chan genResult, 128)
	w := g.world
	for i := 0; i < 2; i++ {
		g.genWG.Add(1)
		go func() {
			defer g.genWG.Done()
			for {
				select {
				case cp, ok := <-g.genJobs:
					if !ok {
						return
					}
					_, created := w.EnsureChunk(cp)
					select {
					case g.genResult <- genResult{pos: cp, ok: created}:
					case <-g.genDone:
						return
					}
				case <-g.genDone:
					return
				}
			}
		}()
	}
}

func (g *Game) stopWorkers() {
	if g.genJobs == nil {
		return
	}
	close(g.genDone)
	close(g.genJobs)
	g.genWG.Wait()
	g.genJobs = nil
	g.genResult = nil
	g.genDone = nil
	g.pending = make(map[[2]int]bool)
}

func (g *Game) updateStreaming(dt float32) {
	if g.world == nil {
		return
	}
	for {
		select {
		case res := <-g.genResult:
			delete(g.pending, res.pos.Key())
			if res.ok {
				g.world.MarkNeighborsDirty(res.pos)
			}
		default:
			goto drained
		}
	}
drained:
	center := world.ChunkPos{}
	if g.player != nil {
		center = world.ChunkOf(int(g.player.Body.Pos.X), int(g.player.Body.Pos.Z))
	}
	g.streamT -= dt
	if g.streamT <= 0 {
		g.streamT = 0.2
		g.requestChunks(center)
	}

	// Remesh one dirty chunk per frame, closest first. Uploading a chunk
	// creates GPU buffers and submits a copy pass, so doing several in one
	// frame produces visible hitches while streaming.
	dirty := g.world.DirtyChunks(center)
	budget := 1
	for _, cp := range dirty {
		if cheb(cp, center) > g.opts.RenderDist+1 {
			continue
		}
		mesh, ok := g.world.Mesh(cp)
		if !ok {
			continue
		}
		if err := g.gfx.UploadChunk(cp, mesh); err != nil {
			fmt.Println("chunk upload:", err)
			continue
		}
		g.world.ClearDirty(cp)
		g.meshed[cp.Key()] = true
		budget--
		if budget <= 0 {
			break
		}
	}

	g.unloadT -= dt
	if g.unloadT <= 0 {
		g.unloadT = 2
		for _, cp := range g.world.ChunkPositions() {
			if cheb(cp, center) > g.opts.RenderDist+2 {
				g.gfx.ReleaseChunk(cp)
				g.world.Unload(cp)
				delete(g.meshed, cp.Key())
			}
		}
	}
}

func (g *Game) requestChunks(center world.ChunkPos) {
	if g.world == nil || g.genJobs == nil {
		return
	}
	queued := len(g.pending)
	if queued >= 8 {
		return
	}
	radius := g.opts.RenderDist
	for r := 0; r <= radius && queued < 8; r++ {
		for dz := -r; dz <= r && queued < 8; dz++ {
			for dx := -r; dx <= r && queued < 8; dx++ {
				if maxAbs(dx) != r && maxAbs(dz) != r {
					continue
				}
				cp := world.ChunkPos{X: center.X + dx, Z: center.Z + dz}
				key := cp.Key()
				if g.pending[key] || g.world.HasChunk(cp) {
					continue
				}
				select {
				case g.genJobs <- cp:
					g.pending[key] = true
					queued++
				default:
					return
				}
			}
		}
	}
}

func (g *Game) saveGame() error {
	if g.world == nil || g.player == nil {
		return nil
	}
	if !g.heldStack.Empty() {
		left := g.inv.Add(g.heldStack.Item, g.heldStack.Count)
		if left > 0 && g.mobs != nil {
			g.mobs.Drop(g.heldStack.Item, left, g.player.Body.Pos, mathx.Vec3{Y: 3})
		}
		g.heldStack = core.Stack{}
	}
	sv := SaveGame{
		Version:  saveVersion,
		Seed:     g.world.Seed,
		Language: i18n.Current(),
		Mode:     int(g.mode),
		Time:     g.time,
		Player: playerSave{
			X: g.player.Body.Pos.X, Y: g.player.Body.Pos.Y, Z: g.player.Body.Pos.Z,
			Yaw: g.player.Yaw, Pitch: g.player.Pitch, Health: g.player.Health,
			Flying: g.player.Flying,
		},
		Inventory: g.inv,
		Edits:     g.world.ExportEdits(),
	}
	dir := filepath.Dir(g.saveAt)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, ".sevg-save-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	enc := gob.NewEncoder(tmp)
	if err := enc.Encode(&sv); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, g.saveAt)
}

func (g *Game) quitToTitle() {
	_ = g.saveGame()
	g.stopWorkers()
	g.gfx.ReleaseAllChunks()
	g.world = nil
	g.player = nil
	g.mobs = nil
	g.heldStack = core.Stack{}
	g.meshed = make(map[[2]int]bool)
	g.pending = make(map[[2]int]bool)
	g.setState(StateMenu)
}

func (g *Game) loadedTarget() int {
	n := g.opts.RenderDist*2 + 1
	return n * n
}

func cheb(a, b world.ChunkPos) int {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dz := a.Z - b.Z
	if dz < 0 {
		dz = -dz
	}
	if dx > dz {
		return dx
	}
	return dz
}

func maxAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (g *Game) timeString() string {
	hours := int(g.time*24) % 24
	mins := int(g.time*24*60) % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}
