package gb

import (
	"fmt"
	"gameboy/bits"
)

// Rotate Left Circular
func (z *Z80) RLC(r byte) byte {
	bit7 := r >> 7      // Pega o bit 7
	r = (r << 1) | bit7 // Rotaciona e insere o bit 7 no bit 0

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit7 != 0
	z.setFlags()
	return r
}

// Rotate right Circular
func (z *Z80) RRC(r byte) byte {
	bit0 := r & 1              // Pega o bit 0
	r = (r >> 1) | (bit0 << 7) // Rotaciona e insere o bit 0 no bit 7

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit0 != 0
	z.setFlags()

	return r
}

// Rotate Left through Carry (RL)
func (z *Z80) RL(r byte) byte {
	bit7 := r >> 7              // Pega o bit 7
	r = (r << 1) | bits.B(z.CF) // Insere carry no bit 0

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit7 != 0
	z.setFlags()

	return r
}

// Rotate Right through Carry (RR)
func (z *Z80) RR(r byte) byte {
	bit0 := r & 1                      // Pega o bit 0
	r = (r >> 1) | (bits.B(z.CF) << 7) // Insere carry no bit 7

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit0 == 1
	z.setFlags()

	return r
}

// Shift Left Arithmetic (SLA)
func (z *Z80) SLA(r byte) byte {
	bit7 := r >> 7 // Pega o bit 7
	r = (r << 1)   // Desloca para a esquerda, zerando bit 0

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit7 != 0
	z.setFlags()

	return r
}

// Shift Right Arithmetic (SRA)
func (z *Z80) SRA(r byte) byte {
	bit0 := r & 1       // Pega o bit 0
	bit7 := r & 0x80    // Mantém o bit 7 (sinal)
	r = (r >> 1) | bit7 // Desloca para a direita, mantendo o bit 7

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit0 != 0
	z.setFlags()

	return r
}

// Swap Nibbles (SWAP)
func (z *Z80) SWAP(r byte) byte {
	r = (r >> 4) | (r << 4) // Troca os 4 bits superiores pelos inferiores

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = false // SWAP sempre limpa o Carry Flag
	z.setFlags()

	return r
}

// Shift Right Logical (SRL)
func (z *Z80) SRL(r byte) byte {
	bit0 := r & 1 // Pega o bit 0 (para o Carry)
	r = r >> 1    // Desloca todos os bits para a direita

	z.Z = r == 0
	z.N = false
	z.HF = false
	z.CF = bit0 != 0
	z.setFlags()

	return r
}

// Bit Test (BIT)
func (z *Z80) BIT(bit uint, r byte) {
	z.Z = (r & (1 << bit)) == 0 // Define Z se o bit estiver limpo (0)
	z.N = false
	z.HF = true
	z.setFlags()
}

// Reset Bit (RES)
func (z *Z80) RES(bit uint, r byte) byte {
	return r &^ (1 << bit) // Reseta o bit escolhido para 0
}

// Set Bit (SET)
func (z *Z80) SET(bit uint, r byte) byte {
	return r | (1 << bit) // Seta o bit escolhido para 1
}

