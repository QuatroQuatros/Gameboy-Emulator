package gb

import (
	"fmt"
	"gameboy/bits"
	_ "sort"
)

type Z80 struct {
	A, B, C, D, E, H, L, F byte   // Registradores de 8 bits
	AF, BC, DE, HL         uint16 // Pares de registradores de 16 bits
	PC, SP                 uint16 // Contador de programa (PC) e ponteiro de pilha (SP)

	Z, N, HF, CF bool // Flags de status (Zero, Negativo, Meio-carro, Carry, Interrupt Master Enable)

	IME bool

	M int // Contador de ciclos de máquina

	Divider int

	Memory *Memory
}

func (z *Z80) SpHiLo() uint16 {
	return z.SP
}

func (z *Z80) setFlag(index byte, on bool) {
	fmt.Println("SET FLAG")
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
		z.Z = true
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

	// z.B = 0x00
	// z.C = 0x13
	// z.D = 0x00
	// z.E = 0xD8
	// z.H = 0x01
	// z.L = 0x4D

	z.setAF()
	z.setBC()
	z.setDE()
	z.setHL()
	z.Memory = memory

	//cpu.AF.mask = 0xFFF0
}

func (z *Z80) readMemory(addr uint16) byte {
	if z.Memory == nil {
		fmt.Println("Erro: Memória não inicializada")
		return 0xFF // Retornar valor padrão em caso de memória não inicializada
	}

	return z.Memory.ReadByte(addr)
}

func (z *Z80) updateFlagsInc(value byte) {
	z.Z = value == 0
	z.N = false
	z.HF = (value & 0x0F) == 0
	z.setFlags()
}

func (z *Z80) ExecuteInstruction(opcode byte) {
	switch opcode {
	case 0x00:
		z.NOP()
	case 0x01:
		z.LD_BC_nn()
	case 0x03:
		z.INC_BC()
	case 0x05:
		z.DEC_B()
	case 0x06:
		z.LD_B_d8()
	case 0x0B:
		z.DEC_BC()
	case 0x0C:
		z.INC_C()
	case 0x0D:
		z.DEC_C()
	case 0x0E:
		z.LD_C_d8()
	case 0x11:
		z.LD_DE_nn()
	case 0x12:
		z.LD_DE_A()
	case 0x13:
		z.INC_DE()
	case 0x14:
		z.INC_D()
	case 0x16:
		z.LD_D_d8()
	case 0x18:
		z.JR_e()
	case 0x19:
		z.ADD_HL_DE()
	case 0x1A:
		z.LD_A_DE_addr()
	case 0x1C:
		z.INC_E()
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
	case 0x2A:
		z.LDI_A_HL()
	case 0x2C:
		z.INC_L()
	case 0x2D:
		z.DEC_L()
	case 0x2F:
		z.CPL()
	case 0x30:
		z.JR_NC_r8()
	case 0x31:
		z.LD_SP_nn()
	case 0x32:
		z.LD_HL_dec_A()
	case 0x34:
		z.INC_HL_addr()
	case 0x36:
		z.LD_HL_d8()
	case 0x3C:
		z.INC_A()
	case 0x3D:
		z.DEC_A()
	case 0x3E:
		z.LD_A_d8()
	case 0x46:
		z.LD_B_HL_addr()
	case 0x47:
		z.LD_B_A()
	case 0x4E:
		z.LD_C_HL_addr()
	case 0x4F:
		z.LD_C_A()
	case 0x56:
		z.LD_D_HL_addr()
	case 0x57:
		z.LD_D_A()
	case 0x5E:
		z.LD_E_HL()
	case 0x5F:
		z.LD_E_A()
	case 0x70:
		z.LD_HL_B()
	case 0x71:
		z.LD_HL_C()
	case 0x72:
		z.LD_HL_D()
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
		z.LD_A_HL()
	case 0x7F:
		z.LD_A_A()
	// case 0x87:
	//     z.RES_0_A()
	case 0xA1:
		z.AND_C()
	case 0xA7:
		z.AND_A()
	case 0xA9:
		z.XOR_C()
	case 0xAE:
		z.XOR_HL_addr()
	case 0xAF:
		z.XOR_A()
	case 0xB0:
		z.OR_B()
	case 0xB1:
		z.OR_C()
	case 0xB7:
		z.OR_A()
	case 0xC0:
		z.RET_NZ()
	case 0xC1:
		z.POP_BC()
	case 0xC3:
		z.JP_nn()
	case 0xC4:
		z.CALL_NZ_nn()
	case 0xC5:
		z.PUSH_BC()
	case 0xC6:
		z.ADD_A_d8()
	case 0xC8:
		z.RET_Z()
	case 0xC9:
		z.RET()
	case 0xCB:
		z.ExecuteCBInstruction()
	case 0xCD:
		z.CALL_nn()
	case 0xCE:
		z.ADC_A_d8()
	case 0xD1:
		z.POP_DE()
	case 0xD5:
		z.PUSH_DE()
	case 0xD6:
		z.SUB_A_d8()
	case 0xD9:
		z.RETI()
	case 0xE0:
		z.LDH_a8_A()
	case 0xE1:
		z.POP_HL()
	case 0xE2:
		z.LD_c_A()
	case 0xE5:
		z.PUSH_HL()
	case 0xE6:
		z.AND_n()
	case 0xE9:
		z.JP_HL()
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
	case 0xF3:
		z.DI()
	case 0xF5:
		z.PUSH_AF()
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
		fmt.Printf("Opcode não suportado: 0x%X\n", opcode)
		fmt.Scanln()
		z.PC++
		z.M = 4
	}
}

func (z *Z80) EmulateCycle() int {
	opcode := z.readMemory(z.PC)
	fmt.Println("\n================ Registros ================")
	fmt.Printf("Z: %t N: %t HF: %t CF: %t\n", z.Z, z.N, z.HF, z.CF)
	fmt.Printf("PC: 0x%X  SP: 0x%X  LY: 0x%X\n", z.PC, z.SP, z.Memory.Hram[0x44])
	fmt.Printf("A:  0x%X  F:  0x%X  AF: 0x%X\n", z.A, z.F, z.AF)
	fmt.Printf("B:  0x%X  C:  0x%X  BC: 0x%X\n", z.B, z.C, z.BC)
	fmt.Printf("D:  0x%X  E:  0x%X  DE: 0x%X\n", z.D, z.E, z.DE)
	fmt.Printf("H:  0x%X  L:  0x%X  HL: 0x%X\n", z.H, z.L, z.HL)
	fmt.Printf("NEXT OPCODE:  0x%X\n", opcode)
	fmt.Println("=============================================")
	//fmt.Scanln()

	z.ExecuteInstruction(opcode)

	return z.M
}
