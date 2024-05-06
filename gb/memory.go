package gb

import (
	"fmt"
	"gameboy/bits"
	"gameboy/cart"
	"log"
)

type Memory struct {
	gb       *Gameboy
	Cart     *cart.Cart
	Bios     [256]byte    // BIOS (Read-only)
	Rom      [0x8000]byte // ROM (Read-only)
	Ram      [0x2000]byte // RAM
	Vram     [0x4000]byte // Video RAM
	VramBank byte
	Wram     [0x9000]byte // Working RAM
	WramBank byte
	Echo     [0x1E00]byte // Echo RAM

	Oam  [0x100]byte // OAM (Object Attribute Memory) Oam  [0xA0]byte
	Hram [0x100]byte // High RAM

	hdmaLength byte
	hdmaActive bool
}

const (
	// DIV is the divider register which is incremented periodically by
	// the Gameboy.
	DIV = 0xFF04
	// TIMA is the timer counter register which is incremented by a clock
	// frequency specified in the TAC register.
	TIMA = 0xFF05
	// TMA is the timer modulo register. When the TIMA overflows, this data
	// will be loaded into the TIMA register.
	TMA = 0xFF06
	// TAC is the timer control register. Writing to this register will
	// start and stop the timer, and select the clock speed for the timer.
	TAC = 0xFF07

	// TODO: move more hardware registers up here.
)

func (m *Memory) Init(gameboy *Gameboy) {
	m.gb = gameboy

	m.Hram[0x04] = 0x1E
	m.Hram[0x05] = 0x00
	m.Hram[0x06] = 0x00
	m.Hram[0x07] = 0xF8
	m.Hram[0x0F] = 0xE1
	m.Hram[0x10] = 0x80
	m.Hram[0x11] = 0xBF
	m.Hram[0x12] = 0xF3
	m.Hram[0x14] = 0xBF
	m.Hram[0x16] = 0x3F
	m.Hram[0x17] = 0x00
	m.Hram[0x19] = 0xBF
	m.Hram[0x1A] = 0x7F
	m.Hram[0x1B] = 0xFF
	m.Hram[0x1C] = 0x9F
	m.Hram[0x1E] = 0xBF
	m.Hram[0x20] = 0xFF
	m.Hram[0x21] = 0x00
	m.Hram[0x22] = 0x00
	m.Hram[0x23] = 0xBF
	m.Hram[0x24] = 0x77
	m.Hram[0x25] = 0xF3
	m.Hram[0x26] = 0xF1
	m.Hram[0x40] = 0x91
	m.Hram[0x41] = 0x85
	m.Hram[0x42] = 0x00
	m.Hram[0x43] = 0x00
	m.Hram[0x45] = 0x00
	m.Hram[0x47] = 0xFC
	m.Hram[0x48] = 0xFF
	m.Hram[0x49] = 0xFF
	m.Hram[0x4A] = 0x00
	m.Hram[0x4B] = 0x00
	m.Hram[0xFF] = 0x00

	//teste
	//m.Hram[0x44] = 20

	m.WramBank = 1
}

func (m *Memory) LoadCart(rom string) (bool, error) {
	var err error
	m.Cart, err = cart.NewCartFromFile(rom)
	if err != nil {
		return false, err
	}
	return m.Cart.GetMode()&cart.CGB != 0, nil
}

func (m *Memory) ReadByte(addr uint16) byte {
	// fmt.Printf("Vou LER na RAM  ADDR: 0x%X\n", addr)
	//fmt.Scanln()
	switch {
	case addr < 0x8000: // ROM
		return m.Cart.Read(addr)

	case addr < 0xA000: // Video RAM
		bankOffset := uint16(m.VramBank) * 0x2000
		return m.Vram[addr-0x8000+bankOffset]

	case addr < 0xC000: // Cartridge RAM
		return m.Cart.Read(addr)

	case addr < 0xD000:
		// Internal RAM - Bank 0
		return m.Wram[addr-0xC000]

	case addr < 0xE000:
		// Internal RAM Bank 1-7
		return m.Wram[(addr-0xC000)+(uint16(m.WramBank)*0x1000)]

	case addr < 0xFE00:
		// Echo RAM
		// TODO: re-enable echo RAM?
		//mem.Data[addr] = value
		//mem.Write(addr-0x2000, value)
		return 0xFF
	case addr < 0xFEA0:
		// Object Attribute Memory
		return m.Oam[addr-0xFE00]

	case addr < 0xFF00: // Unusable memory
		return 0xFF

	default:
		return m.ReadHighRam(addr)
	}
}