func (z *Z80) ExecuteCBInstruction() {
	opcodeCB := z.readMemory(z.PC + 1)

	switch opcodeCB {
	case 0x00:
		// RLC B
		z.B = z.RLC(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x01:
		// RLC C
		z.C = z.RLC(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x02:
		// RLC D
		z.D = z.RLC(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x03:
		// RLC E
		z.E = z.RLC(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x04:
		// RLC H
		z.H = z.RLC(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x05:
		// RLC L
		z.L = z.RLC(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x07:
		// RLC A
		z.A = z.RLC(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x06: // RLC (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RLC(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x08:
		// RRC B
		z.B = z.RRC(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x09:
		// RRC C
		z.C = z.RRC(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x0A:
		// RRC D
		z.D = z.RRC(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x0B:
		// RRC E
		z.E = z.RRC(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x0C:
		// RRC H
		z.H = z.RRC(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x0D:
		// RRC L
		z.L = z.RRC(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x0E: // RRC (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RRC(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x0F:
		// RRC A
		z.A = z.RRC(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x10:
		// RL B
		z.B = z.RL(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x11:
		// RL C
		z.C = z.RL(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x12:
		// RL D
		z.D = z.RL(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x13:
		// RL E
		z.E = z.RL(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x14:
		// RL H
		z.H = z.RL(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x15:
		// RL L
		z.L = z.RL(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x16: // RL (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RL(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x17:
		// RL A
		z.A = z.RL(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x18:
		// RR B
		z.B = z.RR(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x19:
		//RR C
		z.C = z.RR(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x1A:
		//RR D
		z.D = z.RR(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x1B:
		// RR E
		z.E = z.RR(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x1C:
		// RR H
		z.H = z.RR(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x1D:
		// RR L
		z.L = z.RR(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x1E: // RR (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RR(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x1F:
		// RR A
		z.A = z.RR(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x20:
		// SLA B
		z.B = z.SLA(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x21:
		// SLA C
		z.C = z.SLA(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x22:
		// SLA D
		z.D = z.SLA(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x23:
		// SLA E
		z.E = z.SLA(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x24:
		// SLA H
		z.H = z.SLA(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x25:
		// SLA L
		z.L = z.SLA(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x26: // SLA (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SLA(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x27:
		// SLA A
		z.A = z.SLA(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x28:
		// SRA B
		z.B = z.SRA(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x29:
		// SRA C
		z.C = z.SRA(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x2A:
		// SRA D
		z.D = z.SRA(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x2B:
		// SRA E
		z.E = z.SRA(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x2C:
		// SRA H
		z.H = z.SRA(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x2D:
		// SRA L
		z.L = z.SRA(z.L)

		z.PC += 2
		z.M = 8
	case 0x2E: // SRA (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SRA(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x2F:
		// SRA A
		z.A = z.SRA(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x30:
		// SWAP B
		z.B = z.SWAP(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x31:
		// SWAP C
		z.C = z.SWAP(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x32:
		// SWAP D
		z.D = z.SWAP(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x33:
		// SWAP E
		z.E = z.SWAP(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x34:
		// SWAP H
		z.H = z.SWAP(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x35:
		// SWAP L
		z.L = z.SWAP(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x36: // SWAP (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SWAP(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x37:
		// SWAP A
		z.A = z.SWAP(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x38:
		//SRL B
		z.B = z.SRL(z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x39:
		//SRL C
		z.C = z.SRL(z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x3A:
		//SRL D
		z.D = z.SRL(z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x3B:
		//SRL E
		z.E = z.SRL(z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x3C:
		//SRL H
		z.H = z.SRL(z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x3D:
		//SRL L
		z.L = z.SRL(z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x3E: // SRL (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SRL(value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x3F:
		//SRL A
		z.A = z.SRL(z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x40:
		// BIT 0, B
		z.BIT(0, z.B)

		z.PC += 2
		z.M = 8
	case 0x41:
		// BIT 0, C
		z.BIT(0, z.C)

		z.PC += 2
		z.M = 8
	case 0x42:
		// BIT 0, D
		z.BIT(0, z.D)

		z.PC += 2
		z.M = 8
	case 0x43:
		// BIT 0, E
		z.BIT(0, z.E)

		z.PC += 2
		z.M = 8
	case 0x44:
		// BIT 0, H
		z.BIT(0, z.H)

		z.PC += 2
		z.M = 8
	case 0x45:
		// BIT 0, L
		z.BIT(0, z.L)

		z.PC += 2
		z.M = 8
	case 0x46:
		// BIT 0, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(0, value)

		z.PC += 2
		z.M = 12
	case 0x47:
		// BIT 0, A
		z.BIT(0, z.A)

		z.PC += 2
		z.M = 8
	case 0x48:
		// BIT 1, B
		z.BIT(1, z.B)

		z.PC += 2
		z.M = 8
	case 0x49:
		// BIT 1, C
		z.BIT(1, z.C)

		z.PC += 2
		z.M = 8
	case 0x4A:
		// BIT 1, D
		z.BIT(1, z.D)

		z.PC += 2
		z.M = 8
	case 0x4B:
		// BIT 1, E
		z.BIT(1, z.E)

		z.PC += 2
		z.M = 8
	case 0x4C:
		// BIT 1, H
		z.BIT(1, z.H)

		z.PC += 2
		z.M = 8
	case 0x4D:
		// BIT 1, L
		z.BIT(1, z.L)

		z.PC += 2
		z.M = 8
	case 0x4E:
		// BIT 1, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(1, value)

		z.PC += 2
		z.M = 12
	case 0x4F:
		// BIT 1, A
		z.BIT(1, z.A)

		z.PC += 2
		z.M = 8
	case 0x50:
		// BIT 2, B
		z.BIT(2, z.B)

		z.PC += 2
		z.M = 8
	case 0x51:
		// BIT 2, C
		z.BIT(2, z.C)

		z.PC += 2
		z.M = 8
	case 0x52:
		// BIT 2, D
		z.BIT(2, z.D)

		z.PC += 2
		z.M = 8
	case 0x53:
		// BIT 2, E
		z.BIT(2, z.E)

		z.PC += 2
		z.M = 8
	case 0x54:
		// BIT 2, H
		z.BIT(2, z.H)

		z.PC += 2
		z.M = 8
	case 0x55:
		// BIT 2, L
		z.BIT(2, z.L)

		z.PC += 2
		z.M = 8
	case 0x56:
		// BIT 2, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(2, value)

		z.PC += 2
		z.M = 12
	case 0x57:
		// BIT 2, A
		z.BIT(2, z.A)

		z.PC += 2
		z.M = 8
	case 0x58:
		// BIT 3, B
		z.BIT(3, z.B)

		z.PC += 2
		z.M = 8
	case 0x59:
		// BIT 3, C
		z.BIT(3, z.C)

		z.PC += 2
		z.M = 8
	case 0x5A:
		// BIT 3, D
		z.BIT(3, z.D)

		z.PC += 2
		z.M = 8
	case 0x5B:
		// BIT 3, E
		z.BIT(3, z.E)

		z.PC += 2
		z.M = 8
	case 0x5C:
		// BIT 3, H
		z.BIT(3, z.H)

		z.PC += 2
		z.M = 8
	case 0x5D:
		// BIT 3, L
		z.BIT(3, z.L)

		z.PC += 2
		z.M = 8
	case 0x5E:
		// BIT 3, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(3, value)

		z.PC += 2
		z.M = 12
	case 0x5F:
		// BIT 3, A
		z.BIT(3, z.A)

		z.PC += 2
		z.M = 8
	case 0x60:
		// BIT 4, B
		z.BIT(4, z.B)

		z.PC += 2
		z.M = 8
	case 0x61:
		// BIT 4, C
		z.BIT(4, z.C)

		z.PC += 2
		z.M = 8
	case 0x62:
		// BIT 4, D
		z.BIT(4, z.D)

		z.PC += 2
		z.M = 8
	case 0x63:
		// BIT 4, E
		z.BIT(4, z.E)

		z.PC += 2
		z.M = 8
	case 0x64:
		// BIT 4, H
		z.BIT(4, z.H)

		z.PC += 2
		z.M = 8
	case 0x65:
		// BIT 4, L
		z.BIT(4, z.L)

		z.PC += 2
		z.M = 8
	case 0x66:
		// BIT 4, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(4, value)

		z.PC += 2
		z.M = 12
	case 0x67:
		// BIT 4, A
		z.BIT(4, z.A)

		z.PC += 2
		z.M = 8
	case 0x68:
		// BIT 5, B
		z.BIT(5, z.B)

		z.PC += 2
		z.M = 8
	case 0x69:
		// BIT 5, C
		z.BIT(5, z.C)

		z.PC += 2
		z.M = 8
	case 0x6A:
		// BIT 5, D
		z.BIT(5, z.D)

		z.PC += 2
		z.M = 8
	case 0x6B:
		// BIT 5, E
		z.BIT(5, z.E)

		z.PC += 2
		z.M = 8
	case 0x6C:
		// BIT 5, H
		z.BIT(5, z.H)

		z.PC += 2
		z.M = 8
	case 0x6D:
		// BIT 5, L
		z.BIT(5, z.L)

		z.PC += 2
		z.M = 8
	case 0x6E:
		// BIT 5, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(5, value)

		z.PC += 2
		z.M = 12
	case 0x6F:
		// BIT 5, A
		z.BIT(5, z.A)

		z.PC += 2
		z.M = 8
	case 0x70:
		// BIT 6, B
		z.BIT(6, z.B)

		z.PC += 2
		z.M = 8
	case 0x71:
		// BIT 6, C
		z.BIT(6, z.C)

		z.PC += 2
		z.M = 8
	case 0x72:
		// BIT 6, D
		z.BIT(6, z.D)

		z.PC += 2
		z.M = 8
	case 0x73:
		// BIT 6, E
		z.BIT(6, z.E)

		z.PC += 2
		z.M = 8
	case 0x74:
		// BIT 6, H
		z.BIT(6, z.H)

		z.PC += 2
		z.M = 8
	case 0x75:
		// BIT 6, L
		z.BIT(6, z.L)

		z.PC += 2
		z.M = 8
	case 0x76:
		// BIT 6, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(6, value)

		z.PC += 2
		z.M = 12
	case 0x77:
		// BIT 6, A
		z.BIT(6, z.A)

		z.PC += 2
		z.M = 8
	case 0x78:
		// BIT 7, B
		z.BIT(7, z.B)

		z.PC += 2
		z.M = 8
	case 0x79:
		// BIT 7, C
		z.BIT(7, z.C)

		z.PC += 2
		z.M = 8
	case 0x7A:
		// BIT 7, D
		z.BIT(7, z.D)

		z.PC += 2
		z.M = 8
	case 0x7B:
		// BIT 7, E
		z.BIT(7, z.E)

		z.PC += 2
		z.M = 8
	case 0x7C:
		// BIT 7, H
		z.BIT(7, z.H)

		z.PC += 2
		z.M = 8
	case 0x7D:
		// BIT 7, L
		z.BIT(7, z.L)

		z.PC += 2
		z.M = 8
	case 0x7E:
		// BIT 7, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		z.BIT(7, value)

		z.PC += 2
		z.M = 12
	case 0x7F:
		// BIT 7, A
		z.BIT(7, z.A)

		z.PC += 2
		z.M = 8
	case 0x80:
		// RES 0, B
		z.B = z.RES(0, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x81:
		// RES 0, C
		z.C = z.RES(0, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x82:
		// RES 0, D
		z.D = z.RES(0, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x83:
		// RES 0, E
		z.E = z.RES(0, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x84:
		// RES 0, H
		z.H = z.RES(0, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x85:
		// RES 0, L
		z.L = z.RES(0, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x86:
		// RES 0, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(0, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	// case 0x86:
	//     // RES 0, (HL)
	//     address := z.HL
	//     value := z.Memory.ReadByte(address)
	//     value &^= (1 << 0) // Clear bit 0
	//     z.Memory.WriteByte(address, value)
	//     z.M = 16
	case 0x87:
		// RES 0, A
		z.A = z.RES(0, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x88:
		// RES 1, B
		z.B = z.RES(1, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x89:
		// RES 1, C
		z.C = z.RES(1, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x8A:
		// RES 1, D
		z.D = z.RES(1, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x8B:
		// RES 1, E
		z.E = z.RES(1, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x8C:
		// RES 1, H
		z.H = z.RES(1, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x8D:
		// RES 1, L
		z.L = z.RES(1, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x8E:
		// RES 1, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(1, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x8F:
		// RES 1, A
		z.A = z.RES(1, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x90:
		// RES 2, B
		z.B = z.RES(2, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x91:
		// RES 2, C
		z.C = z.RES(2, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x92:
		// RES 2, D
		z.D = z.RES(2, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x93:
		// RES 2, E
		z.E = z.RES(2, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x94:
		// RES 2, H
		z.H = z.RES(2, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x95:
		// RES 2, L
		z.L = z.RES(2, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x96:
		// RES 2, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(2, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x97:
		// RES 2, A
		z.A = z.RES(2, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0x98:
		// RES 3, B
		z.B = z.RES(3, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x99:
		// RES 3, C
		z.C = z.RES(3, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0x9A:
		// RES 3, D
		z.D = z.RES(3, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x9B:
		// RES 3, E
		z.E = z.RES(3, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0x9C:
		// RES 3, H
		z.H = z.RES(3, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x9D:
		// RES 3, L
		z.L = z.RES(3, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0x9E:
		// RES 3, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(3, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0x9F:
		// RES 3, A
		z.A = z.RES(3, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xA0:
		// RES 4, B
		z.B = z.RES(4, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xA1:
		// RES 4, C
		z.C = z.RES(4, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xA2:
		// RES 4, D
		z.D = z.RES(4, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xA3:
		// RES 4, E
		z.E = z.RES(4, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xA4:
		// RES 4, H
		z.H = z.RES(4, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xA5:
		// RES 4, L
		z.L = z.RES(4, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xA6:
		// RES 4, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(4, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xA7:
		// RES 4, A
		z.A = z.RES(4, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xA8:
		// RES 5, B
		z.B = z.RES(5, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xA9:
		// RES 5, C
		z.C = z.RES(5, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xAA:
		// RES 5, D
		z.D = z.RES(5, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xAB:
		// RES 5, E
		z.E = z.RES(5, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xAC:
		// RES 5, H
		z.H = z.RES(5, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xAD:
		// RES 5, L
		z.L = z.RES(5, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xAE:
		// RES 5, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(5, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xAF:
		// RES 5, A
		z.A = z.RES(5, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xB0:
		// RES 6, B
		z.B = z.RES(6, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xB1:
		// RES 6, C
		z.C = z.RES(6, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xB2:
		// RES 6, D
		z.D = z.RES(6, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xB3:
		// RES 6, E
		z.E = z.RES(6, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xB4:
		// RES 6, H
		z.H = z.RES(6, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xB5:
		// RES 6, L
		z.L = z.RES(6, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xB6:
		// RES 6, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(6, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	// case 0xB6:
	// 	// RES 6, (HL)
	// 	address := z.HL
	// 	value := z.Memory.ReadByte(address)
	// 	value &^= (1 << 6) // Clear bit 6
	// 	z.Memory.WriteByte(address, value)

	// 	z.PC += 2
	// 	z.M = 16
	case 0xB7:
		// RES 6, A
		z.A = z.RES(6, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xB8:
		// RES 7, B
		z.B = z.RES(7, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xB9:
		// RES 7, C
		z.C = z.RES(7, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xBA:
		// RES 7, D
		z.D = z.RES(7, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xBB:
		// RES 7, E
		z.E = z.RES(7, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xBC:
		// RES 7, H
		z.H = z.RES(7, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xBD:
		// RES 7, L
		z.L = z.RES(7, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xBE:
		// RES 7, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.RES(7, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xBF:
		// RES 7, A
		z.A = z.RES(7, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xC0:
		// SET 0, B
		z.B = z.SET(0, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xC1:
		// SET 0, C
		z.C = z.SET(0, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xC2:
		// SET 0, D
		z.D = z.SET(0, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xC3:
		// SET 0, E
		z.E = z.SET(0, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xC4:
		// SET 0, H
		z.H = z.SET(0, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xC5:
		// SET 0, L
		z.L = z.SET(0, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xC6:
		// SET 0, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(0, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xC7:
		// SET 0, A
		z.A = z.SET(0, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xC8:
		// SET 1, B
		z.B = z.SET(1, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xC9:
		// SET 1, C
		z.C = z.SET(1, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xCA:
		// SET 1, D
		z.D = z.SET(1, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xCB:
		// SET 1, E
		z.E = z.SET(1, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xCC:
		// SET 1, H
		z.H = z.SET(1, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xCD:
		// SET 1, L
		z.L = z.SET(1, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xCE:
		// SET 1, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(1, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xCF:
		// SET 1, A
		z.A = z.SET(1, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xD0:
		// SET 2, B
		z.B = z.SET(2, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xD1:
		// SET 2, C
		z.C = z.SET(2, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xD2:
		// SET 2, D
		z.D = z.SET(2, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xD3:
		// SET 2, E
		z.E = z.SET(2, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xD4:
		// SET 2, H
		z.H = z.SET(2, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xD5:
		// SET 2, L
		z.L = z.SET(2, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xD6:
		// SET 2, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(2, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xD7:
		// SET 2, A
		z.A = z.SET(2, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xD8:
		// SET 3, B
		z.B = z.SET(3, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xD9:
		// SET 3, C
		z.C = z.SET(3, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xDA:
		// SET 3, D
		z.D = z.SET(3, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xDB:
		// SET 3, E
		z.E = z.SET(3, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xDC:
		// SET 3, H
		z.H = z.SET(3, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xDD:
		// SET 3, L
		z.L = z.SET(3, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xDE:
		// SET 3, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(3, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xDF:
		// SET 3, A
		z.A = z.SET(3, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xE0:
		// SET 4, B
		z.B = z.SET(4, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xE1:
		// SET 4, C
		z.C = z.SET(4, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xE2:
		// SET 4, D
		z.D = z.SET(4, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xE3:
		// SET 4, E
		z.E = z.SET(4, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xE4:
		// SET 4, H
		z.H = z.SET(4, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xE5:
		// SET 4, L
		z.L = z.SET(4, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xE6:
		// SET 4, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(4, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xE7:
		// SET 4, A
		z.A = z.SET(4, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xE8:
		// SET 5, B
		z.B = z.SET(5, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xE9:
		// SET 5, C
		z.C = z.SET(5, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xEA:
		// SET 5, D
		z.D = z.SET(5, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xEB:
		// SET 5, E
		z.E = z.SET(5, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xEC:
		// SET 5, H
		z.H = z.SET(5, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xED:
		// SET 5, L
		z.L = z.SET(5, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xEE:
		// SET 5, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(5, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xEF:
		// SET 5, A
		z.A = z.SET(5, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xF0:
		// SET 6, B
		z.B = z.SET(6, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xF1:
		// SET 6, C
		z.C = z.SET(6, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xF2:
		// SET 6, D
		z.D = z.SET(6, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xF3:
		// SET 6, E
		z.E = z.SET(6, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xF4:
		// SET 6, H
		z.H = z.SET(6, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xF5:
		// SET 6, L
		z.L = z.SET(6, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xF6:
		// SET 6, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(6, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xF7:
		// SET 6, A
		z.A = z.SET(6, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	case 0xF8:
		// SET 7, B
		z.B = z.SET(7, z.B)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xF9:
		// SET 7, C
		z.C = z.SET(7, z.C)
		z.setBC()

		z.PC += 2
		z.M = 8
	case 0xFA:
		// SET 7, D
		z.D = z.SET(7, z.D)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xFB:
		// SET 7, E
		z.E = z.SET(7, z.E)
		z.setDE()

		z.PC += 2
		z.M = 8
	case 0xFC:
		// SET 7, H
		z.H = z.SET(7, z.H)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xFD:
		// SET 7, L
		z.L = z.SET(7, z.L)
		z.setHL()

		z.PC += 2
		z.M = 8
	case 0xFE:
		// SET 7, (HL)
		address := z.HL
		value := z.Memory.ReadByte(address)
		newValue := z.SET(7, value)
		z.Memory.WriteByte(address, newValue)

		z.PC += 2
		z.M = 16
	case 0xFF:
		// SET 7, A
		z.A = z.SET(7, z.A)
		z.setAF()

		z.PC += 2
		z.M = 8
	default:
		fmt.Printf("Opcode CB não suportado: 0xCB%x\n", opcodeCB)
		fmt.Scanln()
		z.M = 4
	}
}
