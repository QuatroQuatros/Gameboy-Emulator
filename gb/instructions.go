package gb

import (
	"fmt"
	"gameboy/bits"
)

// 0X00 - NOP
func (z *Z80) NOP() {
	z.PC++
	z.M = 4
}

// 0x01 - LD BC, nn
func (z *Z80) LD_BC_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	value := (highByte << 8) | lowByte

	z.BC = value

	// Atualiza os registradores B e C com base no valor de BC
	z.B = uint8(z.BC >> 8)   // Obtém o byte mais significativo (high byte)
	z.C = uint8(z.BC & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC += 3
	z.M = 12
}

// 0x02 - LD (BC), A
func (z *Z80) LD_BC_addr_A() {
	z.Memory.WriteByte(z.BC, z.A)
	z.PC++
	z.M = 8
}

// 0x03 - INC BC
func (z *Z80) INC_BC() {
	z.BC++

	z.B = uint8(z.BC >> 8)
	z.C = uint8(z.BC & 0xFF)

	z.PC++
	z.M = 8
}

// 0x04 - INC B
func (z *Z80) INC_B() {
	z.B++
	z.setBC()

	z.updateFlagsInc(z.B)

	z.PC++
	z.M = 4

}

// 0x05 - DEC B
func (z *Z80) DEC_B() {
	z.B--
	z.setBC()

	z.updateFlagsDec(z.B)

	z.PC++
	z.M = 4
}

// 0x06 - LD B, d8
func (z *Z80) LD_B_d8() {
	immediate := z.Memory.ReadByte(z.PC + 1)

	z.B = immediate
	z.setBC()

	z.PC += 2
	z.M = 8

}

// 0x07 - RLCA (Rotate Left through Carry A)
func (z *Z80) RLCA() {
	carry := z.A >> 7

	z.A = (z.A << 1) | carry

	z.Z = false
	z.N = false
	z.HF = false
	z.CF = carry != 0

	z.setFlags()

	z.PC++
	z.M = 4

}

// 0x08 - LD (a16), SP
func (z *Z80) LD_a16_SP() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	z.Memory.WriteByte(address, byte(z.SP&0xFF))
	z.Memory.WriteByte(address+1, byte((z.SP>>8)&0xFF))

	z.PC += 3
	z.M = 20
}

// 0x09 - ADD HL, BC
func (z *Z80) ADD_HL_BC() {
	result := uint32(z.HL) + uint32(z.BC)

	z.N = false
	z.HF = (z.HL&0x0FFF)+(z.BC&0x0FFF) > 0x0FFF
	z.CF = result > 0xFFFF
	z.setFlags()

	z.HL = uint16(result)

	z.H = uint8(result >> 8)
	z.L = uint8(result & 0xFF)

	z.PC++
	z.M = 8
}

// 0x0A - LD A, (BC)
func (z *Z80) LD_A_BC_addr() {
	z.A = z.Memory.ReadByte(z.BC)
	z.setAF()

	z.PC++
	z.M = 8
}

// 0x0B - DEC BC
func (z *Z80) DEC_BC() {
	z.BC--

	z.B = uint8(z.BC >> 8)   // Obtém o byte mais significativo (high byte)
	z.C = uint8(z.BC & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC++
	z.M = 8

}

// 0x0C - INC C
func (z *Z80) INC_C() {
	z.C++
	z.setBC()

	z.updateFlagsInc(z.C)

	z.PC++
	z.M = 4

}

// 0x0D - DEC C
func (z *Z80) DEC_C() {
	z.C--
	z.setBC()
	z.updateFlagsDec(z.C)

	// z.Z = z.C == 0
	// z.N = true
	// z.HF = (z.C & 0x0F) == 0x0F
	// z.setFlags()

	z.PC++
	z.M = 4
}

// 0x0E - LD C, d8
func (z *Z80) LD_C_d8() {
	// Lê o byte imediatamente seguinte ao PC para obter o valor de 8 bits (d8)
	immediate := z.Memory.ReadByte(z.PC + 1)

	z.C = immediate
	z.setBC()

	z.PC += 2
	z.M = 8
}

// 0x0F - RRCA (Rotate Right through Carry A)
func (z *Z80) RRCA() {
	bit0 := z.A & 1                // Pega o bit 0 de A
	z.A = (z.A >> 1) | (bit0 << 7) // Rotaciona e coloca bit 0 no bit 7
	z.CF = bit0 != 0               // Atualiza carry

	// Flags Z, N, H sempre 0
	z.Z = false
	z.N = false
	z.HF = false
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0x10 - STOP
func (z *Z80) STOP() {
	z.Memory.gb.halted = true

	if z.Memory.gb.IsCGB() {
		z.Memory.gb.checkSpeedSwitch()
	}

	z.PC += 2
	z.M = 4

	// fmt.Println("---------------------------0x10---------------------------")
	// opcode := z.readMemory(z.PC)
	// fmt.Println("\n================ Registros ================")
	// fmt.Printf("Z: %t N: %t HF: %t CF: %t\n", z.Z, z.N, z.HF, z.CF)
	// fmt.Printf("PC: 0x%X  SP: 0x%X  LY: 0x%X\n", z.PC, z.SP, z.Memory.Hram[0x44])
	// fmt.Printf("A:  0x%X  F:  0x%X  AF: 0x%X\n", z.A, z.F, z.AF)
	// fmt.Printf("B:  0x%X  C:  0x%X  BC: 0x%X\n", z.B, z.C, z.BC)
	// fmt.Printf("D:  0x%X  E:  0x%X  DE: 0x%X\n", z.D, z.E, z.DE)
	// fmt.Printf("H:  0x%X  L:  0x%X  HL: 0x%X\n", z.H, z.L, z.HL)
	// fmt.Printf("NEXT OPCODE:  0x%X\n", opcode)
	// fmt.Println("=============================================")
	// fmt.Scanln()
}

// 0x11 - LD DE, nn
func (z *Z80) LD_DE_nn() {
	// Lê os bytes imediatos (nn) da memória
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))

	value := (highByte << 8) | lowByte

	z.DE = value

	z.D = uint8(value >> 8)
	z.E = uint8(value & 0xFF)

	z.PC += 3
	z.M = 12
}

// 0x12 - LD (DE), A
func (z *Z80) LD_DE_addr_A() {
	z.Memory.WriteByte(z.DE, z.A)

	z.PC++
	z.M = 8
}

// 0x13 - INC DE
func (z *Z80) INC_DE() {
	z.DE++

	z.D = uint8(z.DE >> 8)
	z.E = uint8(z.DE & 0xFF)

	z.PC++
	z.M = 8
}

// 0x14 - INC D
func (z *Z80) INC_D() {
	z.D++
	z.setDE()
	z.updateFlagsInc(z.D)

	z.PC++
	z.M = 4
}

// 0x15 - DEC D
func (z *Z80) DEC_D() {
	z.D--
	z.setDE()

	z.updateFlagsDec(z.D)

	z.PC++
	z.M = 4
}

