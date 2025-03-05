package gb

import (
	"fmt"
	"gameboy/bits"
	"gameboy/logger"
	_ "sort"
)

type Z80 struct {
	A, B, C, D, E, H, L, F byte   // Registradores de 8 bits
	AF, BC, DE, HL         uint16 // Pares de registradores de 16 bits
	PC, SP                 uint16 // Contador de programa (PC) e ponteiro de pilha (SP)

	Z, N, HF, CF bool // Flags de status (Zero, Negativo, Meio-carro, Carry, Interrupt Master Enable)

	IME, InterruptsEnabling bool

	M int // Contador de ciclos de máquina

	Divider int

	Memory *Memory
}

func (z *Z80) SpHiLo() uint16 {
	return z.SP
}

func (z *Z80) setFlag(index byte, on bool) {
	if on {
		z.F = bits.Set(z.F, index)
		z.setAF()
	} else {
		z.F = z.F & ^(1 << index)
		z.setAF()
	}
}

// SetZ sets the value of the Z flag.
func (z *Z80) setFlags() {
	z.setFlag(7, z.Z)  //Z
	z.setFlag(6, z.N)  //N
	z.setFlag(5, z.HF) //HF
	z.setFlag(4, z.CF) //CF
}

func (z *Z80) setAF() {
	z.AF = uint16(z.A)<<8 | uint16(z.F)
}

func (z *Z80) setBC() {
	z.BC = uint16(z.B)<<8 | uint16(z.C)
}

func (z *Z80) setDE() {
	z.DE = uint16(z.D)<<8 | uint16(z.E)
}

func (z *Z80) setHL() {
	z.HL = uint16(z.H)<<8 | uint16(z.L)
}

func (z *Z80) setFlagsFromF() {
	z.Z = (z.F & 0x80) != 0  // Bit 7: Zero flag
	z.N = (z.F & 0x40) != 0  // Bit 6: Subtract flag
	z.HF = (z.F & 0x20) != 0 // Bit 5: Half Carry flag
	z.CF = (z.F & 0x10) != 0 // Bit 4: Carry flag

	z.setAF()
}

func (z *Z80) Init(memory *Memory, cgb bool) {
	z.PC = 0x0100
	z.SP = 0xFFFE

	if cgb {
		z.A = 0x11
		z.F = 0x80
		z.B = 0x00
		z.C = 0x00
		z.D = 0xFF
		z.E = 0x56
		z.H = 0x00
		z.L = 0x0D
	} else {
		z.A = 0x01
		z.F = 0xB0
		z.B = 0x00
		z.C = 0x13
		z.D = 0x00
		z.E = 0xD8
		z.H = 0x01
		z.L = 0x4D
	}

	z.setFlagsFromF()
	z.setBC()
	z.setDE()
	z.setHL()
	z.Memory = memory

}

func (z *Z80) readMemory(addr uint16) byte {
	if z.Memory == nil {
		fmt.Println("Erro: Memória não inicializada")
		return 0xFF // Retornar valor padrão em caso de memória não inicializada
	}

	return z.Memory.ReadByte(addr)
}

func (z *Z80) readHighRam(addr uint16) byte {
	return z.Memory.ReadHighRam(addr)
}

func (z *Z80) updateFlagsInc(value byte) {
	z.Z = value == 0
	z.N = false
	z.HF = (value & 0x0F) == 0
	z.setFlags()
}

func (z *Z80) updateFlagsDec(reg byte) {
	z.Z = reg == 0
	z.N = true
	z.HF = (reg & 0x0F) == 0x0F
	z.setFlags()
}

func (z *Z80) updateFlagsCP(reg byte) {
	result := uint16(z.A) - uint16(reg)

	z.Z = result == 0
	z.N = true
	z.HF = (z.A & 0x0F) < (reg & 0x0F)
	z.CF = reg > z.A

	z.setFlags()
}

func (z *Z80) updateFlagsAdd(reg byte, value byte) {
	result := uint16(reg) + uint16(value)

	z.Z = byte(result) == 0
	z.N = false
	z.HF = ((int16(reg)&0x0F)+(int16(value)&0xF) > 0xF)
	z.CF = result > 0xFF

	z.A = byte(result)
	z.setFlags()
}

