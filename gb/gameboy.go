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
	//FramesSecond = 59.7
	// CyclesFrame is the number of CPU cycles in each frame.
	//CyclesFrame = 70224
	//CyclesFrame = 65664
	CyclesFrame = ClockSpeed / FramesSecond
)

type Gameboy struct {
	Memory *Memory
	CPU    *Z80
	//Sound  *apu.APU

	timerCounter int

	paused bool

	screenData [ScreenWidth][ScreenHeight][3]uint8
	bgPriority [ScreenWidth][ScreenHeight]bool

	// Track colour of tiles in scanline for priority management.
	tileScanline    [ScreenWidth]uint8
	scanlineCounter int
	screenCleared   bool

	PreparedData [ScreenWidth][ScreenHeight][3]uint8

	halted bool

	// Mask of the currently pressed buttons.
	inputMask byte

	// Flag if the game is running in cgb mode. For this to be true the game
	// rom must support cgb mode and the option be true.
	cgbMode       bool
	BGPalette     *cgbPalette
	SpritePalette *cgbPalette

	currentSpeed byte
	prepareSpeed bool

	keyHandlers map[Button]func()
}

func (gb *Gameboy) joypadValue(current byte) byte {
	var in byte = 0xF
	if bits.Test(current, 4) {
		in = gb.inputMask & 0xF
	} else if bits.Test(current, 5) {
		in = (gb.inputMask >> 4) & 0xF
	}
	return current | 0xc0 | in
}

func (gb *Gameboy) togglePaused() {
	gb.paused = !gb.paused
}

func (gb *Gameboy) getSpeed() int {
	return int(gb.currentSpeed + 1)
}

func (gb *Gameboy) IsCGB() bool {
	// if gb.cgbMode {
	// 	fmt.Printf("CGB: %t\n", gb.cgbMode)
	// 	fmt.Scanln()
	// }
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

func (gb *Gameboy) Update() int {
	cycles := 0
	//targetCycles := CyclesFrame * gb.getSpeed()

	//for cycles+4 < CyclesFrame*gb.getSpeed() {
	for cycles < CyclesFrame*gb.getSpeed() {
		cyclesOp := 4
		if !gb.halted {
			cyclesOp = gb.CPU.EmulateCycle()
		}

		// Atualiza gráficos e timers com o número correto de ciclos
		gb.updateGraphics(cyclesOp)
		gb.updateTimers(cyclesOp)
		cycles += cyclesOp
		cycles += gb.doInterrupts()

		//gb.Sound.Buffer(cyclesOp, gb.getSpeed())
	}

	return cycles
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

// func (gb *Gameboy) ToggleSoundChannel(channel int) {
// 	gb.Sound.ToggleSoundChannel(channel)
// }

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

func (gb *Gameboy) dividerRegister(cycles int) {
	gb.CPU.Divider += cycles
	if gb.CPU.Divider >= 255 {
		gb.CPU.Divider -= 255
		gb.Memory.Hram[DIV-0xFF00]++
	}
}

// Request the Gameboy to perform an interrupt.
func (gb *Gameboy) requestInterrupt(interrupt byte) {
	req := gb.Memory.Hram[0x0F] | 0xE0
	req = bits.Set(req, interrupt)
	gb.Memory.WriteByte(0xFF0F, req)
}

func (gb *Gameboy) doInterrupts() (cycles int) {
	if gb.CPU.InterruptsEnabling {
		gb.CPU.IME = true
		gb.CPU.InterruptsEnabling = false
		return 0
	}
	if !gb.CPU.IME && !gb.halted {
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

func (gb *Gameboy) serviceInterrupt(interrupt byte) {
	// If was halted without interrupts, do not jump or reset IF
	if !gb.CPU.IME && gb.halted {
		gb.halted = false
		return
	}
	gb.CPU.IME = false
	gb.halted = false

	req := gb.Memory.ReadHighRam(0xFF0F)
	req = bits.Reset(req, interrupt)
	gb.Memory.WriteByte(0xFF0F, req)

	gb.pushStack(gb.CPU.PC)
	gb.CPU.PC = interruptAddresses[interrupt]
}

var interruptAddresses = map[byte]uint16{
	0: 0x40, // V-Blank
	1: 0x48, // LCDC Status
	2: 0x50, // Timer Overflow
	3: 0x58, // Serial Transfer
	4: 0x60, // Hi-Lo P10-P13
}

func (gb *Gameboy) iniciar(romFile string, isCBG bool) error {
	gb.setup(isCBG)

	// Load the ROM file
	hasCGB, err := gb.Memory.LoadCart(romFile)
	if err != nil {
		return fmt.Errorf("failed to open rom file: %w", err)
	}
	// gb.cgbMode = false && hasCGB

	gb.cgbMode = isCBG && hasCGB

	return nil
}

func (gb *Gameboy) initKeyHandlers() {
	gb.keyHandlers = map[Button]func(){
		ButtonPause:         gb.togglePaused,
		ButtonChangePallete: changePallete,
		//ButtonToggleBackground:    gb.Debug.toggleBackGround,
		//ButtonToggleSprites:       gb.Debug.toggleSprites,
		//ButttonToggleOutputOpCode: gb.Debug.toggleOutputOpCode,
		//ButtonPrintBGMap:          gb.printBGMap,
		// ButtonToggleSoundChannel1: func() { gb.ToggleSoundChannel(1) },
		// ButtonToggleSoundChannel2: func() { gb.ToggleSoundChannel(2) },
		// ButtonToggleSoundChannel3: func() { gb.ToggleSoundChannel(3) },
		// ButtonToggleSoundChannel4: func() { gb.ToggleSoundChannel(4) },
	}
}

func (gb *Gameboy) setup(isCBG bool) {

	// Initialise the memory
	gb.Memory = &Memory{}
	gb.Memory.Init(gb)

	// Initialise the CPU
	gb.CPU = &Z80{}
	gb.CPU.Init(gb.Memory, isCBG)

	// gb.Sound = &apu.APU{}
	// gb.Sound.Init(true)

	gb.scanlineCounter = 456
	gb.inputMask = 0xFF

	gb.SpritePalette = NewPalette()
	gb.BGPalette = NewPalette()

	gb.initKeyHandlers()
}

// func NewGameboy(romFile string) (*Gameboy, error) {
// 	gameboy := Gameboy{}

// 	err := gameboy.iniciar(romFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &gameboy, nil
// }

func NewGameboy(romFile string, isCBG bool) (*Gameboy, error) {
	// Criar uma única instância de Memory
	memory := &Memory{}

	// Criar a instância do Gameboy e associar a mesma memória
	gameboy := &Gameboy{
		Memory: memory,
		CPU: &Z80{
			Memory: memory, // CPU usa a mesma instância de Memory
		},
	}

	err := gameboy.iniciar(romFile, isCBG)
	if err != nil {
		return nil, err
	}
	return gameboy, nil
}
