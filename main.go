package main

import (
	"flag"
	"fmt"
	"os"

	"sevg/internal/game"
)

func main() {
	var opts game.Options
	flag.IntVar(&opts.Width, "width", 1280, "window width")
	flag.IntVar(&opts.Height, "height", 720, "window height")
	flag.Int64Var(&opts.Seed, "seed", 0, "world seed (0 = random)")
	flag.IntVar(&opts.RenderDist, "renderdist", 6, "chunk render distance")
	flag.StringVar(&opts.SavePath, "save", "savegame.gob", "save game path")
	flag.IntVar(&opts.SmokeFrames, "smoke", 0, "render N frames then exit (testing)")
	flag.BoolVar(&opts.Continue, "continue", false, "load the save game on startup")
	flag.StringVar(&opts.Language, "lang", "zh-CN", "UI language: zh-CN or en-US")
	flag.Parse()

	if err := game.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "SEVG:", err)
		os.Exit(1)
	}
	if opts.SmokeFrames > 0 {
		fmt.Printf("SMOKE OK: rendered %d frames\n", opts.SmokeFrames)
	}
}
