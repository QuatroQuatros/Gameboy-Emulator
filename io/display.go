package io

import (
	"gameboy/gb"
	"image/color"
	"log"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var displayScale float64 = 3

type PixelsIOBinding struct {
	window  *pixelgl.Window
	picture *pixel.PictureData
}

func NewPixelsIOBinding(enableVSync bool, gameboy *gb.Gameboy) *PixelsIOBinding {
	windowConfig := pixelgl.WindowConfig{
		Title: "Batata",
		Bounds: pixel.R(
			0, 0,
			float64(gb.ScreenWidth*displayScale), float64(gb.ScreenHeight*displayScale),
		),
		VSync:     enableVSync,
		Resizable: true,
	}

	window, err := pixelgl.NewWindow(windowConfig)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	picture := &pixel.PictureData{
		Pix:    make([]color.RGBA, gb.ScreenWidth*gb.ScreenHeight),
		Stride: gb.ScreenWidth,
		Rect:   pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight),
	}

	monitor := PixelsIOBinding{
		window:  window,
		picture: picture,
	}

	monitor.updateCamera()

	return &monitor
}

func (mon *PixelsIOBinding) updateCamera() {
	xScale := mon.window.Bounds().W() / 160
	yScale := mon.window.Bounds().H() / 144
	scale := math.Min(yScale, xScale)

	shift := mon.window.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	mon.window.SetMatrix(cam)
}

func (mon *PixelsIOBinding) IsRunning() bool {
	return !mon.window.Closed()
}

func (mon *PixelsIOBinding) Render(screen *[160][144][3]uint8) {

	for y := 0; y < gb.ScreenHeight; y++ {
		for x := 0; x < gb.ScreenWidth; x++ {
			col := screen[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(gb.ScreenHeight-1-y)*gb.ScreenWidth+x] = rgb
		}
	}

	r, g, b := gb.GetPaletteColour(3)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.window.Clear(bg)

	// fmt.Println("TO AQUI")
	// fmt.Println(bg)
	// fmt.Scanln()

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight))
	spr.Draw(mon.window, pixel.IM)

	mon.updateCamera()
	mon.window.Update()
}

// SetTitle sets the title of the game window.
func (mon *PixelsIOBinding) SetTitle(title string) {
	mon.window.SetTitle(title)
}

func (mon *PixelsIOBinding) toggleFullscreen() {
	if mon.window.Monitor() == nil {
		monitor := pixelgl.PrimaryMonitor()
		_, height := monitor.Size()
		mon.window.SetMonitor(monitor)
		displayScale = height / 144
	} else {
		mon.window.SetMonitor(nil)
		displayScale = 3
	}
}

var keyMap = map[pixelgl.Button]gb.Button{
	pixelgl.KeyZ:         gb.ButtonA,
	pixelgl.KeyX:         gb.ButtonB,
	pixelgl.KeyBackspace: gb.ButtonSelect,
	pixelgl.KeyEnter:     gb.ButtonStart,
	pixelgl.KeyRight:     gb.ButtonRight,
	pixelgl.KeyLeft:      gb.ButtonLeft,
	pixelgl.KeyUp:        gb.ButtonUp,
	pixelgl.KeyDown:      gb.ButtonDown,

	pixelgl.KeyEscape: gb.ButtonPause,
	pixelgl.KeyEqual:  gb.ButtonChangePallete,
	pixelgl.KeyQ:      gb.ButtonToggleBackground,
	pixelgl.KeyW:      gb.ButtonToggleSprites,
	pixelgl.KeyE:      gb.ButttonToggleOutputOpCode,
	pixelgl.KeyD:      gb.ButtonPrintBGMap,
	pixelgl.Key7:      gb.ButtonToggleSoundChannel1,
	pixelgl.Key8:      gb.ButtonToggleSoundChannel2,
	pixelgl.Key9:      gb.ButtonToggleSoundChannel3,
	pixelgl.Key0:      gb.ButtonToggleSoundChannel4,
}

// ProcessInput checks the input and process it.
func (mon *PixelsIOBinding) ButtonInput() gb.ButtonInput {

	if mon.window.JustPressed(pixelgl.KeyF) {
		mon.toggleFullscreen()
	}

	var buttonInput gb.ButtonInput

	for handledKey, button := range keyMap {
		if mon.window.JustPressed(handledKey) {
			buttonInput.Pressed = append(buttonInput.Pressed, button)
		}
		if mon.window.JustReleased(handledKey) {
			buttonInput.Released = append(buttonInput.Released, button)
		}
	}

	return buttonInput
}
