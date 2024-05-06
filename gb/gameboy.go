package gb

import (
	"fmt"
	"gameboy/bits"

	_ "github.com/faiface/pixel/pixelgl"
)

const (
	// ClockSpeed is the number of cycles the GameBoy CPU performs each second.
	ClockSpeed = 4194304
	// FramesSecond is the target number of frames for each frame of GameBoy output.
	FramesSecond = 60
	// CyclesFrame is the number of CPU cycles in each frame.
	CyclesFrame = ClockSpeed / FramesSecond
)

type Gameboy struct {
	//options gameboyOptions

	Memory *Memory
	CPU    *Z80
	//Sound  *apu.APU

	//Debug  DebugFlags
	//paused bool

	timerCounter int

	// Matrix of pixel data which is used while the screen is rendering. When a
	// frame has been completed, this data is copied into the PreparedData matrix.
	screenData [ScreenWidth][ScreenHeight][3]uint8
	bgPriority [ScreenWidth][ScreenHeight]bool

	// Track colour of tiles in scanline for priority management.
	tileScanline    [ScreenWidth]uint8
	scanlineCounter int
	screenCleared   bool

	// PreparedData is a matrix of screen pixel data for a single frame which has
	// been fully rendered.
	PreparedData [ScreenWidth][ScreenHeight][3]uint8

	interruptsOn bool
	halted       bool

	// Mask of the currently pressed buttons.
	inputMask byte

	// Flag if the game is running in cgb mode. For this to be true the game
	// rom must support cgb mode and the option be true.
	cgbMode       bool
	BGPalette     *cgbPalette
	SpritePalette *cgbPalette

	currentSpeed byte
	prepareSpeed bool

	//thisCpuTicks int

	//keyHandlers map[Button]func()
}

var interruptAddresses = map[byte]uint16{
	0: 0x40, // V-Blank
	1: 0x48, // LCDC Status
	2: 0x50, // Timer Overflow
	3: 0x58, // Serial Transfer
	4: 0x60, // Hi-Lo P10-P13
}

func (gb *Gameboy) getSpeed() int {
	return int(gb.currentSpeed + 1)
}

func (gb *Gameboy) isClockEnabled() bool {
	return bits.Test(gb.Memory.Hram[0x07] /* TAC */, 2)
}

func (gb *Gameboy) getClockFreq() byte {
	return gb.Memory.Hram[0x07] /* TAC */ & 0x3
}

func (gb *Gameboy) getClockFreqCount() int {
	switch gb.getClockFreq() {
	case 0:
		return 1024
	case 1:
		return 16
	case 2:
		return 64
	default:
		return 256
	}
}

func (gb *Gameboy) setClockFreq() {
	gb.timerCounter = 0
}

func (gb *Gameboy) IsCGB() bool {
	if gb.cgbMode {
		fmt.Printf("CGB: %t\n", gb.cgbMode)
		fmt.Scanln()
	}
	return gb.cgbMode
}

func (gb *Gameboy) IsGameLoaded() bool {
	return gb.Memory != nil && gb.Memory.Cart != nil
}

func (gb *Gameboy) pushStack(addr uint16) {
	gb.Memory.WriteByte(gb.CPU.SP-1, byte(uint16(addr&0xFF00)>>8))
	gb.Memory.WriteByte(gb.CPU.SP-2, byte(addr&0xFF))

	gb.CPU.SP -= 2
}

func (gb *Gameboy) requestInterrupt(interrupt byte) {
	// fmt.Printf("Interrupção: 0x%X\n", interrupt)
	// fmt.Scanln()
	req := gb.Memory.Hram[0x0F] | 0xE0
	req = bits.Set(req, interrupt)
	gb.Memory.WriteByte(0xFF0F, req)
}

func (gb *Gameboy) doInterrupts() (cycles int) {
	if gb.CPU.IME {
		gb.interruptsOn = true
		gb.CPU.IME = false
		return 0
	}
	if !gb.interruptsOn && !gb.halted {
		return 0
	}

	req := gb.Memory.Hram[0x0F] | 0xE0
	enabled := gb.Memory.Hram[0xFF]

	if req > 0 {
		var i byte
		for i = 0; i < 5; i++ {
			if bits.Test(req, i) && bits.Test(enabled, i) {
				gb.serviceInterrupt(i)
				return 20
			}
		}
	}
	return 0
}

