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