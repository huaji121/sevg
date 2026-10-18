package game

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/audio"
	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
	"sevg/internal/player"
	"sevg/internal/render"
	"sevg/internal/ui"
	"sevg/internal/world"
)

type Options struct {
	Width, Height  int
	Seed           int64
	RenderDist     int
	SavePath       string
	SmokeFrames    int
	Continue       bool
	Language       string
	FPSLimit       int
	FOV            int
	Fullscreen     bool
	SmoothLighting bool
	DynamicLight   bool
	VSync          bool
}

type State int

const (
	StateMenu State = iota
	StatePlaying
	StateInventory
	StatePaused
	StateDead
	StateSettings
)

type GameMode int

const (
	ModeSurvival GameMode = iota
	ModeCreative
)

type genResult struct {
	pos    world.ChunkPos
	blocks []core.BlockID
}

type Game struct {
	opts  Options
	Win   *sdl.Window
	gfx   *render.Renderer
	canv  *ui.Canvas
	audio *audio.Engine

	state          State
	settingsBack   State
	mode           GameMode
	seed           int64
	world          *world.World
	player         *player.Player
	mobs           *entity.Manager
	inv            core.Inventory
	inventoryTab   int
	craftGrid      [9]core.Stack
	craftSize      int
	recipeBookOpen bool
	time           float64
	saveAt         string

	mouseX, mouseY float32
	mouseLeft      bool
	mouseRight     bool
	shiftDown      bool
	heldStack      core.Stack

	breakPos      [3]int
	targetPos     [3]int
	breakBlock    core.BlockID
	breakProgress float32
	breakActive   bool
	particles     []breakParticle
	particleTimer float32
	targetValid   bool
	attackCool    float32
	placeCool     float32
	pickCool      float32
	breakCool     float32
	spaceTapTimer float32
	stepTimer     float32
	fluidTimer    float32

	chatOpen          bool
	chatInput         []rune
	chatCaret         int
	chatComposition   string
	chatSuppressSlash bool
	chatScroll        int
	chatMessages      []chatMessage

	toasts     []toastEntry
	heldLightT float32
	helpTime   float32
	debug      bool
	fullscreen bool

	fps        float64
	frameMs    float64
	fpsFrames  int
	fpsElapsed float64
	frame      int
	quit       bool

	saveTimer float32
	unloadT   float32
	streamT   float32
	lastTime  time.Time

	// Async chunk generation.
	genJobs   chan world.ChunkPos
	genResult chan genResult
	genDone   chan struct{}
	genWG     sync.WaitGroup
	pending   map[[2]int]bool
	meshed    map[[2]int]bool
}

func Run(opts Options) error {
	if opts.Width <= 0 {
		opts.Width = 1280
	}
	if opts.Height <= 0 {
		opts.Height = 720
	}
	if opts.RenderDist < 3 {
		opts.RenderDist = 6
	}
	if opts.SavePath == "" {
		opts.SavePath = "savegame.gob"
	}
	if opts.Language == "" {
		opts.Language = i18n.ZH
	}
	if opts.FOV <= 0 {
		opts.FOV = 72
	}
	if opts.SmokeFrames == 0 {
		loadVideoSettings(&opts)
	}
	normalizeVideoOptions(&opts)
	i18n.Set(opts.Language)
	runes := append(i18n.AllRunes(), i18n.ChatRunes()...)
	fontAtlas := ui.BuildFontAtlas(runes, render.BuildAtlas())

	var unload func()
	if err := sdl.LoadLibrary(sdl.Path()); err != nil {
		lib := binsdl.Load()
		unload = lib.Unload
	}
	defer func() {
		sdl.Quit()
		if unload != nil {
			unload()
		}
	}()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return fmt.Errorf("SDL init: %w", err)
	}
	fmt.Println("SEVG - Simple and Extendable Voxel Game")
	fmt.Println("SDL", sdl.GetVersion().String())
	win, err := sdl.CreateWindow(i18n.T("window.title"), opts.Width, opts.Height, sdl.WINDOW_RESIZABLE)
	if err != nil {
		return fmt.Errorf("create window: %w", err)
	}
	defer win.Destroy()

	gfx, err := render.New(win, opts.Width, opts.Height, fontAtlas)
	if err != nil {
		return err
	}
	if !opts.VSync {
		if err := gfx.SetVSync(false); err != nil {
			fmt.Println("disable vsync:", err)
		}
	}
	g := &Game{
		opts:       opts,
		Win:        win,
		gfx:        gfx,
		canv:       ui.NewCanvas(opts.Width, opts.Height, fontAtlas),
		state:      StateMenu,
		saveAt:     opts.SavePath,
		pending:    make(map[[2]int]bool),
		meshed:     make(map[[2]int]bool),
		lastTime:   time.Now(),
		fullscreen: opts.Fullscreen,
	}
	g.seed = opts.Seed
	if g.opts.Fullscreen {
		_ = win.SetFullscreen(true)
	}
	if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err == nil {
		g.audio = audio.Open()
	}
	defer func() { g.audio.Close() }()
	if g.seed == 0 {
		g.seed = time.Now().UnixNano()
	}
	if opts.SmokeFrames > 0 {
		if err := g.startWorld(opts.Continue); err != nil {
			gfx.Close()
			return err
		}
		g.setState(StatePlaying)
	} else if opts.Continue {
		if err := g.startWorld(true); err != nil {
			g.notify(i18n.T("toast.could_not_load_save", err.Error()), 4)
			g.setState(StateMenu)
		}
	}
	g.run()
	g.shutdown()
	gfx.Close()
	return nil
}