func (gb *Gameboy) checkSpeedSwitch() {
	if gb.prepareSpeed {
		// Switch speed
		gb.prepareSpeed = false
		if gb.currentSpeed == 0 {
			gb.currentSpeed = 1
		} else {
			gb.currentSpeed = 0
		}
		gb.halted = false
	}
}

func (gb *Gameboy) serviceInterrupt(interrupt byte) {
	// If was halted without interrupts, do not jump or reset IF
	if !gb.interruptsOn && gb.halted {
		gb.halted = false
		return
	}
	gb.interruptsOn = false
	gb.halted = false

	req := gb.Memory.ReadHighRam(0xFF0F)
	req = bits.Reset(req, interrupt)
	gb.Memory.WriteByte(0xFF0F, req)

	gb.pushStack(gb.CPU.PC)
	gb.CPU.PC = interruptAddresses[interrupt]
	// fmt.Printf("Interrupção ADDR 0x%X\n", interruptAddresses[interrupt])
	// fmt.Scanln()
}

func (gb *Gameboy) BGMapString() string {
	out := ""
	for y := uint16(0); y < 0x20; y++ {
		out += fmt.Sprintf("%2x: ", y)
		for x := uint16(0); x < 0x20; x++ {
			out += fmt.Sprintf("%2x ", gb.Memory.ReadByte(0x9800+(y*0x20)+x))
		}
		out += "\n"
	}
	return out
}

// func (gb *Gameboy) printBGMap() {
// 	fmt.Printf("BG Map:\n%s", gb.BGMapString())
// }

func (gb *Gameboy) Update() int {
	// if gb.paused {
	// 	return 0
	// }

	cycles := 0
	for cycles < CyclesFrame*gb.getSpeed() {
		cyclesOp := 4
		if !gb.halted {
			// if gb.Debug.OutputOpcodes {
			// 	LogOpcode(gb, false)
			// }
			cyclesOp = gb.CPU.EmulateCycle()
			// } else {
			// 	// TODO: This is incorrect
		}
		cycles += cyclesOp

		gb.updateGraphics(cyclesOp)
		//memoryView.RenderMemory(gb)

		gb.updateTimers(cyclesOp)
		// fmt.Println("to aqui")
		// fmt.Scanln()
		cycles += gb.doInterrupts()

		//gb.Sound.Buffer(cyclesOp, gb.getSpeed())
	}
	// gb.printBGMap()
	// fmt.Scanln()
	return cycles
}

func (gb *Gameboy) dividerRegister(cycles int) {
	gb.CPU.Divider += cycles
	if gb.CPU.Divider >= 255 {
		gb.CPU.Divider -= 255
		gb.Memory.Hram[DIV-0xFF00]++
	}
}

func (gb *Gameboy) updateTimers(cycles int) {
	gb.dividerRegister(cycles)
	if gb.isClockEnabled() {
		gb.timerCounter += cycles

		freq := gb.getClockFreqCount()
		for gb.timerCounter >= freq {
			gb.timerCounter -= freq
			tima := gb.Memory.Hram[0x05] /* TIMA */
			if tima == 0xFF {
				gb.Memory.Hram[TIMA-0xFF00] = gb.Memory.Hram[0x06] /* TMA */
				gb.requestInterrupt(2)
			} else {
				gb.Memory.Hram[TIMA-0xFF00] = tima + 1
			}
		}
	}
}

func (gb *Gameboy) iniciar(romFile string) error {
	gb.setup()

	// Load the ROM file
	hasCGB, err := gb.Memory.LoadCart(romFile)
	if err != nil {
		return fmt.Errorf("failed to open rom file: %w", err)
	}
	gb.cgbMode = false && hasCGB
	return nil
}

func (gb *Gameboy) setup() {

	// Initialise the memory
	gb.Memory = &Memory{}
	gb.Memory.Init(gb)

	// Initialise the CPU
	gb.CPU = &Z80{}
	gb.CPU.Init(gb.Memory, true)

	// gb.Sound = &apu.APU{}
	// gb.Sound.Init(gb.options.sound)

	// gb.Debug = DebugFlags{}

	gb.scanlineCounter = 456
	gb.inputMask = 0xFF

	gb.SpritePalette = NewPalette()
	gb.BGPalette = NewPalette()

	// gb.initKeyHandlers()
}

func NewGameboy(romFile string) (*Gameboy, error) {
	gameboy := Gameboy{}

	err := gameboy.iniciar(romFile)
	if err != nil {
		return nil, err
	}
	return &gameboy, nil
}