func (z *Z80) updateFlagsAdc(reg byte, value byte) {
	carry := uint16(0)
	if z.CF {
		carry = 1
	}
	result := uint16(reg) + uint16(value) + carry

	z.Z = byte(result) == 0
	z.N = false
	z.HF = ((z.A & 0xF) + (value & 0x0F) + byte(carry)) > 0x0F
	z.CF = result > 0xFF
	z.A = byte(result)
	z.setFlags()
}

func (z *Z80) updateFlagsSub(reg byte, value byte) {
	result := uint16(reg) - uint16(value)

	z.Z = byte(result) == 0
	z.N = true
	z.HF = ((int16(reg)&0x0F)-(int16(value)&0xF) < 0)
	z.CF = result > 0xFF

	z.A = byte(result)
	z.setFlags()
}

func (z *Z80) updateFlagsSbc(reg byte, value byte) {
	carry := uint16(0)
	if z.CF {
		carry = 1
	}
	result := uint16(reg) - uint16(value) - carry

	z.Z = byte(result) == 0
	z.N = true
	z.HF = ((int16(reg)&0x0F)-(int16(value)&0xF)-int16(carry) < 0)
	z.CF = result > 0xFF

	z.A = byte(result)
	z.setFlags()
}

func (z *Z80) updateFlagsAnd(reg byte) {
	z.A &= reg

	z.Z = z.A == 0
	z.N = false
	z.HF = true
	z.CF = false
	z.setFlags()
}

func (z *Z80) updateFlagsOr(reg byte) {
	z.A |= reg

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()
}

func (z *Z80) updateFlagsXor(reg byte) {
	z.A ^= reg

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()
}