func (m *Memory) ReadHighRam(addr uint16) byte {
	// fmt.Printf("Vou LER na HRAM  ADDR: 0x%X\n", addr)
	//fmt.Scanln()
	switch {
	// Joypad address
	case addr == 0xFF00:
		return 0
	// 	return mem.gb.joypadValue(mem.HighRAM[0x00])

	case addr >= 0xFF10 && addr <= 0xFF26:
		return 0
	// 	return mem.gb.Sound.Read(addr)

	case addr >= 0xFF30 && addr <= 0xFF3F:
		return 0
	// 	// Writing to channel 3 waveform RAM.
	// 	return mem.gb.Sound.Read(addr)

	case addr == 0xFF0F:
		return m.Hram[0x0F] | 0xE0

	case addr >= 0xFF72 && addr <= 0xFF77:
		return 0

	case addr == 0xFF68:
		// BG palette index
		if m.gb.IsCGB() {
			return m.gb.BGPalette.readIndex()
		}
		return 0

	case addr == 0xFF69:
		// BG Palette data
		if m.gb.IsCGB() {
			return m.gb.BGPalette.read()
		}
		return 0

	case addr == 0xFF6A:
		// Sprite palette index
		if m.gb.IsCGB() {
			return m.gb.SpritePalette.readIndex()
		}
		return 0

	case addr == 0xFF6B:
		// Sprite Palette data
		if m.gb.IsCGB() {
			return m.gb.SpritePalette.read()
		}
		return 0

	case addr == 0xFF4D:
		// Speed switch data
		return m.gb.currentSpeed<<7 | bits.B(m.gb.prepareSpeed)

	case addr == 0xFF4F:
		return m.VramBank

	case addr == 0xFF70:
		return m.WramBank

	default:
		// fmt.Println("TO AQUI HRAM")
		// fmt.Printf("VALUE: 0x%X\n", m.Hram[address-0xFF00])
		// fmt.Scanln()
		// fmt.Printf("VOU TER QUE LER AQUI  ADDR: 0x%X  VALUE:  0x%X\n", addr-0xFF00, m.Hram[addr-0xFF00])
		//fmt.Scanln()
		return m.Hram[addr-0xFF00]
	}
}

func (m *Memory) WriteByte(addr uint16, value byte) {
	// fmt.Printf("Vou ESCREVER na RAM  ADDR: 0x%X  VALUE:  0x%X\n", addr, value)
	switch {
	case addr < 0x8000:
		m.Cart.WriteROM(addr, value)

	case addr < 0xA000: // Video RAM
		bankOffset := uint16(m.VramBank) * 0x2000
		m.Vram[addr-0x8000+bankOffset] = value
		// fmt.Println("Escrevi na VRAM")
		// fmt.Println(value)
		// fmt.Scanln()

	case addr < 0xC000:
		// Cartridge ram
		m.Cart.WriteRAM(addr, value)
		//m.Rom[addr] = value

	case addr < 0xD000:
		// Internal RAM - Bank 0
		m.Wram[addr-0xC000] = value

	case addr < 0xE000:
		// Internal RAM Bank 1-7
		m.Wram[(addr-0xC000)+(uint16(m.WramBank)*0x1000)] = value

	case addr < 0xFE00:
		fmt.Println("ECHO")
		fmt.Scanln()

	case addr < 0xFEA0:
		m.Oam[addr-0xFE00] = value

	case addr < 0xFF00:
		// Unusable memory
		break

	default:
		// High RAM
		m.WriteHighRam(addr, value)
	}
}

