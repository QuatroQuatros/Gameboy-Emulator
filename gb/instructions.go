package gb

import "fmt"

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

	z.Z = z.B == 0
	z.N = true
	z.HF = (z.B & 0x0F) == 0x0F
	z.setFlags()

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

	z.Z = z.C == 0
	z.N = true
	z.HF = (z.C & 0x0F) == 0x0F
	z.setFlags()

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
func (z *Z80) LD_DE_A() {
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

// 0x16 - LD D, d8 (Load 8-bit immediate value into D)
func (z *Z80) LD_D_d8() {
	immediate := z.Memory.ReadByte(z.PC + 1)
	z.D = immediate

	z.setDE()

	z.PC += 2
	z.M = 8
}

// 0x18 - JR r8
func (z *Z80) JR_e() {
	displacement := int8(z.readMemory(z.PC + 1))

	newAddress := uint16(int(z.PC) + 2 + int(displacement))

	z.PC = newAddress
	z.M = 12
}

// 0x19 - ADD HL, DE (Add DE to HL)
func (z *Z80) ADD_HL_DE() {
	deValue := z.DE
	hlValue := z.HL

	result := hlValue + deValue

	z.HL = result

	z.H = uint8(result >> 8)
	z.L = uint8(result & 0xFF)

	z.Z = false
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()

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

	z.Z = z.E == 0
	z.N = true
	z.HF = (z.E & 0x0F) == 0x0F
	z.setFlags()

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

	// fmt.Println("0x21")
	// fmt.Scanln()

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

	z.Z = z.H == 0
	z.N = true
	z.HF = (z.H & 0x0F) == 0x0F
	z.setFlags()

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

	z.setAF()
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
	result := z.HL + z.HL

	z.N = false
	z.HF = (z.HL & 0xFFF) > (result & 0xFFF)
	z.CF = result > 0xFFFF
	z.setFlags()

	z.HL = result

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

	z.Z = z.L == 0
	z.N = true
	z.HF = (z.L & 0x0F) == 0x0F
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0x2F - CPL (Complement Accumulator)
func (z *Z80) CPL() {
	z.A = ^z.A // Complemento de dois no registrador A
	z.N = true
	z.HF = true
	z.setAF()

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

	// z.map32[z.ciclos] = z.HL
	// z.ciclos++
	// fmt.Println("0x32")
	// fmt.Scanln()
}

// 0x34 - INC (HL)
func (z *Z80) INC_HL_addr() {
	value := z.Memory.ReadByte(z.HL)
	value++
	z.Memory.WriteByte(z.HL, value)

	// Atualizar os flags com base no resultado da operação de incremento
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

	z.PC++
	z.M = 12
}

// 0x36 - LD (HL), d8
func (z *Z80) LD_HL_d8() {
	immediate := z.readMemory(z.PC + 1)

	address := z.HL

	z.Memory.WriteByte(address, immediate)

	z.PC += 2
	z.M = 12
}

// 0x3C - INC A
func (z *Z80) INC_A() {
	z.A++
	z.setAF()

	z.updateFlagsInc(z.A)

	z.PC++
	z.M = 4

}

// 0x3D - DEC A
func (z *Z80) DEC_A() {
	z.A--
	z.setAF()

	z.Z = z.A == 0
	z.N = true
	z.HF = (z.A & 0x0F) == 0x0F
	z.setFlags()

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

// 0x46 - LD B, (HL)
func (z *Z80) LD_B_HL_addr() {
	z.B = z.Memory.ReadByte(z.HL)
	z.setBC()

	z.PC++
	z.M = 8
}

// 0x47 - LD B, A
func (z *Z80) LD_B_A() {
	z.B = z.A // Carrega o valor do registrador A no registrador B
	z.setBC()

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

// 0x5E - LD E, (HL)
func (z *Z80) LD_E_HL() {
	z.E = z.Memory.ReadByte(z.HL)
	z.setDE()

	z.PC++
	z.M = 8

	// Incrementar HL
	// z.HL++

	// z.H = uint8(z.HL >> 8)
	// z.L = uint8(z.HL & 0xFF)

}

// 0x5F - LD E, A (Load A into E)
func (z *Z80) LD_E_A() {
	z.E = z.A
	z.setDE()

	z.PC++
	z.M = 4
}

// 0x62 - LD H, D
func (z *Z80) LD_H_D() {
	z.H = z.D // Carrega o valor do registrador D no registrador H
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x67 - LD H, A
func (z *Z80) LD_H_A() {
	z.H = z.A // Carrega o valor do registrador A no registrador H
	z.setHL()

	z.PC++
	z.M = 4

	fmt.Println("---------------------------0x67---------------------------")
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
	fmt.Scanln()
}

// 0x6B - LD L, E
func (z *Z80) LD_L_E() {
	z.L = z.E // Carrega o valor do registrador E no registrador L
	z.setHL()

	z.PC++
	z.M = 4
}

// 0x6E - LD L, (HL)
func (z *Z80) LD_L_HL() {
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
func (z *Z80) LD_HL_B() {
	z.Memory.WriteByte(z.HL, z.B)

	z.PC++
	z.M = 8
}

// 0x71 - LD (HL), C
func (z *Z80) LD_HL_C() {
	z.Memory.WriteByte(z.HL, z.C)

	z.PC++
	z.M = 8
	fmt.Println("0x71")
}

// 0x72 - LD (HL), D
func (z *Z80) LD_HL_D() {
	z.Memory.WriteByte(z.HL, z.D)

	z.PC++
	z.M = 8
}

// 0x77 - LD (HL), A
func (z *Z80) LD_HL_addr_A() {
	z.Memory.WriteByte(z.HL, z.A)

	z.PC++
	z.M = 8
}

// 0x78 - LD A, B
func (z *Z80) LD_A_B() {
	z.A = z.B // Carrega o valor do registrador B no registrador A (Acumulador)
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x79 - LD A, C
func (z *Z80) LD_A_C() {
	z.A = z.C // Carrega o valor do registrador C no registrador A (Acumulador)
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7A - LD A, D
func (z *Z80) LD_A_D() {
	z.A = z.D // Carrega o valor do registrador D no registrador A (Acumulador)
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7B - LD A, E
func (z *Z80) LD_A_E() {
	z.A = z.E // Carrega o valor do registrador E no registrador A (Acumulador)
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7C - LD A, H
func (z *Z80) LD_A_H() {
	z.A = z.H // Carrega o valor do registrador H no registrador A (Acumulador)
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7D - LD A, L
func (z *Z80) LD_A_L() {
	z.A = z.L // Carrega o valor do registrador L no registrador A (Acumulador)
	z.setAF()

	z.PC++
	z.M = 4
}

// 0x7E - LD A, (HL)
func (z *Z80) LD_A_HL() {
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

// 0x81 - ADD A, C
func (z *Z80) ADD_A_C() {
	result := uint16(z.A) + uint16(z.C)

	z.A = uint8(result)

	z.Z = z.A == 0
	z.N = false
	z.HF = (z.A & 0x0F) < (z.C & 0x0F)
	z.CF = result > 0xFF
	z.setFlags()

	z.PC++
	z.M = 4
}

// // 0x87 - RES 0, A (Clear bit 0 of A)
// func (z *Z80) RES_0_A() {
//     z.A &^= 0x01  // Zera o bit 0 do registrador A

//     z.setAF() // Atualiza o par de registradores AF

//     z.Z = z.A == 0   // Define a flag Z se A for zero
//     z.N = false      // Reseta a flag N
//     z.HF = false     // Reseta a flag HF
//     z.CF = false     // Reseta a flag CF

//     z.setFlags()

//     z.PC++  // Incrementa o contador de programa (PC)
//     z.M = 8 // Define o número de ciclos de máquina
// }

// 0x91 - SUB C
func (z *Z80) SUB_C() {
	result := uint16(z.A) - uint16(z.C)

	z.A = uint8(result)

	z.Z = z.A == 0
	z.N = true
	z.HF = int16(z.A&0x0F)-int16(z.C&0x0F) < 0 // Verifica se houve empréstimo de meio byte (half carry)
	z.CF = result > 0xFF                       // Seta a flag de carry se houver overflow
	z.setFlags()

	z.PC++
	z.M = 4

}

// 0xA1 - AND C
func (z *Z80) AND_C() {
	z.A &= z.C
	z.setAF()

	z.Z = z.A == 0
	z.N = false
	z.HF = true
	z.CF = false

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xA7 - AND A
func (z *Z80) AND_A() {
	z.A &= z.A
	z.setAF()

	z.Z = z.A == 0
	z.N = false
	z.HF = true
	z.CF = false

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xA9 - XOR C
func (z *Z80) XOR_C() {
	z.A ^= z.C
	z.setAF()

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xAE - XOR (HL)
func (z *Z80) XOR_HL_addr() {
	z.A ^= z.readMemory(z.HL)

	// Atualiza as flags
	z.Z = z.A == 0 // Z é verdadeiro se A for zero
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()

	z.PC++
	z.M = 8
}

// 0xAF - XOR A
func (z *Z80) XOR_A() {
	z.A ^= z.A
	z.setAF()

	z.Z = true
	z.N = false
	z.HF = false
	z.CF = false

	z.setAF()
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xB0 - OR B
func (z *Z80) OR_B() {
	z.A |= z.B
	z.setAF()

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xB1 - OR C
func (z *Z80) OR_C() {
	z.A |= z.C
	z.setAF()

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false

	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xB6 - OR (HL)
func (z *Z80) OR_HL_addr() {
	value := z.Memory.ReadByte(z.HL)
	z.A |= value

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false

	z.setFlags()

	z.PC++
	z.M = 8
}

// 0xB7 - OR A
func (z *Z80) OR_A() {
	z.A |= z.A

	z.Z = z.A == 0 // Z será verdadeiro se A for zero
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xB8 - CP B (Compare B with A)
func (z *Z80) CP_B() {
	// Realiza a operação de subtração de C de A
	result := uint16(z.A) - uint16(z.B)

	z.Z = result == 0
	z.N = true
	z.HF = (z.B & 0x0F) > (z.A & 0x0F)
	z.CF = z.B > z.A
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xB9 - CP C (Compare C with A)
func (z *Z80) CP_C() {
	// Realiza a operação de subtração de C de A
	result := uint16(z.A) - uint16(z.C)

	z.Z = result == 0
	z.N = true
	z.HF = (z.C & 0x0F) > (z.A & 0x0F)
	z.CF = z.C > z.A
	z.setFlags()

	z.PC++
	z.M = 4

	// fmt.Println("---------------------------0xB9---------------------------")
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

// 0xBA - CP D (Compare D with A)
func (z *Z80) CP_D() {
	// Realiza a operação de subtração de D de A
	result := uint16(z.A) - uint16(z.D)

	z.Z = result == 0
	z.N = true
	z.HF = (z.D & 0x0F) > (z.A & 0x0F)
	z.CF = z.D > z.A
	z.setFlags()

	z.PC++
	z.M = 4
}

// 0xBB - CP E (Compare E with A)
func (z *Z80) CP_E() {
	// Realiza a operação de subtração de E de A
	result := uint16(z.A) - uint16(z.E)

	z.Z = result == 0
	z.N = true
	z.HF = (z.E & 0x0F) > (z.A & 0x0F)
	z.CF = z.E > z.A
	//z.CF = result > 0xFF
	//z.A = uint8(result)
	z.setFlags()

	z.PC++
	z.M = 4

	// fmt.Println("---------------------------0xBB---------------------------")
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

	//result pode ser um valor de 16bits
	result := uint16(z.A) + uint16(immediate)
	//A só suporta 8 bits
	z.A = byte(result)

	z.Z = z.A == 0
	z.N = false
	z.HF = (z.A & 0x0F) < 0x0F
	z.CF = result > 0xFF

	z.setFlags()

	z.PC += 2
	z.M = 8
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
	carry := uint16(0)
	if z.CF {
		carry = 1
	}

	result := uint16(z.A) + uint16(immediate) + carry

	z.Z = byte(result) == 0
	z.N = false
	z.HF = ((z.A & 0xF) + (immediate & 0x0F) + byte(carry)) > 0x0F
	z.CF = result > 0xFF
	z.A = byte(result)
	z.setFlags()

	z.PC += 2
	z.M = 8
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

	// fmt.Println("---------------------------0xD1---------------------------")
	// opcode := z.readMemory(z.PC)
	// fmt.Println("\n================ Registros ================")
	// fmt.Printf("lowByte: %X highByte: %X\n", lowByte, highByte)
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

	result := uint16(z.A) - uint16(immediate)
	z.A = byte(result)

	// Define as flags baseadas no resultado da subtração
	z.Z = z.A == 0                             // Z será verdadeiro se A for zero
	z.N = true                                 // N será verdadeiro (operação de subtração)
	z.HF = (z.A & 0x0F) > (z.A-immediate)&0x0F // HF será verdadeiro se houve borrow de meio-carro
	z.CF = result > 0xFF                       // CF será verdadeiro se houve borrow de 8 bits
	z.setFlags()

	z.PC += 2
	z.M = 8

	//SUS dessa função
	//fmt.Println("0xD6")
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
	z.PC = newPC

	z.IME = true
	z.SP += 2
	z.M = 16

}

// 0xE0 - LD address a8, A
func (z *Z80) LDH_a8_A() {
	immediate := z.readMemory(z.PC + 1)

	address := uint16(0xFF00) + uint16(immediate)

	z.Memory.WriteByte(address, z.A)

	z.PC += 2
	z.M = 12
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
func (z *Z80) LD_c_A() {
	address := uint16(0xFF00) + uint16(z.C) // Calcular o endereço baseado em 0xFF00 + valor do registrador C

	z.Memory.WriteByte(address, z.A) // Armazenar o valor do registrador A no endereço calculado

	z.PC += 2
	z.M = 8

	fmt.Println("---------------------------0xE2---------------------------")
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
	// fmt.Scanln()
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

	z.A &= immediate
	z.setAF()

	z.Z = z.A == 0
	z.N = false
	z.HF = true
	z.CF = false

	z.setFlags()

	z.PC += 2
	z.M = 8
}

// 0xE9 - JP (HL)
func (z *Z80) JP_HL() {
	z.PC = z.HL
	z.M = 4
}

// 0xEA - LD nn, A
func (z *Z80) LD_nn_A() {
	address := uint16(z.readMemory(z.PC+1)) | (uint16(z.readMemory(z.PC+2)) << 8)

	z.Memory.WriteByte(address, z.A)

	z.PC += 3
	z.M = 16
}

// 0xEE - XOR d8 (Exclusive OR immediate 8-bit with A)
func (z *Z80) XOR_d8() {
	immediate := z.readMemory(z.PC + 1)
	z.A ^= immediate

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()

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

	// Carrega o valor do endereço de memória (0xFF00 + a8) para o registrador A
	z.A = z.readMemory(address)

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
	//z.F = uint8(z.AF & 0xFF) // Obtém o byte menos significativo (low byte)
	z.F = uint8(z.AF & 0xF0) // Obtém o byte menos significativo (low byte)

	z.Z = z.F&0x80 != 0
	z.N = z.F&0x40 != 0
	z.HF = z.F&0x20 != 0
	z.CF = z.F&0x10 != 0
	z.setFlags()

	z.SP += 2

	z.PC++
	z.M = 12

	fmt.Println("---------------------------0xF1---------------------------")
	fmt.Println("\n================ Registros ================")
	fmt.Printf("Low:  0x%X  High:  0x%X\n", lowByte, highByte)
	fmt.Println("=============================================")
	// fmt.Scanln()
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
	z.A |= immediate

	z.Z = z.A == 0
	z.N = false
	z.HF = false
	z.CF = false
	z.setFlags()

	z.PC += 2
	z.M = 8
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

	z.HF = ((z.SP&0x0F)+uint16(displacement))&0x10 != 0
	z.CF = ((z.SP&0xFF)+uint16(displacement))&0x100 != 0

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
	z.IME = true // Habilita as interrupções após o próximo ciclo de instrução
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

// 0xFF - RST 0x0038
func (z *Z80) RST_38H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0038
	z.PC = 0x0038

	z.M = 16
}
