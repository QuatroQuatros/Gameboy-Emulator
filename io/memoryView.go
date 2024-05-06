package io

import (
	"fmt"
	"gameboy/gb"
	"image/color"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

var basicAtlas *text.Atlas

const (
	Width  = 500
	Height = 500
)

func NewMemoryView(gameboy *gb.Gameboy) *PixelsIOBinding {
	windowConfig := pixelgl.WindowConfig{
		Title: "Memory",
		Bounds: pixel.R(
			0, 0, float64(Width), float64(Height),
		),
		VSync:     true,
		Resizable: true,
	}

	window, err := pixelgl.NewWindow(windowConfig)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	monitor := PixelsIOBinding{
		window: window,
	}

	return &monitor
}

func (mon *PixelsIOBinding) RenderMemory(gb *gb.Gameboy) {
	basicAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
	mon.window.Clear(color.Black)
	//basicTxt.Clear()

	startAddress := 0x0000
	endAddress := 0x100

	var x float64 = 10
	lineHeight := 15
	//maxLines := 30  // Número máximo de linhas que cabem na tela
	maxX := 490 // Limite máximo de coordenada X para iniciar uma nova linha à direita

	basicTxt := text.New(pixel.V(x, 490), basicAtlas)

	for addr := startAddress; addr < endAddress; addr++ {

		value := gb.CPU.Memory.Hram[addr]
		addrString := fmt.Sprintf("0x%X: 0x%X", addr, value)

		// textWidth := basicAtlas.TextWidth(addrString)

		if basicTxt.BoundsOf(addrString).Max.Y > float64(maxX) {
			// Se ultrapassar, incrementa a coordenada Y e reseta a coordenada X
			x += 10
			basicTxt.Dot.Y -= float64(lineHeight)
		}

		fmt.Fprintln(basicTxt, addrString)
	}
	basicTxt.Draw(mon.window, pixel.IM)
	// basicTxt.Draw(mon.window, pixel.IM.Scaled(basicTxt.Orig, 0.5))

	mon.window.Update()
}