func (g *Game) openInventory() {
	g.craftSize = 2
	g.recipeBookOpen = false
	g.setState(StateInventory)
}

func (g *Game) openCraftingTable() {
	if g.state != StatePlaying {
		return
	}
	g.craftSize = 3
	g.recipeBookOpen = false
	g.inventoryTab = 0
	g.setState(StateInventory)
}

func (g *Game) setState(s State) {
	if g.state == StateInventory && s != StateInventory {
		g.returnHeldStack()
		if g.mode != ModeCreative {
			g.returnCraftGrid()
		}
		g.recipeBookOpen = false
		g.craftSize = 2
	}
	if s != StatePlaying && g.chatOpen {
		g.chatOpen = false
		g.chatInput = g.chatInput[:0]
		g.chatCaret = 0
		g.chatComposition = ""
		g.chatSuppressSlash = false
		g.chatScroll = 0
		if g.Win != nil {
			_ = g.Win.StopTextInput()
		}
	}
	g.state = s
	captured := s == StatePlaying
	if g.Win != nil {
		_ = g.Win.SetRelativeMouseMode(captured)
	}
	if captured {
		_ = sdl.HideCursor()
	} else {
		_ = sdl.ShowCursor()
		g.mouseLeft = false
		g.mouseRight = false
	}
}

func (g *Game) run() {
	for !g.quit {
		frameStart := time.Now()
		now := frameStart
		rawDt := now.Sub(g.lastTime).Seconds()
		g.lastTime = now
		if rawDt <= 0 {
			rawDt = 1.0 / 1000
		}
		g.fpsFrames++
		g.fpsElapsed += rawDt
		if g.fpsElapsed >= 0.5 {
			g.fps = float64(g.fpsFrames) / g.fpsElapsed
			g.frameMs = 1000 * g.fpsElapsed / float64(g.fpsFrames)
			g.fpsFrames, g.fpsElapsed = 0, 0
		}
		dt := rawDt
		if dt > 0.1 {
			dt = 0.1
		}
		g.frame++

		g.poll()
		if g.quit {
			break
		}
		g.refreshMouse()
		g.updateToasts(float32(dt))
		if g.helpTime > 0 {
			g.helpTime -= float32(dt)
		}
		switch g.state {
		case StatePlaying:
			g.updatePlaying(float32(dt))
		case StateInventory:
			// Keep the world, mobs, fluids and item drops running while the
			// inventory is open; only suppress player movement and actions.
			g.updateWorld(float32(dt), false)
		case StateMenu:
			// idle title animation
		}
		g.updateHeldLight(float32(dt))
		g.updateStreaming(float32(dt))
		g.audio.Update()
		g.drawUI()
		g.renderFrame()
		if g.opts.SmokeFrames > 0 && g.frame >= g.opts.SmokeFrames {
			g.quit = true
		}
		if g.opts.FPSLimit > 0 {
			target := time.Second / time.Duration(g.opts.FPSLimit)
			if elapsed := time.Since(frameStart); elapsed < target {
				time.Sleep(target - elapsed)
			}
		}
	}
}

