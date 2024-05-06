package cart

type ROM struct {
	rom []byte
}

func NewROM(data []byte) BankingController {
	return &ROM{
		rom: data,
	}
}

func (r *ROM) Read(address uint16) byte {
	return r.rom[address]
}

func (r *ROM) WriteROM(address uint16, value byte) {}

func (r *ROM) WriteRAM(address uint16, value byte) {}

func (r *ROM) GetSaveData() []byte {
	return []byte{}
}

// LoadSaveData loads the save data into the cartridge. As RAM is not supported
// on this memory controller, this is a noop.
func (r *ROM) LoadSaveData([]byte) {}
