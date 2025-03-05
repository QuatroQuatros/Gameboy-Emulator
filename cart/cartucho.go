package cart

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Mode int

const (
	DMG Mode = 1 << iota
	CGB
)

type BankingController interface {
	// Read returns the value from the cartridges ROM or RAM depending on
	// the banking.
	Read(address uint16) byte

	// WriteROM attempts to write a value to an address in ROM. This is
	// generally used for switching memory banks depending on the implementation.

	WriteROM(address uint16, value byte)

	// WriteRAM sets a value on an address in the internal cartridge RAM.
	// Like the ROM, this can be banked depending on the implementation
	// of the memory controller. Furthermore, if the cartridge supports
	// RAM+BATTERY, then this data can be saved between sessions.

	WriteRAM(address uint16, value byte)

	// GetSaveData returns the save data for this banking controller. In
	// general this will the contents of the RAM, however controllers may
	// choose to store this data in their own format.

	//GetSaveData() []byte

	// LoadSaveData loads some save data into the cartridge. The banking
	// controller implementation can decide how this data should be loaded.

	//LoadSaveData(data []byte)
}

type Cart struct {
	BankingController
	title    string
	filename string
	mode     Mode
}

func (c *Cart) GetMode() Mode {
	return c.mode
}

func (c *Cart) GetName() string {
	if c.title == "" {
		for i := uint16(0x0134); i < 0x142; i++ {
			chr := c.Read(i)
			if chr != 0x00 {
				c.title += string(chr)
			}
		}
		c.title = strings.TrimSpace(c.title)
	}
	return c.title
}

func NewCartFromFile(filename string) (*Cart, error) {
	rom, err := loadROMData(filename)
	if err != nil {
		return nil, err
	}
	return NewCart(rom, filename), nil
}

func NewCart(rom []byte, filename string) *Cart {
	cartridge := Cart{
		filename: filename,
	}

	// Check for GB mode
	switch rom[0x0143] {
	case 0x80:
		cartridge.mode = DMG | CGB
	case 0xC0:
		cartridge.mode = CGB
	default:
		cartridge.mode = DMG
	}

	// Determine cartridge type
	mbcFlag := rom[0x147]
	cartType := "Unknown"
	switch mbcFlag {
	case 0x00, 0x08, 0x09, 0x0B, 0x0C, 0x0D:
		cartType = "ROM"
		cartridge.BankingController = NewROM(rom)
	default:
		switch {
		case mbcFlag <= 0x03:
			cartridge.BankingController = NewMBC1(rom)
			cartType = "MBC1"
		case mbcFlag <= 0x06:
			log.Println("Warning: MBC2 carts are not supported.")
			//cartridge.BankingController = NewMBC2(rom)
			//cartType = "MBC2"
		case mbcFlag <= 0x13:
			//log.Println("Warning: MBC3 carts are not supported.")
			cartridge.BankingController = NewMBC3(rom)
			cartType = "MBC3"
		case mbcFlag < 0x17:
			log.Println("Warning: MBC4 carts are not supported.")
			//cartridge.BankingController = NewMBC1(rom)
			//cartType = "MBC4"
		case mbcFlag < 0x1F:
			//cartridge.BankingController = NewMBC5(rom)
			//cartType = "MBC5"
		default:
			log.Printf("Warning: This cart may not be supported: %02x", mbcFlag)
			cartridge.BankingController = NewMBC1(rom)
		}
	}
	log.Printf("Cart type: %#02x (%v)", mbcFlag, cartType)
	log.Printf("Cart mode: %v", cartridge.mode)
	fmt.Scanln()

	// switch mbcFlag {
	// case 0x3, 0x6, 0x9, 0xD, 0xF, 0x10, 0x13, 0x17, 0x1B, 0x1E, 0xFF:
	// 	cartridge.initGameSaves()
	// }
	return &cartridge
}

func loadROMData(filename string) ([]byte, error) {
	var data []byte
	// if strings.HasSuffix(filename, ".zip") {
	// 	return loadZIPData(filename)
	// }
	// Load the file as a rom
	var err error
	data, err = os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}