func (g *Game) poll() {
	var e sdl.Event
	for sdl.PollEvent(&e) {
		switch e.Type {
		case sdl.EVENT_QUIT:
			g.quit = true
		case sdl.EVENT_WINDOW_RESIZED, sdl.EVENT_WINDOW_PIXEL_SIZE_CHANGED:
			w, h, err := g.Win.SizeInPixels()
			if err == nil {
				g.gfx.Resize(int(w), int(h))
				g.canv.Resize(int(w), int(h))
			}
		case sdl.EVENT_TEXT_INPUT:
			if g.chatOpen {
				g.chatInsert(e.TextInputEvent().Text)
			}
		case sdl.EVENT_TEXT_EDITING:
			if g.chatOpen {
				g.chatComposition = e.TextEditingEvent().Text
			}
		case sdl.EVENT_KEY_DOWN:
			ke := e.KeyboardEvent()
			if ke.Scancode == sdl.SCANCODE_LSHIFT || ke.Scancode == sdl.SCANCODE_RSHIFT {
				g.shiftDown = true
			}
			if ke.Repeat {
				if g.chatOpen {
					switch ke.Key {
					case sdl.K_LEFT, sdl.K_RIGHT, sdl.K_BACKSPACE, sdl.K_DELETE:
						g.onChatKey(ke.Key)
					}
				}
				continue
			}
			g.onKey(ke.Key)
		case sdl.EVENT_KEY_UP:
			ke := e.KeyboardEvent()
			if ke.Scancode == sdl.SCANCODE_LSHIFT || ke.Scancode == sdl.SCANCODE_RSHIFT {
				g.shiftDown = false
			}
		case sdl.EVENT_MOUSE_MOTION:
			me := e.MouseMotionEvent()
			g.mouseX, g.mouseY = g.toUI(me.X, me.Y)
			if g.state == StatePlaying && !g.chatOpen {
				g.player.Rotate(me.Xrel, me.Yrel, 0.0022)
			}
		case sdl.EVENT_MOUSE_BUTTON_DOWN:
			be := e.MouseButtonEvent()
			g.mouseX, g.mouseY = g.toUI(be.X, be.Y)
			g.onMouseDown(be.Button)
		case sdl.EVENT_MOUSE_BUTTON_UP:
			be := e.MouseButtonEvent()
			g.mouseX, g.mouseY = g.toUI(be.X, be.Y)
			g.onMouseUp(be.Button)
		case sdl.EVENT_MOUSE_WHEEL:
			we := e.MouseWheelEvent()
			if g.chatOpen {
				if we.Y > 0 {
					g.scrollChat(3)
				} else if we.Y < 0 {
					g.scrollChat(-3)
				}
				continue
			}
			if g.state == StatePlaying && !g.chatOpen {
				if we.Y < 0 {
					g.inv.Scroll(1)
				} else if we.Y > 0 {
					g.inv.Scroll(-1)
				}
			}
		}
	}
}

