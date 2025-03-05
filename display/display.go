package display

import (
	"github.com/veandco/go-sdl2/sdl"
)

const (
	ScreenWidth  = 160
	ScreenHeight = 144
	TITLE        = "Gameboy boladão"
	WIDTH        = 160
	HEIGHT       = 144
)

func InitDisplay() (*sdl.Window, *sdl.Surface) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(TITLE, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WIDTH, HEIGHT, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	return window, surface
}

func UpdateDisplay(window *sdl.Window, surface *sdl.Surface) {

	window.UpdateSurface()
}