// 0x16 - LD D, d8 (Load 8-bit immediate value into D)
func (z *Z80) LD_D_d8() {
	immediate := z.Memory.ReadByte(z.PC + 1)
	z.D = immediate

	z.setDE()

	z.PC += 2
	z.M = 8
}

// 0x17 - RLA (Rotate Left A through Carry)
func (z *Z80) RLA() {
	carry := z.CF

	z.CF = (z.A & 0x80) != 0         // O bit 7 de A vai para o carry
	z.A = (z.A << 1) | bits.B(carry) // Rotaciona e coloca o carry antigo no bit 0

	z.Z = false
	z.N = false
	z.HF = false

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0x18 - JR r8
func (z *Z80) JR_e() {
	displacement := int8(z.readMemory(z.PC + 1))

	newAddress := uint16(int(z.PC) + 2 + int(displacement))

	z.PC = newAddress
	z.M = 12
}

// 0x19 - ADD HL, DE
func (z *Z80) ADD_HL_DE() {
	result := uint32(z.HL) + uint32(z.DE)

	z.N = false
	z.HF = (z.HL&0x0FFF)+(z.DE&0x0FFF) > 0x0FFF
	z.CF = result > 0xFFFF
	z.setFlags()

	z.HL = uint16(result)

	z.H = uint8(result >> 8)
	z.L = uint8(result & 0xFF)

	z.PC++
	z.M = 8
}

// 0x1A - LD A, (DE)
func (z *Z80) LD_A_DE_addr() {
	z.A = z.Memory.ReadByte(z.DE)
	z.setAF()

	z.PC++
	z.M = 8
}

// 0x1B - DEC DE (Decrementa o registrador DE)
func (z *Z80) DEC_DE() {
	z.DE--

	z.D = uint8(z.DE >> 8)
	z.E = uint8(z.DE & 0xFF)

	z.PC++
	z.M = 8
}

// 0x1C - INC E
func (z *Z80) INC_E() {
	z.E++
	z.setDE()
	z.updateFlagsInc(z.E)

	z.PC++
	z.M = 4
}

// 0x1D - DEC E
func (z *Z80) DEC_E() {
	z.E--
	z.setDE()
	z.updateFlagsDec(z.E)

	z.PC++
	z.M = 4
}

// 0x1E - LD E, d8
func (z *Z80) LD_E_d8() {
	immediate := z.Memory.ReadByte(z.PC + 1)

	z.E = immediate
	z.setDE()

	z.PC += 2
	z.M = 8
}

// 0x1F - RRA (Rotate Right through Carry A)
func (z *Z80) RRA() {
	var carry byte
	if z.CF {
		carry = 0x80
	}
	result := byte(z.A>>1) | carry

	z.Z = false
	z.N = false
	z.HF = false
	z.CF = (1 & z.A) == 1
	z.A = result
	z.setFlags()

	z.PC++
	z.M = 4

}

// 0x20 - Jump r8 if not zero (!Z)
func (z *Z80) JR_nz_e() {
	if !z.Z {
		displacement := int8(z.readMemory(z.PC + 1))

		newAddress := uint16(int(z.PC) + 2 + int(displacement))

		z.PC = newAddress
		z.M = 12
	} else {
		z.PC += 2
		z.M = 8
	}
}

// 0x21 - LD HL, nn
func (z *Z80) LD_HL_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	z.HL = (highByte << 8) | lowByte

	z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
	z.L = uint8(z.HL & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC += 3
	z.M = 12
}

// 0x22 - LD (HL+), A
func (z *Z80) LD_HL_inc_A() {
	z.Memory.WriteByte(z.HL, z.A)

	z.HL++

	z.H = uint8(z.HL >> 8)
	z.L = uint8(z.HL & 0xFF)

	z.PC++
	z.M = 8

}

// 0x23 - INC HL
func (z *Z80) INC_HL() {
	z.HL++

	z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
	z.L = uint8(z.HL & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC++
	z.M = 8

}

// 0x24 INC H
func (z *Z80) INC_H() {
	z.H++
	z.setHL()

	z.updateFlagsInc(z.H)

	z.PC++
	z.M = 4
}

// 0x25 - DEC H
func (z *Z80) DEC_H() {
	z.H--
	z.setHL()
	z.updateFlagsDec(z.H)

	z.PC++
	z.M = 4
}

// 0x26 - LD H, d8
func (z *Z80) LD_H_d8() {
	immediate := z.Memory.ReadByte(z.PC + 1)
	z.H = immediate
	z.setHL()

	z.PC += 2
	z.M = 8
}

// 0x27 - DAA (Decimal Adjust Accumulator)
func (z *Z80) DAA() {
	a := uint16(z.A)

	adjust := uint16(0)
	carry := uint16(0)

	if z.HF || (!z.N && (a&0x0F) > 0x09) {
		adjust = 0x06
	}
	if z.CF || (!z.N && a > 0x99) {
		adjust |= 0x60
		carry = 0x60
	}

	if z.N {
		a = (a - adjust) & 0xFF
	} else {
		a = (a + adjust) & 0xFF
	}

	z.A = uint8(a)
	z.Z = z.A == 0
	z.HF = false
	z.CF = (a & 0x100) != 0

	if carry != 0 {
		z.CF = true
	}

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0x28 - jump if zero
func (z *Z80) JR_Z_e() {
	displacement := int8(z.readMemory(z.PC + 1))

	if z.Z {
		newAddress := uint16(int(z.PC) + 2 + int(displacement))

		z.PC = newAddress
		z.M = 12
	} else {
		z.PC += 2
		z.M = 8
	}
}

// 0x29 - ADD HL, HL
func (z *Z80) ADD_HL_HL() {
	result := uint32(z.HL) + uint32(z.HL)

	z.N = false
	z.HF = (z.HL&0x0FFF)+(z.HL&0x0FFF) > 0x0FFF
	//z.HF = (z.HL & 0xFFF) > (result & 0xFFF)
	z.CF = result > 0xFFFF
	z.setFlags()

	z.HL = uint16(result)

	z.H = uint8(result >> 8)
	z.L = uint8(result & 0xFF)

	z.PC++
	z.M = 8

}

// 0x2A - LDI A, (HL)
func (z *Z80) LDI_A_HL() {
	// Obter o byte da memória no endereço apontado por HL e carregar em A
	z.A = z.Memory.ReadByte(z.HL)
	z.setAF()

	z.HL++

	z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
	z.L = uint8(z.HL & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC++
	z.M = 8

}

// 0x2B - DEC HL (Decrementa o registrador HL)
func (z *Z80) DEC_HL() {
	z.HL--

	z.H = uint8(z.HL >> 8)
	z.L = uint8(z.HL & 0xFF)

	z.PC++
	z.M = 8
}

// 0x2C - INC L
func (z *Z80) INC_L() {
	z.L++
	z.setHL()

	z.updateFlagsInc(z.L)

	z.PC++
	z.M = 4
}

// 0x2D - DEC L
func (z *Z80) DEC_L() {
	z.L--
	z.setHL()

	z.updateFlagsDec(z.L)

	z.PC++
	z.M = 4
}

// 0x2E - LD L, d8
func (z *Z80) LD_L_d8() {
	immediate := z.Memory.ReadByte(z.PC + 1)
	z.L = immediate
	z.setHL()

	z.PC += 2
	z.M = 8

}

// 0x2F - CPL (Complement Accumulator)
func (z *Z80) CPL() {
	z.A = ^z.A // Complemento de dois no registrador A
	z.N = true
	z.HF = true
	z.setFlags()

	z.PC++
	z.M = 4

}

// 0x30 - JR NC, r8 (Jump to relative address r8 if C flag is reset)
func (z *Z80) JR_NC_r8() {
	if !z.CF {
		offset := int8(z.readMemory(z.PC + 1))
		newPC := uint16(int32(z.PC) + int32(offset) + 2)

		z.PC = newPC
		z.M = 12
	} else {
		z.PC += 2
		z.M = 8
	}

}

// 0x31 - LD SP, nn
func (z *Z80) LD_SP_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	value := (highByte << 8) | lowByte

	z.SP = value
	z.PC += 3
	z.M = 12
}

// 0x32 - LD (HL-), A
func (z *Z80) LD_HL_dec_A() {
	z.Memory.WriteByte(z.HL, z.A)

	z.HL--

	z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
	z.L = uint8(z.HL & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC++
	z.M = 8

}

// 0x33 - INC SP (Increment Stack Pointer)
func (z *Z80) INC_SP() {
	z.SP++

	z.PC++
	z.M = 8
}

// 0x34 - INC (HL)
func (z *Z80) INC_HL_addr() {
	value := z.Memory.ReadByte(z.HL)
	value++
	z.Memory.WriteByte(z.HL, value)

	z.updateFlagsInc(value)
	z.PC++
	z.M = 12
}

// 0x35 - DEC (HL)
func (z *Z80) DEC_HL_addr() {
	value := z.Memory.ReadByte(z.HL)

	result := value - 1

	z.Memory.WriteByte(z.HL, result)

	z.Z = result == 0
	z.N = true
	z.HF = (value&0x0F == 0)
	z.setFlags()

	z.PC++
	z.M = 12
}

// 0x36 - LD (HL), d8
func (z *Z80) LD_HL_addr_d8() {
	immediate := z.readMemory(z.PC + 1)

	address := z.HL

	z.Memory.WriteByte(address, immediate)

	z.PC += 2
	//z.M = 4
	z.M = 12
}

// 0x37 - SCF (Set Carry Flag)
func (z *Z80) SCF() {
	// Define o Carry Flag (CF = 1) e limpa os flags N e HF
	z.CF = true
	z.N = false
	z.HF = false
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0x38 - jump if carry
func (z *Z80) JR_C_r8() {
	displacement := int8(z.readMemory(z.PC + 1))

	if z.CF {
		newAddress := uint16(int(z.PC) + 2 + int(displacement))

		z.PC = newAddress
		z.M = 12
	} else {
		z.PC += 2
		z.M = 8
	}
}

// 0x39 - ADD HL, SP
func (z *Z80) ADD_HL_SP() {
	result := int32(z.HL) + int32(z.SP)

	z.N = false
	z.HF = (int32(z.HL&0xFFF) > (result & 0xFFF))
	//z.HF = ((z.HL & 0xFFF) + (z.SP & 0xFFF)) > 0xFFF
	z.CF = result > 0xFFFF
	z.setFlags()

	z.HL = uint16(result)

	z.H = uint8(z.HL >> 8)
	z.L = uint8(z.HL & 0xFF)

	z.PC++
	z.M = 8
}

// 0x3A - LDD A, (HL-)
func (z *Z80) LDD_A_HL() {
	// Obter o byte da memória no endereço apontado por HL e carregar em A
	z.A = z.Memory.ReadByte(z.HL)
	z.setAF()

	z.HL--

	z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
	z.L = uint8(z.HL & 0xFF) // Obtém o byte menos significativo (low byte)

	z.PC++
	z.M = 8

}

// 0x3B - DEC SP (Decrement Stack Pointer)
func (z *Z80) DEC_SP() {
	z.SP--

	z.PC++
	z.M = 8
}

// 0x3C - INC A
func (z *Z80) INC_A() {
	z.A++
	z.updateFlagsInc(z.A)

	z.PC++
	z.M = 4

}

// 0x3D - DEC A
func (z *Z80) DEC_A() {
	z.A--
	z.updateFlagsDec(z.A)

	z.PC++
	z.M = 4
}

// 0x3E -  LD A, d8
func (z *Z80) LD_A_d8() {
	immediate := z.readMemory(z.PC + 1)

	z.A = immediate
	z.setAF()

	z.PC += 2
	z.M = 8

}

// 0x3F - CCF (Complement Carry Flag)
func (z *Z80) CCF() {
	// Inverte o Carry Flag (CF) e limpa os flags N e HF
	z.CF = !z.CF
	z.N = false
	z.HF = false
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0x40 - LD B, B
func (z *Z80) LD_B_B() {
	// Nenhuma operação necessária, pois B já contém B
	z.PC++
	z.M = 4
}

// 0x41 - LD B, C
func (z *Z80) LD_B_C() {
	z.B = z.C
	z.setBC()

	z.PC++
	z.M = 4
}

// 0x42 - LD B, D
func (z *Z80) LD_B_D() {
	z.B = z.D
	z.setBC()

	z.PC++
	z.M = 4
}

// 0x43 - LD B, E
func (z *Z80) LD_B_E() {
	z.B = z.E
	z.setBC()

	z.PC++
	z.M = 4
}

// 0x44 - LD B, H
func (z *Z80) LD_B_H() {
	z.B = z.H
	z.setBC()

	z.PC++
	z.M = 4
}

// 0x45 - LD B, L
func (z *Z80) LD_B_L() {
	z.B = z.L
	z.setBC()

	z.PC++
	z.M = 4
}

// 0x46 - LD B, (HL)
func (z *Z80) LD_B_HL_addr() {
	z.B = z.Memory.ReadByte(z.HL)
	z.setBC()

	z.PC++
	z.M = 8
}

// 0x47 - LD B, A
func (z *Z80) LD_B_A() {
	z.B = z.A
	z.setBC()

	z.PC++
	z.M = 4

}

// 0x48 - LD C, B
func (z *Z80) LD_C_B() {
	z.C = z.B
	z.setBC()

	z.PC++
	z.M = 4

}

// 0x49 - LD C, C
func (z *Z80) LD_C_C() {

	z.PC++
	z.M = 4

}

// 0x4A - LD C, D
func (z *Z80) LD_C_D() {
	z.C = z.D
	z.PC++
	z.M = 4

}

// 0x4B - LD C, E
func (z *Z80) LD_C_E() {
	z.C = z.E
	z.PC++
	z.M = 4

}

// 0x4C - LD C, H
func (z *Z80) LD_C_H() {
	z.C = z.H
	z.PC++
	z.M = 4

}

// 0x4D - LD C, L
func (z *Z80) LD_C_L() {
	z.C = z.L
	z.PC++
	z.M = 4

}

// 0x4E - LD C, (HL)
func (z *Z80) LD_C_HL_addr() {
	z.C = z.Memory.ReadByte(z.HL)
	z.setBC()

	z.PC++
	z.M = 8
}

// 0x4F - LD C, A
func (z *Z80) LD_C_A() {
	z.C = z.A
	z.setBC()

	z.PC++
	z.M = 4
}

// 0x50 - LD D, B
func (z *Z80) LD_D_B() {
	z.D = z.B
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x51 - LD D, C
func (z *Z80) LD_D_C() {
	z.D = z.C
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x52 - LD D, D
func (z *Z80) LD_D_D() {
	// Nenhuma operação necessária, pois D já contém D
	z.PC++
	z.M = 4
}

// 0x53 - LD D, E
func (z *Z80) LD_D_E() {
	z.D = z.E
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x54 - LD D, H
func (z *Z80) LD_D_H() {
	z.D = z.H
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x55 - LD D, L
func (z *Z80) LD_D_L() {
	z.D = z.L
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x56 - LD D, (HL)
func (z *Z80) LD_D_HL_addr() {
	z.D = z.Memory.ReadByte(z.HL)
	z.setDE()

	z.PC++
	z.M = 8

}

// 0x57 - LD D, A
func (z *Z80) LD_D_A() {
	z.D = z.A
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x58 - LD E, B
func (z *Z80) LD_E_B() {
	z.E = z.B
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x59 - LD E, C
func (z *Z80) LD_E_C() {
	z.E = z.C
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x5A - LD E, D
func (z *Z80) LD_E_D() {
	z.E = z.D
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x5B - LD E, E
func (z *Z80) LD_E_E() {
	// Nenhuma operação necessária, pois E já contém E
	z.PC++
	z.M = 4
}

// 0x5C - LD E, H
func (z *Z80) LD_E_H() {
	z.E = z.H
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x5D - LD E, L
func (z *Z80) LD_E_L() {
	z.E = z.L
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x5E - LD E, (HL)
func (z *Z80) LD_E_HL_addr() {
	z.E = z.Memory.ReadByte(z.HL)
	z.setDE()

	z.PC++
	z.M = 8
}

// 0x5F - LD E, A (Load A into E)
func (z *Z80) LD_E_A() {
	z.E = z.A
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x60 - LD H, B
func (z *Z80) LD_H_B() {
	z.H = z.B
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x61 - LD H, C
func (z *Z80) LD_H_C() {
	z.H = z.C
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x62 - LD H, D
func (z *Z80) LD_H_D() {
	z.H = z.D
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x63 - LD H, E
func (z *Z80) LD_H_E() {
	z.H = z.E
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x64 - LD H, H
func (z *Z80) LD_H_H() {
	// Nenhuma operação necessária, pois H já contém H
	z.PC++
	z.M = 4
}

// 0x65 - LD H, L
func (z *Z80) LD_H_L() {
	z.H = z.L
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x66 - LD H, (HL)
func (z *Z80) LD_H_HL_addr() {
	z.H = z.Memory.ReadByte(z.HL)
	z.setHL()

	z.PC++
	z.M = 8
}

// 0x67 - LD H, A
func (z *Z80) LD_H_A() {
	z.H = z.A
	z.setHL()

	z.PC++
	z.M = 4

}

// 0x68 - LD L, B
func (z *Z80) LD_L_B() {
	z.L = z.B
	z.setHL()

	z.PC++
	z.M = 4

}

// 0x69 - LD L, C
func (z *Z80) LD_L_C() {
	z.L = z.C
	z.setHL()

	z.PC++
	z.M = 4

}

// 0x6A - LD L, D
func (z *Z80) LD_L_D() {
	z.L = z.D
	z.setHL()

	z.PC++
	z.M = 4

}

// 0x6B - LD L, E
func (z *Z80) LD_L_E() {
	z.L = z.E
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x6C - LD L, H
func (z *Z80) LD_L_H() {
	z.L = z.H
	z.setHL()

	z.PC++
	z.M = 4

}

// 0x6D - LD L, L
func (z *Z80) LD_L_L() {
	// Nenhuma operação necessária, pois L já contém L
	z.PC++
	z.M = 4

}

// 0x6E - LD L, (HL)
func (z *Z80) LD_L_HL_addr() {
	z.L = z.Memory.ReadByte(z.HL)
	z.setHL()

	z.PC++
	z.M = 8
}

// 0x6F - LD L, A
func (z *Z80) LD_L_A() {
	z.L = z.A
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x70 - LD (HL), B
func (z *Z80) LD_HL_addr_B() {
	z.Memory.WriteByte(z.HL, z.B)

	z.PC++
	z.M = 8
}

// 0x71 - LD (HL), C
func (z *Z80) LD_HL_addr_C() {
	z.Memory.WriteByte(z.HL, z.C)

	z.PC++
	z.M = 8
}

// 0x72 - LD (HL), D
func (z *Z80) LD_HL_addr_D() {
	z.Memory.WriteByte(z.HL, z.D)

	z.PC++
	z.M = 8
}

// 0x73 - LD (HL), E
func (z *Z80) LD_HL_addr_E() {
	z.Memory.WriteByte(z.HL, z.E)
	z.PC++
	z.M = 8
}

// 0x74 - LD (HL), H
func (z *Z80) LD_HL_addr_H() {
	z.Memory.WriteByte(z.HL, z.H)
	z.PC++
	z.M = 8
}

// 0x75 - LD (HL), L
func (z *Z80) LD_HL_addr_L() {
	z.Memory.WriteByte(z.HL, z.L)
	z.PC++
	z.M = 8
}

// 0x76 - HALT
func (z *Z80) HALT() {
	z.Memory.gb.halted = true

	z.PC++
	z.M = 4
}

// 0x77 - LD (HL), A
func (z *Z80) LD_HL_addr_A() {
	z.Memory.WriteByte(z.HL, z.A)

	z.PC++
	z.M = 8
}

// 0x78 - LD A, B
func (z *Z80) LD_A_B() {
	z.A = z.B
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x79 - LD A, C
func (z *Z80) LD_A_C() {
	z.A = z.C
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7A - LD A, D
func (z *Z80) LD_A_D() {
	z.A = z.D
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7B - LD A, E
func (z *Z80) LD_A_E() {
	z.A = z.E
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7C - LD A, H
func (z *Z80) LD_A_H() {
	z.A = z.H
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7D - LD A, L
func (z *Z80) LD_A_L() {
	z.A = z.L
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7E - LD A, (HL)
func (z *Z80) LD_A_HL_addr() {
	z.A = z.Memory.ReadByte(z.HL)
	z.setAF()

	z.PC++
	z.M = 8
}

// 0x7F - LD A, A
func (z *Z80) LD_A_A() {
	z.PC++
	z.M = 4
}

// 0x80 - ADD A, B
func (z *Z80) ADD_A_B() {
	z.updateFlagsAdd(z.A, z.B)

	z.PC++
	z.M = 4
}

// 0x81 - ADD A, C
func (z *Z80) ADD_A_C() {
	z.updateFlagsAdd(z.A, z.C)

	z.PC++
	z.M = 4
}

// 0x82 - ADD A, D
func (z *Z80) ADD_A_D() {
	z.updateFlagsAdd(z.A, z.D)

	z.PC++
	z.M = 4
}

// 0x83 - ADD A, E
func (z *Z80) ADD_A_E() {
	z.updateFlagsAdd(z.A, z.E)

	z.PC++
	z.M = 4
}

// 0x84 - ADD A, H
func (z *Z80) ADD_A_H() {
	z.updateFlagsAdd(z.A, z.H)

	z.PC++
	z.M = 4
}

// 0x85 - ADD A, L
func (z *Z80) ADD_A_L() {
	z.updateFlagsAdd(z.A, z.L)

	z.PC++
	z.M = 4
}

// 0x86 - ADD A, (HL)
func (z *Z80) ADD_A_HL_addr() {
	z.updateFlagsAdd(z.A, z.readMemory(z.HL))

	z.PC++
	z.M = 8
}

// 0x87 - ADD A, A
func (z *Z80) ADD_A_A() {
	z.updateFlagsAdd(z.A, z.A)

	z.PC++
	z.M = 4
}

// 0x88 - ADC A, B
func (z *Z80) ADC_A_B() {
	z.updateFlagsAdc(z.A, z.B)

	z.PC++
	z.M = 4
}

// 0x89 - ADC A, C
func (z *Z80) ADC_A_C() {
	z.updateFlagsAdc(z.A, z.C)

	z.PC++
	z.M = 4
}

// 0x8A - ADC A, D
func (z *Z80) ADC_A_D() {
	z.updateFlagsAdc(z.A, z.D)

	z.PC++
	z.M = 4
}

// 0x8B - ADC A, E
func (z *Z80) ADC_A_E() {
	z.updateFlagsAdc(z.A, z.E)

	z.PC++
	z.M = 4
}

// 0x8C - ADC A, H
func (z *Z80) ADC_A_H() {
	z.updateFlagsAdc(z.A, z.H)

	z.PC++
	z.M = 4
}

// 0x8D - ADC A, L
func (z *Z80) ADC_A_L() {
	z.updateFlagsAdc(z.A, z.L)

	z.PC++
	z.M = 4
}

// 0x8E - ADC A, (HL)
func (z *Z80) ADC_A_HL_addr() {
	z.updateFlagsAdc(z.A, z.readMemory(z.HL))

	z.PC++
	z.M = 8
}

// 0x8F - ADC A, A
func (z *Z80) ADC_A_A() {
	z.updateFlagsAdc(z.A, z.A)

	z.PC++
	z.M = 4
}

// 0x90 - SUB B
func (z *Z80) SUB_B() {
	z.updateFlagsSub(z.A, z.B)

	z.PC++
	z.M = 4
}

// 0x91 - SUB C
func (z *Z80) SUB_C() {
	z.updateFlagsSub(z.A, z.C)

	z.PC++
	z.M = 4
}

// 0x92 - SUB D
func (z *Z80) SUB_D() {
	z.updateFlagsSub(z.A, z.D)

	z.PC++
	z.M = 4
}

// 0x93 - SUB E
func (z *Z80) SUB_E() {
	z.updateFlagsSub(z.A, z.E)

	z.PC++
	z.M = 4
}

// 0x94 - SUB H
func (z *Z80) SUB_H() {
	z.updateFlagsSub(z.A, z.H)

	z.PC++
	z.M = 4
}

// 0x95 - SUB L
func (z *Z80) SUB_L() {
	z.updateFlagsSub(z.A, z.L)

	z.PC++
	z.M = 4
}

// 0x96 - SUB A, (HL)
func (z *Z80) SUB_HL_addr() {
	z.updateFlagsSub(z.A, z.readMemory(z.HL))

	z.PC++
	z.M = 8
}

// 0x97 - SUB A
func (z *Z80) SUB_A() {
	z.updateFlagsSub(z.A, z.A)

	z.PC++
	z.M = 4
}

// 0x98 - SBC A, B
func (z *Z80) SBC_A_B() {
	z.updateFlagsSbc(z.A, z.B)

	z.PC++
	z.M = 4
}

// 0x99 - SBC A, C
func (z *Z80) SBC_A_C() {
	z.updateFlagsSbc(z.A, z.C)

	z.PC++
	z.M = 4
}

// 0x9A - SBC A, D
func (z *Z80) SBC_A_D() {
	z.updateFlagsSbc(z.A, z.D)

	z.PC++
	z.M = 4
}

// 0x9B - SBC A, E
func (z *Z80) SBC_A_E() {
	z.updateFlagsSbc(z.A, z.E)

	z.PC++
	z.M = 4
}

// 0x9C - SBC A, H
func (z *Z80) SBC_A_H() {
	z.updateFlagsSbc(z.A, z.H)

	z.PC++
	z.M = 4
}

// 0x9D - SBC A, L
func (z *Z80) SBC_A_L() {
	z.updateFlagsSbc(z.A, z.L)

	z.PC++
	z.M = 4
}

// 0x9E - SBC A, (HL)
func (z *Z80) SBC_A_HL_addr() {
	z.updateFlagsSbc(z.A, z.readMemory(z.HL))

	z.PC++
	z.M = 8
}

// 0x9F - SBC A, A
func (z *Z80) SBC_A_A() {
	z.updateFlagsSbc(z.A, z.A)

	z.PC++
	z.M = 4
}

// 0xA0 - AND B
func (z *Z80) AND_B() {
	z.updateFlagsAnd(z.B)

	z.PC++
	z.M = 4
}

// 0xA1 - AND C
func (z *Z80) AND_C() {
	z.updateFlagsAnd(z.C)

	z.PC++
	z.M = 4
}

// 0xA2 - AND D
func (z *Z80) AND_D() {
	z.updateFlagsAnd(z.D)

	z.PC++
	z.M = 4
}

// 0xA3 - AND E
func (z *Z80) AND_E() {
	z.updateFlagsAnd(z.E)

	z.PC++
	z.M = 4
}

// 0xA4 - AND H
func (z *Z80) AND_H() {
	z.updateFlagsAnd(z.H)

	z.PC++
	z.M = 4
}

// 0xA5 - AND L
func (z *Z80) AND_L() {
	z.updateFlagsAnd(z.L)

	z.PC++
	z.M = 4
}

// 0xA6 - AND (HL)
func (z *Z80) AND_HL_addr() {
	z.updateFlagsAnd(z.readMemory(z.HL))

	z.PC++
	z.M = 8
}

// 0xA7 - AND A
func (z *Z80) AND_A() {
	z.updateFlagsAnd(z.A)

	z.PC++
	z.M = 4
}

// 0xA8 - XOR B
func (z *Z80) XOR_B() {
	z.updateFlagsXor(z.B)

	z.PC++
	z.M = 4
}

// 0xA9 - XOR C
func (z *Z80) XOR_C() {
	z.updateFlagsXor(z.C)

	z.PC++
	z.M = 4
}

// 0xAA - XOR D
func (z *Z80) XOR_D() {
	z.updateFlagsXor(z.D)

	z.PC++
	z.M = 4
}

// 0xAB - XOR E
func (z *Z80) XOR_E() {
	z.updateFlagsXor(z.E)

	z.PC++
	z.M = 4
}

// 0xAC - XOR H
func (z *Z80) XOR_H() {
	z.updateFlagsXor(z.H)

	z.PC++
	z.M = 4
}

// 0xAD - XOR L (Exclusive OR A with L)
func (z *Z80) XOR_L() {
	z.updateFlagsXor(z.L)

	z.PC++
	z.M = 4
}

// 0xAE - XOR (HL)
func (z *Z80) XOR_HL_addr() {
	z.updateFlagsXor(z.readMemory(z.HL))

	z.PC++
	z.M = 8
}

// 0xAF - XOR A
func (z *Z80) XOR_A() {
	z.updateFlagsXor(z.A)

	z.PC++
	z.M = 4
}

// 0xB0 - OR B
func (z *Z80) OR_B() {
	z.updateFlagsOr(z.B)

	z.PC++
	z.M = 4
}

// 0xB1 - OR C
func (z *Z80) OR_C() {
	z.updateFlagsOr(z.C)

	z.PC++
	z.M = 4
}

// 0xB2 - OR D
func (z *Z80) OR_D() {
	z.updateFlagsOr(z.D)

	z.PC++
	z.M = 4
}

// 0xB3 - OR E
func (z *Z80) OR_E() {
	z.updateFlagsOr(z.E)

	z.PC++
	z.M = 4
}

// 0xB4 - OR H
func (z *Z80) OR_H() {
	z.updateFlagsOr(z.H)

	z.PC++
	z.M = 4
}

// 0xB5 - OR L
func (z *Z80) OR_L() {
	z.updateFlagsOr(z.L)

	z.PC++
	z.M = 4
}

// 0xB6 - OR (HL)
func (z *Z80) OR_HL_addr() {
	value := z.Memory.ReadByte(z.HL)
	z.updateFlagsOr(value)

	z.PC++
	z.M = 8
}

// 0xB7 - OR A
func (z *Z80) OR_A() {
	z.updateFlagsOr(z.A)

	z.PC++
	z.M = 4
}

// 0xB8 - CP B (Compare B with A)
func (z *Z80) CP_B() {
	z.updateFlagsCP(z.B)

	z.PC++
	z.M = 4
}

// 0xB9 - CP C (Compare C with A)
func (z *Z80) CP_C() {
	z.updateFlagsCP(z.C)

	z.PC++
	z.M = 4

}

// 0xBA - CP D (Compare D with A)
func (z *Z80) CP_D() {
	z.updateFlagsCP(z.D)

	z.PC++
	z.M = 4
}

// 0xBB - CP E (Compare E with A)
func (z *Z80) CP_E() {
	z.updateFlagsCP(z.E)

	z.PC++
	z.M = 4
}

// 0xBC - CP H (Compare H with A)
func (z *Z80) CP_H() {
	z.updateFlagsCP(z.H)

	z.PC++
	z.M = 4
}

// 0xBD - CP L (Compare L with A)
func (z *Z80) CP_L() {
	z.updateFlagsCP(z.L)

	z.PC++
	z.M = 4
}

// 0xBE - CP (HL) (Compara A com o valor apontado por HL)
func (z *Z80) CP_HL_addr() {
	value := z.readMemory(z.HL)
	z.updateFlagsCP(value)

	z.PC++
	z.M = 8
}

// 0xBF - CP A (Compare A with A)
func (z *Z80) CP_A() {
	z.updateFlagsCP(z.A)

	z.PC++
	z.M = 4
}

// 0xC0 - RET NZ (Return if Not Zero)
func (z *Z80) RET_NZ() {
	if !z.Z {
		lowByte := z.Memory.ReadWord(z.SP)
		highByte := z.Memory.ReadWord(z.SP + 1)
		z.SP += 2

		z.PC = (highByte << 8) | lowByte
		z.M = 20
	} else {
		z.PC++
		z.M = 8
	}
}

// 0xC1 - POP BC
func (z *Z80) POP_BC() {
	lowByte := uint16(z.readMemory(z.SP))         //C
	highByte := uint16(z.readMemory(z.SP+1)) << 8 //B

	z.BC = lowByte | highByte
	z.B = uint8(z.BC >> 8)   // Obtém o byte mais significativo (high byte)
	z.C = uint8(z.BC & 0xFF) // Obtém o byte menos significativo (low byte)
	z.SP += 2

	z.PC++
	z.M = 12
}

// 0xC2 - JP NZ, nn (Jump to address nn if not zero !Z)
func (z *Z80) JP_NZ_nn() {
	// Lê os bytes imediatos (nn) da memória
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	if !z.Z {
		z.PC = address
		z.M = 16
	} else {
		z.PC += 3
		z.M = 12
	}
}

// 0xC3 - JP nn
func (z *Z80) JP_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	targetAddress := (highByte << 8) | lowByte

	z.PC = targetAddress
	z.M = 16
}

// 0xC4 - CALL NZ nn
func (z *Z80) CALL_NZ_nn() {
	if !z.Z {
		lowByte := uint16(z.readMemory(z.PC + 1))
		highByte := uint16(z.readMemory(z.PC + 2))
		address := (highByte << 8) | lowByte

		returnAddress := z.PC + 3

		z.SP -= 2
		z.Memory.WriteWord(z.SP, returnAddress)

		z.PC = address
		z.M = 24
	} else {
		z.PC += 3
		z.M = 12
	}

}

// 0xC5 - PUSH BC
func (z *Z80) PUSH_BC() {
	z.Memory.WriteByte(z.SP-1, byte(z.B))
	z.Memory.WriteByte(z.SP-2, byte(z.C))

	z.SP -= 2

	z.PC++
	z.M = 16
}

// 0xC6 - ADD A, d8
func (z *Z80) ADD_A_d8() {
	immediate := z.readMemory(z.PC + 1)
	z.updateFlagsAdd(z.A, immediate)

	z.PC += 2
	z.M = 8

}

// 0xC7 - RST 00H
func (z *Z80) RST_00H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0000
	z.PC = 0x0000
	z.M = 16
}

// 0xC8 - RET Z (Return if Zero)
func (z *Z80) RET_Z() {
	if z.Z {
		returnAddress := z.Memory.ReadWord(z.SP)
		z.SP += 2

		z.PC = returnAddress
		z.M = 20
	} else {
		z.PC++
		z.M = 8
	}
}

// 0xC9 - RET
func (z *Z80) RET() {
	returnAddress := z.Memory.ReadWord(z.SP)
	z.SP += 2

	z.PC = returnAddress
	z.M = 16
}

// 0xCA - JP Z, nn
func (z *Z80) JP_Z_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	if z.Z {
		z.PC = address
		z.M = 16
	} else {
		z.PC += 3
		z.M = 12
	}

}

// 0xCC - CALL Z nn
func (z *Z80) CALL_Z_nn() {
	if z.Z {
		lowByte := uint16(z.readMemory(z.PC + 1))
		highByte := uint16(z.readMemory(z.PC + 2))
		address := (highByte << 8) | lowByte

		returnAddress := z.PC + 3

		z.SP -= 2
		z.Memory.WriteWord(z.SP, returnAddress)

		z.PC = address
		z.M = 24
	} else {
		z.PC += 3
		z.M = 12
	}

}

// 0xCD - CALL nn
func (z *Z80) CALL_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	returnAddress := z.PC + 3

	z.SP -= 2
	z.Memory.WriteWord(z.SP, returnAddress)

	z.PC = address
	z.M = 24
}

// 0xCE - ADC A, d8 (Add with Carry immediate 8-bit to A)
func (z *Z80) ADC_A_d8() {
	immediate := z.readMemory(z.PC + 1)
	z.updateFlagsAdc(z.A, immediate)

	z.PC += 2
	z.M = 8
}

// 0xCF - RST 08H
func (z *Z80) RST_08H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0008
	z.PC = 0x0008
	z.M = 16
}

// 0xD0 - RET NC (Return if Not Carry)
func (z *Z80) RET_NC() {
	if !z.CF {
		lowByte := uint16(z.Memory.ReadByte(z.SP))
		highByte := uint16(z.Memory.ReadByte(z.SP+1)) << 8

		z.SP += 2

		z.PC = uint16(highByte | lowByte)
		z.M = 20
	} else {
		z.PC++
		z.M = 8
	}
}

// 0xD1 - POP DE
func (z *Z80) POP_DE() {
	// Desempilha dois bytes da pilha (SP) para DE
	lowByte := uint16(z.readMemory(z.SP))         //E
	highByte := uint16(z.readMemory(z.SP+1)) << 8 //D

	z.DE = lowByte | highByte
	z.D = uint8(z.DE >> 8)   // Obtém o byte mais significativo (high byte)
	z.E = uint8(z.DE & 0xFF) // Obtém o byte menos significativo (low byte)
	z.SP += 2

	z.PC++
	z.M = 12
}

// 0xD2 - JP NC, nn (Jump to address nn if not carry !CF)
func (z *Z80) JP_NC_nn() {
	// Lê os bytes imediatos (nn) da memória
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	if !z.CF {
		z.PC = address
		z.M = 16
	} else {
		z.PC += 3
		z.M = 12
	}
}

// 0xD4 - CALL NC nn
func (z *Z80) CALL_NC_nn() {
	if !z.CF {
		lowByte := uint16(z.readMemory(z.PC + 1))
		highByte := uint16(z.readMemory(z.PC + 2))
		address := (highByte << 8) | lowByte

		returnAddress := z.PC + 3

		z.SP -= 2
		z.Memory.WriteWord(z.SP, returnAddress)

		z.PC = address
		z.M = 24
	} else {
		z.PC += 3
		z.M = 12
	}

}

// 0xD5 - PUSH DE
func (z *Z80) PUSH_DE() {
	z.Memory.WriteByte(z.SP-1, byte(z.D))
	z.Memory.WriteByte(z.SP-2, byte(z.E))
	z.SP -= 2

	z.PC++
	z.M = 16
}

// 0xD6 - SUB A, d8
func (z *Z80) SUB_A_d8() {
	immediate := z.readMemory(z.PC + 1)
	z.updateFlagsSub(z.A, immediate)

	z.PC += 2
	z.M = 8

}

// 0xD7 - RST 10H
func (z *Z80) RST_10H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0010
	z.PC = 0x0010
	z.M = 16
}

// 0xD8 - RET C (Return if Carry)
func (z *Z80) RET_C() {
	if z.CF {
		returnAddress := z.Memory.ReadWord(z.SP)
		z.SP += 2

		z.PC = returnAddress
		z.M = 20
	} else {
		z.PC++
		z.M = 8
	}

}

// 0xD9 - RET I
func (z *Z80) RETI() {

	lowByte := uint16(z.readMemory(z.SP))
	highByte := uint16(z.readMemory(z.SP+1)) << 8
	newPC := highByte | lowByte

	z.InterruptsEnabling = true
	z.SP += 2

	z.PC = newPC
	z.M = 16

}

// 0xDA - JP C, nn
func (z *Z80) JP_C_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	if z.CF {
		z.PC = address
		z.M = 16
	} else {
		z.PC += 3
		z.M = 12
	}

}

// 0xDC - CALL C nn
func (z *Z80) CALL_C_nn() {
	if z.CF {
		lowByte := uint16(z.readMemory(z.PC + 1))
		highByte := uint16(z.readMemory(z.PC + 2))
		address := (highByte << 8) | lowByte

		returnAddress := z.PC + 3

		z.SP -= 2
		z.Memory.WriteWord(z.SP, returnAddress)

		z.PC = address
		z.M = 24
	} else {
		z.PC += 3
		z.M = 12
	}

}

// 0xDE - SBC A, d8 (Subtract with Carry)
func (z *Z80) SBC_A_d8() {
	immediate := z.readMemory(z.PC + 1)

	carry := uint8(0)
	if z.CF {
		carry = 1
	}

	result := int16(z.A) - int16(immediate) - int16(carry)

	z.Z = (uint8(result) == 0)                                    // Z: Zero flag se o resultado for 0
	z.N = true                                                    // N: Sempre true porque é subtração
	z.HF = (int16(z.A&0xF)-int16(immediate&0xF))-int16(carry) < 0 // H: Half Carry
	z.CF = result < 0                                             // C: Carry flag se houve underflow

	z.A = uint8(result)
	z.setFlags()

	z.PC += 2
	z.M = 8
}

// 0xDF - RST 18H
func (z *Z80) RST_18H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0018
	z.PC = 0x0018
	z.M = 16
}

// 0xE0 - LD address a8, A
func (z *Z80) LDH_a8_A() {
	immediate := z.readMemory(z.PC + 1)

	address := uint16(0xFF00) + uint16(immediate)

	z.Memory.WriteByte(address, z.A)

	z.PC += 2
	//z.M = 12
	z.M = 8

}

// 0xE1 - POP HL (Pop value from stack to HL)
func (z *Z80) POP_HL() {
	// Desempilhar um valor de 16 bits da pilha (SP) para HL
	lowByte := uint16(z.readMemory(z.SP))         //L
	highByte := uint16(z.readMemory(z.SP+1)) << 8 //H

	z.HL = lowByte | highByte
	z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
	z.L = uint8(z.HL & 0xFF) // Obtém o byte menos significativo (low byte)
	z.SP += 2

	z.PC++
	z.M = 12
}

// 0xE2 - LD (C), A
func (z *Z80) LD_C_addr_A() {
	address := uint16(0xFF00) + uint16(z.C) // Calcular o endereço baseado em 0xFF00 + valor do registrador C

	z.Memory.WriteByte(address, z.A) // Armazenar o valor do registrador A no endereço calculado

	z.PC++
	//z.PC += 2
	z.M = 8
	fmt.Printf("0xE2  Address: 0x%X  Value: 0x%X\n", address, z.A)
}

// 0xE5 - PUSH HL
func (z *Z80) PUSH_HL() {
	z.Memory.WriteByte(z.SP-1, byte(z.H))
	z.Memory.WriteByte(z.SP-2, byte(z.L))
	z.SP -= 2

	z.PC++
	z.M = 16
}

// 0xE6 - AND n
func (z *Z80) AND_n() {
	immediate := z.readMemory(z.PC + 1)
	z.updateFlagsAnd(immediate)

	z.PC += 2
	z.M = 8
}

// 0xE7 - RST 20H
func (z *Z80) RST_20H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0018
	z.PC = 0x0020
	z.M = 16
}

// 0xE8 - ADD SP, r8
func (z *Z80) ADD_SP_r8() {
	immediate := int8((z.readMemory(z.PC + 1)))
	total := uint16(int32(z.SP) + int32(immediate))
	tmpVal := z.SP ^ uint16(immediate) ^ total

	z.SP = total

	z.Z = false
	z.N = false
	z.HF = (tmpVal & 0x10) == 0x10
	z.CF = (tmpVal & 0x100) == 0x100

	z.setFlags()

	z.PC += 2
	z.M = 16

}

// 0xE9 - JP (HL)
func (z *Z80) JP_HL_addr() {
	z.PC = z.HL
	z.M = 4
}

// 0xEA - LD nn, A
func (z *Z80) LD_nn_A() {
	address := uint16(z.readMemory(z.PC+1)) | (uint16(z.readMemory(z.PC+2)) << 8)

	z.Memory.WriteByte(address, z.A)

	z.PC += 3
	//z.M = 16
	z.M = 12
}

// 0xEE - XOR d8 (Exclusive OR immediate 8-bit with A)
func (z *Z80) XOR_d8() {
	immediate := z.readMemory(z.PC + 1)
	z.updateFlagsXor(immediate)

	z.PC += 2
	z.M = 8
}

// 0xEF - RST 28H
func (z *Z80) RST_28H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salta para o endereço 0x0028
	z.PC = 0x0028
	z.M = 16
}

// 0xF0 - LDH A a8
func (z *Z80) LDH_A_a8() {
	immediate := z.readMemory(z.PC + 1)

	// Calcula o endereço completo: 0xFF00 + a8
	address := uint16(0xFF00) + uint16(immediate)

	z.A = z.readHighRam(address)

	z.setAF()

	z.PC += 2
	z.M = 12

}

// 0xF1 - POP AF (Pop value from stack to AF)
func (z *Z80) POP_AF() {
	lowByte := uint16(z.readMemory(z.SP))         //F
	highByte := uint16(z.readMemory(z.SP+1)) << 8 //A

	z.AF = lowByte | highByte
	z.A = uint8(z.AF >> 8) // Obtém o byte mais significativo (high byte)

	z.Z = lowByte&0x80 != 0
	z.N = lowByte&0x40 != 0
	z.HF = lowByte&0x20 != 0
	z.CF = lowByte&0x10 != 0
	z.setFlags()

	z.SP += 2

	z.PC++
	z.M = 12
}

// 0xF2 - LD A, (C)
func (z *Z80) LD_A_C_addr() {
	address := uint16(0xFF00) + uint16(z.C)
	z.A = z.Memory.ReadByte(address)
	z.setAF()

	z.PC++
	z.M = 8
}

// 0xF3 - DI (Disable interrupt)
func (z *Z80) DI() {
	z.IME = false

	z.PC++
	z.M = 4
}

// 0xF5 - PUSH AF
func (z *Z80) PUSH_AF() {
	z.Memory.WriteByte(z.SP-1, byte(z.A))
	z.Memory.WriteByte(z.SP-2, byte(z.F))
	z.SP -= 2

	z.PC++
	z.M = 16

}

// 0xF6 - OR d8 (Logical OR immediate with A)
func (z *Z80) OR_d8() {
	immediate := z.readMemory(z.PC + 1)
	z.updateFlagsOr(immediate)

	z.PC += 2
	z.M = 8
}

// 0xF7 - RST 30H
func (z *Z80) RST_30H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salta para o endereço 0x0030
	z.PC = 0x0030
	z.M = 16
}

// 0xF8 - LD HL, SP+r8 (Load HL with SP plus signed 8-bit offset)
func (z *Z80) LD_HL_SP_r8() {
	displacement := int8(z.readMemory(z.PC + 1))

	total := uint16(int32(z.SP) + int32(displacement))

	z.HL = total

	z.H = uint8(z.HL >> 8)
	z.L = uint8(z.HL & 0xFF)

	z.Z = false
	z.N = false

	// Flag HF: Ativada se houver carry do bit 3 para o bit 4
	z.HF = ((z.SP & 0x0F) + (uint16(displacement) & 0x0F)) > 0x0F

	// Flag CF: Ativada se houver carry do bit 7 para o bit 8
	z.CF = ((z.SP & 0xFF) + (uint16(displacement) & 0xFF)) > 0xFF

	z.setFlags()

	z.PC += 2
	z.M = 12
	//Talvez esteja errado
}

// 0xF9 - LD SP, HL (Load HL into SP)
func (z *Z80) LD_SP_HL() {
	z.SP = z.HL

	z.PC++
	z.M = 8
}

// 0xFA - LD A, nn
func (z *Z80) LD_A_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	z.A = z.readMemory(address)
	z.setAF()

	z.PC += 3
	z.M = 16
}

// 0xFB - EI (Enable Interrupts)
func (z *Z80) EI() {
	z.InterruptsEnabling = true // Habilita as interrupções após o próximo ciclo de instrução

	z.PC++
	z.M = 4
}

// 0xFE - CP n
func (z *Z80) CP_n() {
	immediate := z.readMemory(z.PC + 1)

	result := z.A - immediate

	z.Z = result == 0                        // Zero flag (definido se resultado é zero)
	z.N = true                               // Flag Negativo é sempre definido
	z.HF = (z.A & 0x0F) < (immediate & 0x0F) // Half-Carry flag
	z.CF = z.A < immediate                   // Carry flag (definido se A < n)

	z.setFlags()

	z.PC += 2
	z.M = 8
}

// 0xFF - RST 38H
func (z *Z80) RST_38H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0038
	z.PC = 0x0038
	z.M = 16
}