func (z *Z80) ExecuteInstruction(opcode byte) {
	switch opcode {
	case 0x00:
		z.NOP()
	case 0x01:
		z.LD_BC_nn()
	case 0x02:
		z.LD_BC_addr_A()
	case 0x03:
		z.INC_BC()
	case 0x04:
		z.INC_B()
	case 0x05:
		z.DEC_B()
	case 0x06:
		z.LD_B_d8()
	case 0x07:
		z.RLCA()
	case 0x08:
		z.LD_a16_SP()
	case 0x09:
		z.ADD_HL_BC()
	case 0x0A:
		z.LD_A_BC_addr()
	case 0x0B:
		z.DEC_BC()
	case 0x0C:
		z.INC_C()
	case 0x0D:
		z.DEC_C()
	case 0x0E:
		z.LD_C_d8()
	case 0x0F:
		z.RRCA()
	case 0x10:
		z.STOP()
	case 0x11:
		z.LD_DE_nn()
	case 0x12:
		z.LD_DE_addr_A()
	case 0x13:
		z.INC_DE()
	case 0x14:
		z.INC_D()
	case 0x15:
		z.DEC_D()
	case 0x16:
		z.LD_D_d8()
	case 0x17:
		z.RLA()
	case 0x18:
		z.JR_e()
	case 0x19:
		z.ADD_HL_DE()
	case 0x1A:
		z.LD_A_DE_addr()
	case 0x1B:
		z.DEC_DE()
	case 0x1C:
		z.INC_E()
	case 0x1D:
		z.DEC_E()
	case 0x1E:
		z.LD_E_d8()
	case 0x1F:
		z.RRA()
	case 0x20:
		z.JR_nz_e()
	case 0x21:
		z.LD_HL_nn()
	case 0x22:
		z.LD_HL_inc_A()
	case 0x23:
		z.INC_HL()
	case 0x24:
		z.INC_H()
	case 0x25:
		z.DEC_H()
	case 0x26:
		z.LD_H_d8()
	case 0x27:
		z.DAA()
	case 0x28:
		z.JR_Z_e()
	case 0x29:
		z.ADD_HL_HL()
	case 0x2A:
		z.LDI_A_HL()
	case 0x2B:
		z.DEC_HL()
	case 0x2C:
		z.INC_L()
	case 0x2D:
		z.DEC_L()
	case 0x2E:
		z.LD_L_d8()
	case 0x2F:
		z.CPL()
	case 0x30:
		z.JR_NC_r8()
	case 0x31:
		z.LD_SP_nn()
	case 0x32:
		z.LD_HL_dec_A()
	case 0x33:
		z.INC_SP()
	case 0x34:
		z.INC_HL_addr()
	case 0x35:
		z.DEC_HL_addr()
	case 0x36:
		z.LD_HL_addr_d8()
	case 0x37:
		z.SCF()
	case 0x38:
		z.JR_C_r8()
	case 0x39:
		z.ADD_HL_SP()
	case 0x3A:
		z.LDD_A_HL()
	case 0x3B:
		z.DEC_SP()
	case 0x3C:
		z.INC_A()
	case 0x3D:
		z.DEC_A()
	case 0x3E:
		z.LD_A_d8()
	case 0x3F:
		z.CCF()
	case 0x40:
		z.LD_B_B()
	case 0x41:
		z.LD_B_C()
	case 0x42:
		z.LD_B_D()
	case 0x43:
		z.LD_B_E()
	case 0x44:
		z.LD_B_H()
	case 0x45:
		z.LD_B_L()
	case 0x46:
		z.LD_B_HL_addr()
	case 0x47:
		z.LD_B_A()
	case 0x48:
		z.LD_C_B()
	case 0x49:
		z.LD_C_C()
	case 0x4A:
		z.LD_C_D()
	case 0x4B:
		z.LD_C_E()
	case 0x4C:
		z.LD_C_H()
	case 0x4D:
		z.LD_C_L()
	case 0x4E:
		z.LD_C_HL_addr()
	case 0x4F:
		z.LD_C_A()
	case 0x50:
		z.LD_D_B()
	case 0x51:
		z.LD_D_C()
	case 0x52:
		z.LD_D_D()
	case 0x53:
		z.LD_D_E()
	case 0x54:
		z.LD_D_H()
	case 0x55:
		z.LD_D_L()
	case 0x56:
		z.LD_D_HL_addr()
	case 0x57:
		z.LD_D_A()
	case 0x58:
		z.LD_E_B()
	case 0x59:
		z.LD_E_C()
	case 0x5A:
		z.LD_E_D()
	case 0x5B:
		z.LD_E_E()
	case 0x5C:
		z.LD_E_H()
	case 0x5D:
		z.LD_E_L()
	case 0x5E:
		z.LD_E_HL_addr()
	case 0x5F:
		z.LD_E_A()
	case 0x60:
		z.LD_H_B()
	case 0x61:
		z.LD_H_C()
	case 0x62:
		z.LD_H_D()
	case 0x63:
		z.LD_H_E()
	case 0x64:
		z.LD_H_H()
	case 0x65:
		z.LD_H_L()
	case 0x66:
		z.LD_H_HL_addr()
	case 0x67:
		z.LD_H_A()
	case 0x68:
		z.LD_L_B()
	case 0x69:
		z.LD_L_C()
	case 0x6A:
		z.LD_L_D()
	case 0x6B:
		z.LD_L_E()
	case 0x6C:
		z.LD_L_H()
	case 0x6D:
		z.LD_L_L()
	case 0x6E:
		z.LD_L_HL_addr()
	case 0x6F:
		z.LD_L_A()
	case 0x70:
		z.LD_HL_addr_B()
	case 0x71:
		z.LD_HL_addr_C()
	case 0x72:
		z.LD_HL_addr_D()
	case 0x73:
		z.LD_HL_addr_E()
	case 0x74:
		z.LD_HL_addr_H()
	case 0x75:
		z.LD_HL_addr_L()
	case 0x76:
		z.HALT()
	case 0x77:
		z.LD_HL_addr_A()
	case 0x78:
		z.LD_A_B()
	case 0x79:
		z.LD_A_C()
	case 0x7A:
		z.LD_A_D()
	case 0x7B:
		z.LD_A_E()
	case 0x7C:
		z.LD_A_H()
	case 0x7D:
		z.LD_A_L()
	case 0x7E:
		z.LD_A_HL_addr()
	case 0x7F:
		z.LD_A_A()
	case 0x80:
		z.ADD_A_B()
	case 0x81:
		z.ADD_A_C()
	case 0x82:
		z.ADD_A_D()
	case 0x83:
		z.ADD_A_E()
	case 0x84:
		z.ADD_A_H()
	case 0x85:
		z.ADD_A_L()
	case 0x86:
		z.ADD_A_HL_addr()
	case 0x87:
		z.ADD_A_A()
	case 0x88:
		z.ADC_A_B()
	case 0x89:
		z.ADC_A_C()
	case 0x8A:
		z.ADC_A_D()
	case 0x8B:
		z.ADC_A_E()
	case 0x8C:
		z.ADC_A_H()
	case 0x8D:
		z.ADC_A_L()
	case 0x8E:
		z.ADC_A_HL_addr()
	case 0x8F:
		z.ADC_A_A()
	case 0x90:
		z.SUB_B()
	case 0x91:
		z.SUB_C()
	case 0x92:
		z.SUB_D()
	case 0x93:
		z.SUB_E()
	case 0x94:
		z.SUB_H()
	case 0x95:
		z.SUB_L()
	case 0x96:
		z.SUB_HL_addr()
	case 0x97:
		z.SUB_A()
	case 0x98:
		z.SBC_A_B()
	case 0x99:
		z.SBC_A_C()
	case 0x9A:
		z.SBC_A_D()
	case 0x9B:
		z.SBC_A_E()
	case 0x9C:
		z.SBC_A_H()
	case 0x9D:
		z.SBC_A_L()
	case 0x9E:
		z.SBC_A_HL_addr()
	case 0x9F:
		z.SBC_A_A()
	case 0xA0:
		z.AND_B()
	case 0xA1:
		z.AND_C()
	case 0xA2:
		z.AND_D()
	case 0xA3:
		z.AND_E()
	case 0xA4:
		z.AND_H()
	case 0xA5:
		z.AND_L()
	case 0xA6:
		z.AND_HL_addr()
	case 0xA7:
		z.AND_A()
	case 0xA8:
		z.XOR_B()
	case 0xA9:
		z.XOR_C()
	case 0xAA:
		z.XOR_D()
	case 0xAB:
		z.XOR_E()
	case 0xAC:
		z.XOR_H()
	case 0xAD:
		z.XOR_L()
	case 0xAE:
		z.XOR_HL_addr()
	case 0xAF:
		z.XOR_A()
	case 0xB0:
		z.OR_B()
	case 0xB1:
		z.OR_C()
	case 0xB2:
		z.OR_D()
	case 0xB3:
		z.OR_E()
	case 0xB4:
		z.OR_H()
	case 0xB5:
		z.OR_L()
	case 0xB6:
		z.OR_HL_addr()
	case 0xB7:
		z.OR_A()
	case 0xB8:
		z.CP_B()
	case 0xB9:
		z.CP_C()
	case 0xBA:
		z.CP_D()
	case 0xBB:
		z.CP_E()
	case 0xBC:
		z.CP_H()
	case 0xBD:
		z.CP_L()
	case 0xBE:
		z.CP_HL_addr()
	case 0xBF:
		z.CP_A()
	case 0xC0:
		z.RET_NZ()
	case 0xC1:
		z.POP_BC()
	case 0xC2:
		z.JP_NZ_nn()
	case 0xC3:
		z.JP_nn()
	case 0xC4:
		z.CALL_NZ_nn()
	case 0xC5:
		z.PUSH_BC()
	case 0xC6:
		z.ADD_A_d8()
	case 0xC7:
		z.RST_00H()
	case 0xC8:
		z.RET_Z()
	case 0xC9:
		z.RET()
	case 0xCA:
		z.JP_Z_nn()
	case 0xCB:
		z.ExecuteCBInstruction()
	case 0xCC:
		z.CALL_Z_nn()
	case 0xCD:
		z.CALL_nn()
	case 0xCE:
		z.ADC_A_d8()
	case 0xCF:
		z.RST_08H()
	case 0xD0:
		z.RET_NC()
	case 0xD1:
		z.POP_DE()
	case 0xD2:
		z.JP_NC_nn()
	case 0xD4:
		z.CALL_NC_nn()
	case 0xD5:
		z.PUSH_DE()
	case 0xD6:
		z.SUB_A_d8()
	case 0xD7:
		z.RST_10H()
	case 0xD8:
		z.RET_C()
	case 0xD9:
		z.RETI()
	case 0xDA:
		z.JP_C_nn()
	case 0xDC:
		z.CALL_C_nn()
	case 0xDE:
		z.SBC_A_d8()
	case 0xDF:
		z.RST_18H()
	case 0xE0:
		z.LDH_a8_A()
	case 0xE1:
		z.POP_HL()
	case 0xE2:
		z.LD_C_addr_A()
	case 0xE5:
		z.PUSH_HL()
	case 0xE6:
		z.AND_n()
	case 0xE7:
		z.RST_20H()
	case 0xE8:
		z.ADD_SP_r8()
	case 0xE9:
		z.JP_HL_addr()
	case 0xEA:
		z.LD_nn_A()
	case 0xEE:
		z.XOR_d8()
	case 0xEF:
		z.RST_28H()
	case 0xF0:
		z.LDH_A_a8()
	case 0xF1:
		z.POP_AF()
	case 0xF2:
		z.LD_A_C_addr()
	case 0xF3:
		z.DI()
	case 0xF5:
		z.PUSH_AF()
	case 0xF6:
		z.OR_d8()
	case 0xF7:
		z.RST_30H()
	case 0xF8:
		z.LD_HL_SP_r8()
	case 0xF9:
		z.LD_SP_HL()
	case 0xFA:
		z.LD_A_nn()
	case 0xFB:
		z.EI()
	case 0xFE:
		z.CP_n()
	case 0xFF:
		z.RST_38H()

	default:
		//logger.DisplayLogs()
		//panic(fmt.Sprintf("Opcode não suportado: 0x%X\n", opcode))
		// logger.LogMessage(fmt.Sprintf("Opcode não suportado: 0x%X\n", opcode))
		// logger.CloseLogger()
		// os.Exit(1) // Encerra o programa

		logger.LogMessage(fmt.Sprintf("Opcode não suportado: 0x%X\n", opcode))
		//fmt.Printf("Opcode não suportado: 0x%X\n", opcode)
		fmt.Scanln()
		// fmt.Printf("Opcode não suportado: 0x%X\n", opcode)

		// logger.LogMessage(fmt.Sprintf("Opcode não suportado: 0x%X\n", opcode))
		// fmt.Print(logger.GetRemainingLogs())
		// panic("Opcode não suportado")
		// time.Sleep(100 * time.Millisecond)

		// z.PC++
		// z.M = 4
	}
}

