package gb

import(
	"fmt"
	"gameboy/bits"
	_"github.com/faiface/pixel/pixelgl"
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
	paused bool

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

	interruptsEnabling bool
	interruptsOn       bool
	halted             bool

	//cbInst [0x100]func()

	// Mask of the currently pressed buttons.
	inputMask byte

	// Flag if the game is running in cgb mode. For this to be true the game
	// rom must support cgb mode and the option be true.
	cgbMode       bool
	// BGPalette     *cgbPalette
	// SpritePalette *cgbPalette

	currentSpeed byte
	prepareSpeed bool

	thisCpuTicks int

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

func (gb *Gameboy) getClockFreq() byte {
	return gb.Memory.Hram[0x07] /* TAC */ & 0x3
}

func (gb *Gameboy) setClockFreq() {
	gb.timerCounter = 0
}

func (gb *Gameboy) pushStack(address uint16) {
	sp := gb.CPU.SP
	gb.Memory.WriteByte(sp-1, byte(uint16(address&0xFF00)>>8))
	gb.Memory.WriteByte(sp-2, byte(address&0xFF))
	gb.CPU.SP = gb.CPU.SP - 2
}

func (gb *Gameboy) requestInterrupt(interrupt byte) {
	req := gb.Memory.Hram[0x0F] | 0xE0
	req = bits.Set(req, interrupt)
	gb.Memory.WriteByte(0xFF0F, req)
}

func (gb *Gameboy) doInterrupts() (cycles int) {
	// if gb.interruptsEnabling {
	// 	gb.interruptsOn = true
	// 	gb.interruptsEnabling = false
	// 	return 0
	// }
	// if !gb.interruptsOn && !gb.halted {
	// 	return 0
	// }

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
}



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

		//gb.updateTimers(cyclesOp)

		cycles += gb.doInterrupts()

		//gb.Sound.Buffer(cyclesOp, gb.getSpeed())
	}
	return cycles
}

// Get the current CPU speed multiplier (either 1 or 2).
// func (gb *Gameboy) getSpeed() int {
// 	return int(gb.currentSpeed + 1)
// }


func (gb *Gameboy) iniciar(romFile string) error {
	gb.setup()

	// Load the ROM file
	err := gb.Memory.LoadGame(romFile)
	if err != nil {
		return fmt.Errorf("failed to open rom file: %w", err)
	}
	return nil
}

func (gb *Gameboy) setup() {

	// Initialise the memory
	gb.Memory = &Memory{}
	gb.Memory.Init(gb)

	// Initialise the CPU
	gb.CPU = &Z80{}
	gb.CPU.Init(gb.Memory)


	// gb.Sound = &apu.APU{}
	// gb.Sound.Init(gb.options.sound)

	// gb.Debug = DebugFlags{}

	gb.scanlineCounter = 456
	gb.inputMask = 0xFF

	// gb.cbInst = gb.cbInstructions()

	// gb.SpritePalette = NewPalette()
	// gb.BGPalette = NewPalette()

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