func (m *Memory) WriteHighRam(addr uint16, value byte) {
	// fmt.Printf("Vou ESCREVER na HRAM  ADDR: 0x%X  VALUE:  0x%X\n", addr, value)
	//fmt.Scanln()
	switch {

	case addr >= 0xFEA0 && addr < 0xFEFF:
		// Restricted RAM
		return

	case addr >= 0xFF10 && addr <= 0xFF26:
		return
		// m.gb.Sound.Write(address, value)

	case addr >= 0xFF30 && addr <= 0xFF3F:
		return
		// Writing to channel 3 waveform RAM.
		// m.gb.Sound.WriteWaveform(addr, value)

	case addr == 0xFF02:
		//Serial transfer control
		// if value == 0x81{
		// 	fmt.Println("Transfer")
		// 	fmt.Println(m.ReadHighRam(0xFF01))

		// 	fmt.Scanln()
		// }else{
		// 	fmt.Printf("Transfer Value: 0x%X\n", value)
		// 	fmt.Println(m.ReadHighRam(0xFF01))
		// 	fmt.Scanln()
		// }

		// if value == 0x81 {
		// 	f := m.gb.options.transferFunction
		// 	if f != nil {
		// 		f(m.ReadHighRam(0xFF01))
		// 	}
		// }

	case addr == DIV:
		// Trap divider register
		m.gb.setClockFreq()
		m.gb.CPU.Divider = 0
		m.Hram[DIV-0xFF00] = 0

	case addr == TIMA:
		m.Hram[TIMA-0xFF00] = value

	case addr == TMA:
		m.Hram[TMA-0xFF00] = value

	case addr == TAC:
		// Timer control
		currentFreq := m.gb.getClockFreq()
		m.Hram[TAC-0xFF00] = value | 0xF8
		newFreq := m.gb.getClockFreq()

		if currentFreq != newFreq {
			m.gb.setClockFreq()
		}

	case addr == 0xFF41:
		m.Hram[0x41] = value | 0x80

	case addr == 0xFF44:
		// Trap scanline register
		// fmt.Println("batata")
		// fmt.Scanln()
		m.Hram[0x44] = 0

	case addr == 0xFF46:
		// DMA transfer
		m.doDMATransfer(value)

	case addr == 0xFF4D:
		//CGB speed change
		if m.gb.IsCGB() {
			m.gb.prepareSpeed = bits.Test(value, 0)
		}

	case addr == 0xFF4F:
		//VRAM bank (CGB only), blocked when HDMA is active
		if m.gb.IsCGB() && !m.hdmaActive {
			m.VramBank = value & 0x1
		}

	case addr == 0xFF55:
		// CGB DMA transfer
		if m.gb.IsCGB() {
			m.doNewDMATransfer(value)
		}

	case addr == 0xFF68:
		//BG palette index
		if m.gb.IsCGB() {
			m.gb.BGPalette.updateIndex(value)
		}
		return

	case addr == 0xFF69:
		//BG Palette data
		if m.gb.IsCGB() {
			m.gb.BGPalette.write(value)
		}

	case addr == 0xFF6A:
		//Sprite palette index
		if m.gb.IsCGB() {
			m.gb.SpritePalette.updateIndex(value)
		}

	case addr == 0xFF6B:
		//Sprite Palette data
		if m.gb.IsCGB() {
			m.gb.SpritePalette.write(value)
		}

	case addr == 0xFF70:
		//WRAM1 bank (CGB mode)
		if m.gb.IsCGB() {
			m.WramBank = value & 0x7
			if m.WramBank == 0 {
				m.WramBank = 1
			}
		}

	case addr >= 0xFF72 && addr <= 0xFF77:
		log.Print("write to ", addr)

	default:
		// fmt.Printf("VOU TER QUE ESCREVER AQUI ADDR: 0x%X VALUE: 0x%X\n", addr-0xFF00, value)
		m.Hram[addr-0xFF00] = value
	}
}

func (m *Memory) doNewDMATransfer(value byte) {
	if m.hdmaActive && bits.Val(value, 7) == 0 {
		// Abort a HDMA transfer
		m.hdmaActive = false
		m.Hram[0x55] |= 0x80 // Set bit 7
		return
	}

	length := ((uint16(value) & 0x7F) + 1) * 0x10

	// The 7th bit is DMA mode
	if value>>7 == 0 {
		// Mode 0, general purpose DMA
		m.performNewDMATransfer(length)
		m.Hram[0x55] = 0xFF
	} else {
		// Mode 1, H-Blank DMA
		m.hdmaLength = byte(value)
		m.hdmaActive = true
	}
}

func (m *Memory) doHDMATransfer() {
	if !m.hdmaActive {
		return
	}

	m.performNewDMATransfer(0x10)
	if m.hdmaLength > 0 {
		m.hdmaLength--
		m.Hram[0x55] = m.hdmaLength
	} else {
		// DMA has finished
		m.Hram[0x55] = 0xFF
		m.hdmaActive = false
	}
}

func (m *Memory) doDMATransfer(value byte) {
	address := uint16(value) << 8 // (data * 100)

	var i uint16
	for i = 0; i < 0xA0; i++ {
		m.WriteByte(0xFE00+i, m.ReadByte(address+i))
	}
}

func (mem *Memory) performNewDMATransfer(length uint16) {
	// Load the source and destination from RAM
	source := (uint16(mem.Hram[0x51])<<8 | uint16(mem.Hram[0x52])) & 0xFFF0
	destination := (uint16(mem.Hram[0x53])<<8 | uint16(mem.Hram[0x54])) & 0x1FF0
	destination += 0x8000

	// Transfer the data from the source to the destination
	for i := uint16(0); i < length; i++ {
		mem.WriteByte(destination, mem.ReadByte(source))
		destination++
		source++
	}

	// Update the source and destination in RAM
	mem.Hram[0x51] = byte(source >> 8)
	mem.Hram[0x52] = byte(source & 0xFF)
	mem.Hram[0x53] = byte(destination >> 8)
	mem.Hram[0x54] = byte(destination & 0xF0)
}

func (m *Memory) ReadWord(addr uint16) uint16 {
	// Lê dois bytes consecutivos da memória e combina-os em um valor de 16 bits
	lowByte := uint16(m.ReadByte(addr))
	highByte := uint16(m.ReadByte(addr + 1))
	return (highByte << 8) | lowByte
}

func (m *Memory) WriteWord(addr uint16, value uint16) {
	// Escreve os bytes da palavra (little-endian) no endereço especificado
	m.WriteByte(addr, byte(value&0xFF))        // Escreve o byte menos significativo
	m.WriteByte(addr+1, byte((value>>8)&0xFF)) // Escreve o byte mais significativo
}