func (g *Game) onKey(k sdl.Keycode) {
	if g.chatOpen {
		g.onChatKey(k)
		return
	}
	if g.state == StatePlaying && k == sdl.K_SLASH {
		g.openChatWith("/")
		return
	}
	if g.state == StatePlaying && (k == sdl.K_T || k == sdl.K_RETURN || k == sdl.K_KP_ENTER) {
		g.openChat()
		return
	}
	switch g.state {
	case StateMenu:
		if k == sdl.K_RETURN || k == sdl.K_KP_ENTER {
			if err := g.startWorld(false); err == nil {
				g.setState(StatePlaying)
			}
		} else if k == sdl.K_ESCAPE {
			g.quit = true
		} else if k == sdl.K_L {
			g.toggleLanguage()
		}
	case StatePlaying:
		switch k {
		case sdl.K_ESCAPE:
			g.setState(StatePaused)
		case sdl.K_E:
			g.openInventory()
		case sdl.K_SPACE:
			if g.mode == ModeCreative {
				if g.spaceTapTimer > 0 {
					g.spaceTapTimer = 0
					g.player.Flying = !g.player.Flying
					g.player.Body.Vel.Y = 0
					g.notify(i18n.T("toast.flying", onOff(g.player.Flying)), 1.6)
				} else {
					g.spaceTapTimer = 0.30
				}
			}
		case sdl.K_F3:
			g.debug = !g.debug
		case sdl.K_H:
			if g.helpTime > 0 {
				g.helpTime = 0
			} else {
				g.helpTime = 10
			}
		case sdl.K_F5:
			if err := g.saveGame(); err != nil {
				g.notify(i18n.T("toast.save_failed", err.Error()), 3)
			} else {
				g.notify(i18n.T("toast.saved"), 1.8)
			}
		case sdl.K_L:
			g.toggleLanguage()
		case sdl.K_F11:
			g.setFullscreen(!g.fullscreen)
		case sdl.K_Q:
			g.dropSelected()
		case sdl.K_1, sdl.K_2, sdl.K_3, sdl.K_4, sdl.K_5, sdl.K_6, sdl.K_7, sdl.K_8, sdl.K_9:
			g.inv.Select(int(k - sdl.K_1))
		}
	case StateInventory:
		if k == sdl.K_Q {
			g.dropFromInventory()
		} else if k == sdl.K_R && g.mode != ModeCreative {
			g.recipeBookOpen = !g.recipeBookOpen
		} else if (k == sdl.K_E || k == sdl.K_ESCAPE) && g.recipeBookOpen {
			g.recipeBookOpen = false
		} else if k >= sdl.K_1 && k <= sdl.K_9 {
			g.inv.Select(int(k - sdl.K_1))
		} else if k == sdl.K_L {
			g.toggleLanguage()
		} else if k == sdl.K_E || k == sdl.K_ESCAPE {
			g.setState(StatePlaying)
		}
	case StateSettings:
		if k == sdl.K_ESCAPE {
			g.setState(g.settingsBack)
		} else if k == sdl.K_F11 {
			g.setFullscreen(!g.fullscreen)
		}
	case StatePaused:
		if k == sdl.K_ESCAPE {
			g.setState(StatePlaying)
		} else if k == sdl.K_L {
			g.toggleLanguage()
		}
	case StateDead:
		if k == sdl.K_R || k == sdl.K_RETURN {
			g.respawn()
		}
	}
}

func (g *Game) onMouseDown(button uint8) {
	if g.chatOpen {
		return
	}
	if g.state != StatePlaying {
		if button == uint8(sdl.BUTTON_LEFT) {
			g.handleUIClick(false)
		} else if button == uint8(sdl.BUTTON_RIGHT) {
			g.handleUIClick(true)
		}
		return
	}
	g.mouseLeft = false
	g.mouseRight = false
	if int(button) == int(sdl.BUTTON_LEFT) {
		g.mouseLeft = true
	} else if int(button) == int(sdl.BUTTON_RIGHT) {
		g.mouseRight = true
	} else if int(button) == int(sdl.BUTTON_MIDDLE) {
		g.pickBlock()
	}
}

func (g *Game) onMouseUp(button uint8) {
	if int(button) == int(sdl.BUTTON_LEFT) {
		g.mouseLeft = false
		g.breakActive = false
		g.breakProgress = 0
	} else if int(button) == int(sdl.BUTTON_RIGHT) {
		g.mouseRight = false
	}
}

func (g *Game) refreshMouse() {
	_, x, y := sdl.GetMouseState()
	g.mouseX, g.mouseY = g.toUI(x, y)
}

func (g *Game) toUI(x, y float32) (float32, float32) {
	ww, wh, err := g.Win.Size()
	if err != nil || ww <= 0 || wh <= 0 {
		return x, y
	}
	return x * float32(g.gfx.W) / float32(ww), y * float32(g.gfx.H) / float32(wh)
}

func (g *Game) updatePlaying(dt float32) {
	g.updateWorld(dt, true)
}

