package main

import (
	"flag"
	"fmt"
	"gameboy/gb"
	"gameboy/io"
	"gameboy/logger"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faiface/pixel/pixelgl"
)

var (
	vsyncOff = flag.Bool("disableVsync", false, "set to disable vsync (debugging)")
	unlocked = flag.Bool("unlocked", false, "if to unlock the cpu speed (debugging)")
)

func setupExitHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // Captura Ctrl+C e SIGTERM

	go func() {
		<-c                                  // Aguarda o sinal
		fmt.Print(logger.GetRemainingLogs()) // Exibe os logs restantes
		os.Exit(0)                           // Encerra o programa
	}()
}

func main() {
	setupExitHandler()

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println("Programa interrompido:", r)
	// 		logger.GetRemainingLogs() // Exibe os logs restantes
	// 	}
	// }()

	flag.Parse()
	pixelgl.Run(start)

}

func start() {
	// Load the rom from the flag argument, or prompt with file select
	//rom := getROM()

	//rom := "./jogos/PokemonRed.gb"
	//rom := "./jogos/Pokemon - Yellow.gbc"
	rom := "./jogos/Pokemon - Gold.gbc"
	//rom := "./jogos/Tetris.gb"
	//rom := "./jogos/BombermanGB.gb"
	//rom := "./gb-test-roms/cpu_instrs/cpu_instrs.gb"
	//rom := "./gb-test-roms/instr_timing/instr_timing.gb"
	//rom := "./gb-test-roms/interrupt_time/interrupt_time.gb"
	//rom := "./gb-test-roms/mem_timing/mem_timing.gb"
	//rom := "./gb-test-roms/mem_timing-2/rom_singles/01-read_timing.gb"
	//rom := "./gb-test-roms/mem_timing-2/rom_singles/02-write_timing.gb"
	//rom := "./gb-test-roms/mem_timing-2/rom_singles/03-modify_timing.gb"

	// if *unlocked {
	// 	*mute = true
	// }

	// Initialise the GameBoy with the flag options
	gameboy, err := gb.NewGameboy(rom, false)
	if err != nil {
		log.Fatal(err)
	}

	// Create the monitor for pixels
	enableVSync := !(*vsyncOff || *unlocked)
	monitor := io.NewPixelsIOBinding(enableVSync, gameboy)

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
	}

	for range ticker.C {
		if !monitor.IsRunning() {
			return
		}

		frames++

		buttons := monitor.ButtonInput()
		gameboy.ProcessInput(buttons)

		_ = gameboy.Update()

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