func (z *Z80) EmulateCycle() int {
	opcode := z.readMemory(z.PC)
	var cb byte
	if opcode == 0xCB {
		cb = z.readMemory(z.PC + 1)
	}
	// fmt.Println("\n================ Registros ================")
	// fmt.Printf("Z: %t | N: %t | HF: %t | CF: %t\n", z.Z, z.N, z.HF, z.CF)
	// fmt.Printf("PC: 0x%X | SP: 0x%X | LY:  0x%X\n", z.PC, z.SP, z.Memory.Hram[0x44])
	// fmt.Printf("LCDC: %X | STAT: %X | LYC: %X  IF: 0x%X\n", z.Memory.Hram[0x40], z.Memory.Hram[0x41], z.Memory.Hram[0x45], z.Memory.Hram[0x0F])
	// fmt.Printf("A:  0x%X | F:  0x%X | AF: 0x%X\n", z.A, z.F, z.AF)
	// fmt.Printf("B:  0x%X | C:  0x%X | BC: 0x%X\n", z.B, z.C, z.BC)
	// fmt.Printf("D:  0x%X | E:  0x%X | DE: 0x%X\n", z.D, z.E, z.DE)
	// fmt.Printf("H:  0x%X | L:  0x%X | HL: 0x%X\n", z.H, z.L, z.HL)
	// fmt.Printf("0xFF80:  0x%X | 0xFF81:  0x%X | 0xFF82:  0x%X\n", z.Memory.Hram[0x80], z.Memory.Hram[0x81], z.Memory.Hram[0x82])
	// // fmt.Printf("Scanline:  %d\n", z.Memory.gb.scanlineCounter)
	// // fmt.Printf("TIMA: %X | TAC: %X  | IE: 0x%X\n", z.Memory.Hram[0x05], z.Memory.Hram[0x07], z.Memory.Hram[0xFF])
	// // fmt.Printf("IME:  %t\n", z.InterruptsEnabling)
	// fmt.Printf("NEXT OPCODE:  0x%X%s\n", opcode, fmt.Sprintf(" 0x%X", cb))
	// fmt.Println("=============================================")

	// fmt.Printf(`
	// 	================ Registros ================
	// 	Z: %t | N: %t | HF: %t | CF: %t
	// 	PC: 0x%X | SP: 0x%X | LY:  0x%X
	// 	LCDC: %X | STAT: %X | LYC: %X  IF: 0x%X
	// 	A:  0x%X | F:  0x%X | AF: 0x%X
	// 	B:  0x%X | C:  0x%X | BC: 0x%X
	// 	D:  0x%X | E:  0x%X | DE: 0x%X
	// 	H:  0x%X | L:  0x%X | HL: 0x%X
	// 	0xFF80:  0x%X | 0xFF81:  0x%X | 0xFF82:  0x%X
	// 	NEXT OPCODE:  0x%X%s
	// 	=============================================
	// 	`,
	// 	z.Z, z.N, z.HF, z.CF,
	// 	z.PC, z.SP, z.Memory.Hram[0x44],
	// 	z.Memory.Hram[0x40], z.Memory.Hram[0x41], z.Memory.Hram[0x45], z.Memory.Hram[0x0F],
	// 	z.A, z.F, z.AF,
	// 	z.B, z.C, z.BC,
	// 	z.D, z.E, z.DE,
	// 	z.H, z.L, z.HL,
	// 	z.Memory.Hram[0x80], z.Memory.Hram[0x81], z.Memory.Hram[0x82],
	// 	opcode, fmt.Sprintf(" 0x%X", cb),
	// )

	logger.LogMessage(fmt.Sprintf(`
================ Registros ================
Z: %t | N: %t | HF: %t | CF: %t
PC: 0x%X | SP: 0x%X | LY:  0x%X
LCDC: %X | STAT: %X | LYC: %X  IF: 0x%X
A:  0x%X | F:  0x%X | AF: 0x%X
B:  0x%X | C:  0x%X | BC: 0x%X
D:  0x%X | E:  0x%X | DE: 0x%X
H:  0x%X | L:  0x%X | HL: 0x%X
0xFF80:  0x%X | 0xFF81:  0x%X | 0xFF82:  0x%X
NEXT OPCODE:  0x%X%s
=============================================
			`,
		z.Z, z.N, z.HF, z.CF,
		z.PC, z.SP, z.Memory.Hram[0x44],
		z.Memory.Hram[0x40], z.Memory.Hram[0x41], z.Memory.Hram[0x45], z.Memory.Hram[0x0F],
		z.A, z.F, z.AF,
		z.B, z.C, z.BC,
		z.D, z.E, z.DE,
		z.H, z.L, z.HL,
		z.Memory.Hram[0x80], z.Memory.Hram[0x81], z.Memory.Hram[0x82],
		opcode,
		fmt.Sprintf(" 0x%X", cb),
	))

	// logger.LogMessage("\n================ Registros ================")
	// logger.LogMessage(fmt.Sprintf("Z: %t | N: %t | HF: %t | CF: %t", z.Z, z.N, z.HF, z.CF))
	// logger.LogMessage(fmt.Sprintf("PC: 0x%X | SP: 0x%X | LY:  0x%X", z.PC, z.SP, z.Memory.Hram[0x44]))
	// logger.LogMessage(fmt.Sprintf("LCDC: %X | STAT: %X | LYC: %X  IF: 0x%X", z.Memory.Hram[0x40], z.Memory.Hram[0x41], z.Memory.Hram[0x45], z.Memory.Hram[0x0F]))
	// logger.LogMessage(fmt.Sprintf("A:  0x%X | F:  0x%X | AF: 0x%X", z.A, z.F, z.AF))
	// logger.LogMessage(fmt.Sprintf("B:  0x%X | C:  0x%X | BC: 0x%X", z.B, z.C, z.BC))
	// logger.LogMessage(fmt.Sprintf("D:  0x%X | E:  0x%X | DE: 0x%X", z.D, z.E, z.DE))
	// logger.LogMessage(fmt.Sprintf("H:  0x%X | L:  0x%X | HL: 0x%X", z.H, z.L, z.HL))
	// logger.LogMessage(fmt.Sprintf("0xFF80:  0x%X | 0xFF81:  0x%X | 0xFF82:  0x%X", z.Memory.Hram[0x80], z.Memory.Hram[0x81], z.Memory.Hram[0x82]))
	// logger.LogMessage(fmt.Sprintf("NEXT OPCODE:  0x%X", opcode))
	// logger.LogMessage("=============================================")

	// if z.PC == 0xC428 {
	// 	fmt.Println("---------------------------0x30---------------------------")
	// 	fmt.Scanln()
	// }
	z.ExecuteInstruction(opcode)

	return z.M
}