func (g *Game) updateWorld(dt float32, allowInput bool) {
	if g.player == nil {
		return
	}
	in := player.Input{}
	if allowInput && !g.chatOpen {
		ks := sdl.GetKeyboardState()
		down := func(code sdl.Scancode) bool { return int(code) < len(ks) && ks[code] }
		in.Forward = float32(boolToInt(down(sdl.SCANCODE_W))) - float32(boolToInt(down(sdl.SCANCODE_S)))
		in.Strafe = float32(boolToInt(down(sdl.SCANCODE_D))) - float32(boolToInt(down(sdl.SCANCODE_A)))
		in.Jump = down(sdl.SCANCODE_SPACE)
		in.Descend = down(sdl.SCANCODE_LCTRL) || down(sdl.SCANCODE_RCTRL)
		in.Sprint = down(sdl.SCANCODE_LSHIFT) || down(sdl.SCANCODE_RSHIFT)
	}
	g.ageChat(dt)
	g.player.Update(dt, g.world, in)
	if g.player.HurtPulse {
		g.playSound("hurt")
		g.player.HurtPulse = false
	}
	g.updateFootsteps(dt)
	g.mobs.Update(dt, g.world, g.player.Body.Pos, &g.inv)
	if g.mobs.PickupEvents > 0 {
		g.playSound("pickup")
		g.mobs.PickupEvents = 0
	}
	g.updateParticles(dt)
	g.time += float64(dt) / 600.0
	if g.time >= 1 {
		g.time -= 1
	}
	if g.attackCool > 0 {
		g.attackCool -= dt
	}
	if g.placeCool > 0 {
		g.placeCool -= dt
	}
	if g.pickCool > 0 {
		g.pickCool -= dt
	}
	if g.breakCool > 0 {
		g.breakCool -= dt
	}
	if g.spaceTapTimer > 0 {
		g.spaceTapTimer -= dt
	}
	if allowInput && !g.chatOpen {
		g.handleActions(dt)
	}
	g.fluidTimer -= dt
	if g.fluidTimer <= 0 {
		g.world.TickFluids()
		g.fluidTimer = 0.12
	}
	g.saveTimer += dt
	if g.saveTimer > 30 && g.state == StatePlaying && g.opts.SmokeFrames == 0 {
		g.saveTimer = 0
		if err := g.saveGame(); err == nil {
			g.notify(i18n.T("toast.autosaved"), 1.2)
		}
	}
	if g.player.Health <= 0 {
		g.setState(StateDead)
	}
}

func (g *Game) renderFrame() {
	if g.player == nil {
		g.gfx.Render(&render.Frame{
			Camera: render.Camera{Pos: mathx.V3(0, 70, 8), Dir: mathx.V3(0, -0.2, -1).Normalize(), Up: mathx.V3(0, 1, 0), Fov: mathx.Radians(70), Near: 0.08, Far: 400, Aspect: float32(g.gfx.W) / float32(g.gfx.H)},
			World:  g.sky(),
			UI:     g.canv,
		})
		return
	}
	dir := g.player.LookDir()
	cam := render.Camera{
		Pos: g.player.Eye(), Dir: dir, Up: mathx.V3(0, 1, 0),
		Fov: mathx.Radians(float32(g.opts.FOV)), Near: 0.06, Far: 500,
		Aspect: float32(g.gfx.W) / float32(g.gfx.H),
	}
	sky := g.sky()
	eye := g.player.Eye()
	eyeBlock := g.world.GetBlock(int(math.Floor(float64(eye.X))), int(math.Floor(float64(eye.Y))), int(math.Floor(float64(eye.Z))))
	if eyeBlock == core.Water {
		sky.FogColor = [3]float32{0.04, 0.20, 0.33}
		sky.Clear = [4]float32{0.04, 0.20, 0.33, 1}
		sky.FogStart = 1
		sky.FogEnd = 22
	}
	opaque, transparent, entityIcons := g.buildEntityMeshes()
	f := &render.Frame{
		Camera:            cam,
		World:             sky,
		Entity:            opaque,
		EntityTransparent: transparent,
		EntityIcon:        entityIcons,
		Held:              g.buildHeldMesh(),
		HeldTransparent:   g.buildHeldTransparentMesh(),
		HeldIcon:          g.buildHeldIconMesh(),
		UI:                g.canv,
	}
	if err := g.gfx.Render(f); err != nil {
		fmt.Println("render error:", err)
		g.quit = true
	}
}

