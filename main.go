package main

import (
	"flag"
	"fmt"
	"gameboy/gb"
	"gameboy/io"
	"log"
	"time"

	"github.com/faiface/pixel/pixelgl"
)

var (
	vsyncOff = flag.Bool("disableVsync", false, "set to disable vsync (debugging)")
	unlocked = flag.Bool("unlocked", false, "if to unlock the cpu speed (debugging)")
)

func main() {
	flag.Parse()
	pixelgl.Run(start)

}

func start() {
	// Load the rom from the flag argument, or prompt with file select
	//rom := getROM()

	//rom := "./PokemonRed.gb"
	//rom := "./Tetris.gb"
	//rom := "./BombermanGB.gb"
	rom := "./gb-test-roms/cpu_instrs/cpu_instrs.gb"
	//rom := "./gb-test-roms/instr_timing/instr_timing.gb"

	// if *unlocked {
	// 	*mute = true
	// }

	// Initialise the GameBoy with the flag options
	gameboy, err := gb.NewGameboy(rom)
	if err != nil {
		log.Fatal(err)
	}
	// if *stepThrough {
	// 	gameboy.Debug.OutputOpcodes = true
	// }

	// Create the monitor for pixels
	enableVSync := !(*vsyncOff || *unlocked)
	monitor := io.NewPixelsIOBinding(enableVSync, gameboy)

	//Debug
	// memoryView := io.NewMemoryView(gameboy)

	emulateCycle(gameboy, monitor)

}

func emulateCycle(gameboy *gb.Gameboy, monitor gb.IOBinding) {
	frameTime := time.Second / gb.FramesSecond

	if *unlocked {
		frameTime = 1
	}

	ticker := time.NewTicker(frameTime)
	start := time.Now()
	frames := 0

	var cartName string
	if gameboy.IsGameLoaded() {
		cartName = gameboy.Memory.Cart.GetName()
		// fmt.Println(cartName)
		// fmt.Println(!monitor.IsRunning())
		// fmt.Scanln()
	}

	for range ticker.C {
		if !monitor.IsRunning() {
			return
		}

		frames++

		// buttons := monitor.ButtonInput()
		// gameboy.ProcessInput(buttons)

		// fmt.Println(gameboy.CPU.Memory.Hram[0x44])
		// fmt.Scanln()

		//memoryView.RenderMemory(gameboy)
		_ = gameboy.Update()
		// memoryView.RenderMemory(gameboy)
		// fmt.Println("to aqui")
		// fmt.Scanln()
		monitor.Render(&gameboy.PreparedData)

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()

			title := fmt.Sprintf("Batata - %s (FPS: %2v)", cartName, frames)
			monitor.SetTitle(title)
			frames = 0
		}
	}
}
