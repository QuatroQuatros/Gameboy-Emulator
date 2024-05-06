package gb

import (
	"fmt"
	"gameboy/bits"
	"os"
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

// 0X00
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

// 0x03 INC BC
func (z *Z80) INC_BC() {
	z.BC++

	z.B = uint8(z.BC >> 8)
	z.C = uint8(z.BC & 0xFF)

	z.PC++
	z.M = 8
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

// 0x0C
func (z *Z80) INC_C() {
	z.C++
	z.setBC()

	z.updateFlagsInc(z.C)

	z.PC++
	z.M = 4

}

// 0x0D
func (z *Z80) DEC_C() {
	z.C--
	z.setBC()

	z.Z = z.C == 0
	z.N = true
	z.HF = (z.C & 0x0F) == 0x0F
	z.setFlags()

	z.PC++
	z.M = 4

	// var keys []int
	// for key := range z.map32 {
	//     keys = append(keys, key)
	// }
	// sort.Ints(keys)

	// for _, key := range keys {
	//     value := z.map32[key]
	//     fmt.Printf("Ciclo: %d, Valor: 0x%X\n", key, value)
	// }
	// fmt.Scanln()

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

// 0x13 INC DE
func (z *Z80) INC_DE() {
	z.DE++

	z.D = uint8(z.DE >> 8)
	z.E = uint8(z.DE & 0xFF)

	z.PC++
	z.M = 8
}

// 0x14
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

// 0x18
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

// 0x1C
func (z *Z80) INC_E() {
	z.E++
	z.setDE()
	z.updateFlagsInc(z.E)

	z.PC++
	z.M = 4

}

// 0x20
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

	// fmt.Println("0x20")
	// fmt.Scanln()
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

// 0x28
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

// 0x2F - CPL (Complement Accumulator)
func (z *Z80) CPL() {
	z.A = ^z.A // Complemento de dois no registrador A
	z.N = true
	z.HF = true
	z.setAF()

	z.PC++
	z.M = 4

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

// 0x3E
func (z *Z80) LD_A_d8() {
	immediate := z.readMemory(z.PC + 1)

	z.A = immediate
	z.setAF()

	z.PC += 2
	z.M = 8
}

// 0x47 - LD B, A
func (z *Z80) LD_B_A() {
	z.B = z.A // Carrega o valor do registrador A no registrador B
	z.setBC()

	z.PC++
	z.M = 4

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

	// Incrementar HL
	// z.HL++

	// z.H = uint8(z.HL >> 8)
	// z.L = uint8(z.HL & 0xFF)

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

// 0xA9
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

// 0xAF
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

// 0xCD
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

// 0xE0
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

	z.PC++
	z.M = 8
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

// 0xEA
func (z *Z80) LD_nn_A() {
	address := uint16(z.readMemory(z.PC+1)) | (uint16(z.readMemory(z.PC+2)) << 8)

	z.Memory.WriteByte(address, z.A)

	z.PC += 3
	z.M = 16
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

	fmt.Println("0xF0")
	fmt.Printf("ADDR: 0x%X valor: 0x%X  AF:0x%X\n", address, z.readMemory(address), z.AF)
	// fmt.Println(z.Memory.Hram)
	//fmt.Scanln()

	z.PC += 2
	z.M = 12

}

// 0xF1 - POP AF (Pop value from stack to AF)
func (z *Z80) POP_AF() {
	lowByte := uint16(z.readMemory(z.SP))         //F
	highByte := uint16(z.readMemory(z.SP+1)) << 8 //A

	z.AF = lowByte | highByte
	z.A = uint8(z.AF >> 8)   // Obtém o byte mais significativo (high byte)
	z.F = uint8(z.AF & 0xFF) // Obtém o byte menos significativo (low byte)
	z.SP += 2

	z.PC++
	z.M = 12
}

// 0xF3
func (z *Z80) DI() {
	z.IME = false
	//gb.interruptsOn = false
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

// 0xF9 - LD SP, HL (Load HL into SP)
func (z *Z80) LD_SP_HL() {
	z.SP = z.HL

	z.PC++
	z.M = 8
}

// 0xFA
func (z *Z80) LD_A_nn() {
	lowByte := uint16(z.readMemory(z.PC + 1))
	highByte := uint16(z.readMemory(z.PC + 2))
	address := (highByte << 8) | lowByte

	z.A = z.readMemory(address)
	z.setAF()

	z.PC += 3
	z.M = 13
}

// 0xFB - EI (Enable Interrupts)
func (z *Z80) EI() {
	z.IME = true // Habilita as interrupções após o próximo ciclo de instrução
	z.PC++
	z.M = 4
}

// 0xFE
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

// 0xFF
func (z *Z80) RST_38H() {
	z.SP -= 2
	z.Memory.WriteWord(z.SP, z.PC+1)

	// Salto para o endereço 0x0038
	z.PC = 0x0038

	z.M = 16
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
	case 0x27:
		z.DAA()
	case 0x28:
		z.JR_Z_e()
	case 0x2A:
		z.LDI_A_HL()
	case 0x2C:
		z.INC_L()
	case 0x2F:
		z.CPL()
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
	case 0x47:
		z.LD_B_A()
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
	case 0xAF:
		z.XOR_A()
	case 0xB0:
		z.OR_B()
	case 0xB1:
		z.OR_C()
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
		// fmt.Printf("Opcode não suportado: " , opcode)
		_, err := fmt.Scanln()
		if err != nil {
			fmt.Println("Erro ao aguardar entrada:", err)
			os.Exit(1)
		}
		z.PC++
		z.M = 4
	}
}

func (z *Z80) ExecuteCBInstruction() {
	opcodeCB := z.readMemory(z.PC + 1)
	z.PC += 2

	switch opcodeCB {
	case 0x37:
		// SWAP A
		z.A = (z.A >> 4) | (z.A << 4)
		z.setAF()
		z.Z = z.A == 0
		z.N = false
		z.HF = false
		z.CF = false
		z.setFlags()
		z.M = 8

	// case 0x86:
	//     // RES 0, (HL)
	//     address := z.HL
	//     value := z.Memory.ReadByte(address)
	//     value &^= (1 << 0) // Clear bit 0
	//     z.Memory.WriteByte(address, value)
	//     z.M = 16
	case 0x87:
		// RES 0, A
		z.A &^= (1 << 0) // Clear bit 0 of A
		z.setAF()
		z.Z = z.A == 0
		z.N = false
		z.HF = false
		z.CF = false
		z.setFlags()
		z.M = 8
	// case 0x88:
	//     // RES 0, B
	//     z.B &^= (1 << 0) // Clear bit 0 of B
	//     z.Z = z.B == 0
	//     z.N = false
	//     z.HF = false
	//     z.CF = false
	//     z.M = 8
	// case 0x89:
	//     // RES 0, C
	//     z.C &^= (1 << 0) // Clear bit 0 of C
	//     z.Z = z.C == 0
	//     z.N = false
	//     z.HF = false
	//     z.CF = false
	//     z.M = 8

	// Outras instruções CB...

	default:
		fmt.Printf("Opcode CB não suportado: 0xCB%x\n", opcodeCB)
		fmt.Scanln()
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
	//fmt.Scanln() // Aguardar entrada do usuário (pressionar Enter)
	// fmt.Printf("Hram 0x44:  0x%X\n", z.Memory.Hram[0x44])

	z.ExecuteInstruction(opcode)

	return z.M

}