func (g *Game) sky() render.WorldUniforms {
	t := g.time
	sunH := float32(math.Sin((t - 0.25) * 6.2831853))
	day := clampF(sunH*1.9+0.30, 0, 1)
	nightSky := [3]float32{0.012, 0.022, 0.055}
	daySky := [3]float32{0.42, 0.64, 0.93}
	sky := lerp3(nightSky, daySky, day)
	// Warm sunrise and sunset band.
	warm := clampF(1-absF(sunH)*3.5, 0, 1)
	sunset := [3]float32{0.92, 0.42, 0.20}
	sky = lerp3(sky, sunset, warm*0.45)
	sunTint := lerp3([3]float32{0.10, 0.13, 0.24}, [3]float32{1.0, 0.97, 0.90}, day)
	sunTint = lerp3(sunTint, [3]float32{1.0, 0.72, 0.50}, warm*0.35)
	fogEnd := float32(g.opts.RenderDist*16) - 8
	if fogEnd < 48 {
		fogEnd = 48
	}
	return render.WorldUniforms{
		FogColor: sky,
		FogStart: fogEnd * 0.55,
		FogEnd:   fogEnd,
		SunTint:  sunTint,
		Clear:    [4]float32{sky[0], sky[1], sky[2], 1},
	}
}

func (g *Game) toggleLanguage() {
	next := i18n.Next()
	i18n.Set(next)
	if g.Win != nil {
		_ = g.Win.SetTitle(i18n.T("window.title"))
	}
	g.notify(i18n.T("toast.language", i18n.LocaleName(next)), 2)
}

func (g *Game) shutdown() {
	if g.world != nil && g.player != nil {
		g.returnCraftGrid()
	}
	g.audio.Close()
	if g.world != nil && g.opts.SmokeFrames == 0 {
		_ = g.saveGame()
	}
	g.stopWorkers()
}

func (g *Game) dropSelected() {
	s := g.inv.SelectedStack()
	if s.Empty() {
		return
	}
	item := s.Item
	count := 1
	if g.shiftDown {
		count = s.Count
	}
	g.inv.TakeSelected(count)
	g.throwItem(item, count)
}

func (g *Game) respawn() {
	g.player.Respawn()
	g.player.Body.Pos = g.world.SafeSpawn(g.player.Body.Pos, 24)
	g.player.Spawning = g.player.Body.Pos
	g.player.FallingReset()
	g.setState(StatePlaying)
	g.notify(i18n.T("toast.respawned"), 1.5)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func onOff(b bool) string {
	if b {
		return i18n.T("common.on")
	}
	return i18n.T("common.off")
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
func clampF(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func lerp3(a, b [3]float32, t float32) [3]float32 {
	return [3]float32{a[0] + (b[0]-a[0])*t, a[1] + (b[1]-a[1])*t, a[2] + (b[2]-a[2])*t}
}
func (g *Game) playSound(name string) {
	if g.audio != nil {
		g.audio.Play(name)
	}
}
func (g *Game) updateFootsteps(dt float32) {
	if g.player == nil || g.chatOpen || g.player.Flying || !g.player.Body.OnGround {
		g.stepTimer = 0
		return
	}
	speed := float32(math.Hypot(float64(g.player.Body.Vel.X), float64(g.player.Body.Vel.Z)))
	if speed < 0.8 {
		g.stepTimer = 0
		return
	}
	g.stepTimer -= dt * speed
	if g.stepTimer <= 0 {
		g.playSound("step")
		g.stepTimer = 1.5
	}
}
func (g *Game) updateHeldLight(dt float32) {
	if g.world == nil || g.player == nil {
		return
	}
	if g.heldLightT > 0 {
		g.heldLightT -= dt
		if g.heldLightT > 0 {
			return
		}
	}
	g.heldLightT = 0.08
	level := uint8(0)
	if g.opts.DynamicLight && g.state != StateDead {
		if st := g.inv.SelectedStack(); !st.Empty() && st.Item.IsBlock() {
			level = st.Item.Def().Block.Def().Light
		}
	}
	pos := g.player.Body.Pos
	g.world.SetDynamicLight(
		int(math.Floor(float64(pos.X))),
		int(math.Floor(float64(pos.Y+1))),
		int(math.Floor(float64(pos.Z))),
		level,
	)
	g.world.RefreshDynamicLightIfNeeded()
}
