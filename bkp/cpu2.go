package gb

import (
    "fmt"
    "os"
)

type Z80 struct {
    A, B, C, D, E, H, L, F byte // Registradores de 8 bits
    AF, BC, DE, HL uint16   // Pares de registradores de 16 bits
    PC, SP uint16           // Contador de programa (PC) e ponteiro de pilha (SP)
    
    Z, N, HF, CF bool // Flags de status (Zero, Negativo, Meio-carro, Carry, Interrupt Master Enable)

    IME bool
    
    M int // Contador de ciclos de máquina

    Divider int

    Memory *Memory
}

func (z *Z80) SpHiLo() uint16{
    return z.SP
}

// func (z *Z80) setFlag(index byte, on bool) {
// 	if on {
//         //value = uint8(z.AF & 0xFF)
//         //z.F // AF lower byte
//         z.F = z.F | (1 << index)
// 		z.setLowAF()
// 	} else {
//         z.F = z.F & ^(1 << index)
// 		z.setLowAF()
// 	}
// }


// SetZ sets the value of the Z flag.
// func (z *Z80) setFlagZ(on bool) {
// 	z.setFlag(7, z.Z)
// }

// // SetN sets the value of the N flag.
// func (cpu *CPU) SetN(on bool) {
// 	cpu.setFlag(6, on)
// }

// // SetH sets the value of the H flag.
// func (cpu *CPU) SetH(on bool) {
// 	cpu.setFlag(5, on)
// }

// // SetC sets the value of the C flag.
// func (cpu *CPU) SetC(on bool) {
// 	cpu.setFlag(4, on)
// }


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

//-------------------- SET High Byte ----------------------------------

func (z *Z80) setHighAF() {
    z.AF = uint16(z.A)<<8 | (uint16(z.F) & 0xFF)
}

func (z *Z80) setHighBC() {
    z.BC = uint16(z.B)<<8 | (uint16(z.C) & 0xFF)
}

func (z *Z80) setHighDE() {
    z.DE = uint16(z.D)<<8 | (uint16(z.E) & 0xFF)
}

func (z *Z80) setHighHL() {
    z.HL = uint16(z.H)<<8 | (uint16(z.L) & 0xFF)
}

//-------------------- SET Low Byte ----------------------------------
// func (z *Z80) setLowAF() {
//     z.AF = (uint16(z.AF) & 0xFF00) | (uint16(z.F))
// }

func (z *Z80) setLowBC() {
    z.BC = (uint16(z.BC) & 0xFF00) | (uint16(z.C))
}

func (z *Z80) setLowDE() {
    z.DE = (uint16(z.DE) & 0xFF00) | (uint16(z.E))
}

func (z *Z80) setLowHL() {
    z.HL = (uint16(z.HL) & 0xFF00) | (uint16(z.L))
}

func (z *Z80) Init(memory *Memory) {
    // if cgb {
	// 	z.AF.Set(0x1180)
	// } else {
	// 	z.AF.Set(0x01B0)
	// }

	z.PC = 0x0100
    z.SP = 0xFFFE

    z.A = 0x01
    z.F = 0xB0
    z.B = 0x00
    z.C = 0x13
    z.D = 0x00
    z.E = 0xD8
    z.H = 0x01
    z.L = 0x4D

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
    // Z - Definido se o resultado for zero
    z.Z = value == 0

    // N - Resetado
    z.N = false

    // HF - Definido se houve carry do bit 3 para o bit 4
    z.HF = (value & 0x0F) == 0 // Verifica se o lower nibble é zero
}

func (z *Z80) updateFlagsDec(value byte) {
    z.Z = value == 0

    // N - Definido após operação de decremento
    z.N = true

    // HF - Definido se ocorrer um empréstimo do bit 4 durante o decremento
    z.HF = (value & 0x0F) == 0x0F // Verifica se o lower nibble é 0x0F (decremento de 1)
}

//0X00
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
    z.B = uint8(value >> 8)   // Obtém o byte mais significativo (high byte)
    z.C = uint8(value & 0xFF)  // Obtém o byte menos significativo (low byte)

    z.PC += 3
    z.M = 12
}

// 0x05 - DEC B
func (z *Z80) DEC_B() {
    z.B-- 
    z.setHighBC()

    z.Z = z.B == 0     
    z.N = true         
    z.HF = (z.B & 0x0F) == 0x0F  

    z.PC++ 
    z.M = 4 
}

// 0x06 - LD B, d8
func (z *Z80) LD_B_d8() {
    immediate := z.Memory.ReadByte(z.PC + 1)

    z.B = immediate
    z.setHighBC()

    z.M = 8
    z.PC += 2
}

//0x0B - DEC BC
func (z *Z80) DEC_BC() {
    z.BC--

    z.B = uint8(z.BC >> 8)   // Obtém o byte mais significativo (high byte)
    z.C = uint8(z.BC & 0xFF)  // Obtém o byte menos significativo (low byte)

    z.M = 8
    z.PC++
}

//0x0C
func (z *Z80) INC_C() {
    z.C++
    z.setLowBC()

    z.updateFlagsInc(z.C)

    z.M = 4
    z.PC++

}


//0x0D
func (z *Z80) DEC_C() {
    z.C--
    z.setLowBC()

    z.Z = z.B == 0     
    z.N = true         
    z.HF = (z.C & 0x0F) == 0x0F  

    z.M = 4
    z.PC++

}

// 0x0E - LD C, d8
func (z *Z80) LD_C_d8() {
    // Lê o byte imediatamente seguinte ao PC para obter o valor de 8 bits (d8)
    immediate := z.Memory.ReadByte(z.PC + 1)

    z.C = immediate
    z.setLowBC()

    z.M = 8
    z.PC += 2
}

// 0x16 - LD D, d8 (Load 8-bit immediate value into D)
func (z *Z80) LD_D_d8() {
    immediate := z.Memory.ReadByte(z.PC + 1) 

    z.D = immediate

    z.setDE() 

    z.PC += 2  
    z.M = 8    
}


//0x18
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

    z.PC++
    z.M = 8
}

//0x20
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
    value := (highByte << 8) | lowByte

    z.HL = value

    z.H = uint8(value >> 8)   // Obtém o byte mais significativo (high byte)
    z.L = uint8(value & 0xFF)  // Obtém o byte menos significativo (low byte)

    z.PC += 3
    z.M = 12
}

// 0x22 - LD (HL+), A
func (z *Z80) LD_HL_inc_A() {
    address := z.HL

    z.Memory.WriteByte(address, z.A)

    z.HL++

    z.H = uint8(z.HL >> 8) 
    z.L = uint8(z.HL & 0xFF)  

    z.M = 8
    z.PC++
}

// 0x23 - INC HL
func (z *Z80) INC_HL() {
    z.HL++

    z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
    z.L = uint8(z.HL & 0xFF)  // Obtém o byte menos significativo (low byte)

    //TEM ISSO?
    //z.updateFlagsInc(z.HL)

    z.M = 8
    z.PC++
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

    z.setHighAF()

    z.PC++
    z.M = 4
}

//0x28
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
    address := z.HL
    z.A = z.Memory.ReadByte(address)
    z.setHighAF()

    z.HL++

    z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
    z.L = uint8(z.HL & 0xFF)  // Obtém o byte menos significativo (low byte)

    z.M = 8
    z.PC++
}

// 0x2F - CPL (Complement Accumulator)
func (z *Z80) CPL() {
    z.A = ^z.A // Complemento de dois no registrador A
    z.N = true 
    z.HF = true 
    z.setHighAF() 

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
    address := z.HL

    z.Memory.WriteByte(address, z.A)

    z.HL--

    z.H = uint8(z.HL >> 8)   // Obtém o byte mais significativo (high byte)
    z.L = uint8(z.HL & 0xFF)  // Obtém o byte menos significativo (low byte)

    z.PC++
    z.M = 8
}

// 0x36 - LD (HL), d8
func (z *Z80) LD_HL_d8() {
    immediate := z.readMemory(z.PC + 1)

    address := z.HL

    z.Memory.WriteByte(address, immediate)

    z.PC += 2
    z.M = 12
}

//0x3C - INC A
func (z *Z80) INC_A() {
    z.A++
    z.setHighAF()

    z.updateFlagsInc(z.A)

    z.M = 4
    z.PC++

}

//0x3E
func (z *Z80) LD_A_d8() {
    immediate := z.readMemory(z.PC + 1)

    z.A = immediate
    z.setHighAF()

    z.PC += 2
    z.M = 8
}

// 0x47 - LD B, A
func (z *Z80) LD_B_A() {
    z.B = z.A // Carrega o valor do registrador A no registrador B

    z.setHighBC()

    z.M = 4
    z.PC++
}

// 0x4F - LD C, A
func (z *Z80) LD_C_A() {
    z.C = z.A 
    z.setLowBC() 

    z.PC++ 
    z.M = 4 
}

// 0x57 - LD D, A
func (z *Z80) LD_D_A() {
    z.D = z.A  
    z.setHighDE()

    z.PC++  
    z.M = 4 
}

// 0x5F - LD E, A (Load A into E)
func (z *Z80) LD_E_A() {
    z.E = z.A // Carrega o valor do registrador A no registrador E

    z.setDE() // Atualiza o par de registradores DE

    z.PC++   // Incrementa o contador de programa (PC)
    z.M = 4  // Define o número de ciclos de máquina
}

// 0x78 - LD A, B
func (z *Z80) LD_A_B() {
    z.A = z.B  // Carrega o valor do registrador B no registrador A (Acumulador)
    z.setHighAF()

    z.PC++ 
    z.M = 4
}

// 0x79 - LD A, C
func (z *Z80) LD_A_C() {
    z.A = z.C  // Carrega o valor do registrador C no registrador A (Acumulador)
    z.setHighAF()

    z.PC++ 
    z.M = 4
}

// 0x7A - LD A, D
func (z *Z80) LD_A_D() {
    z.A = z.D  // Carrega o valor do registrador D no registrador A (Acumulador)
    z.setHighAF()

    z.PC++ 
    z.M = 4
}

// 0x7B - LD A, E
func (z *Z80) LD_A_E() {
    z.A = z.E  // Carrega o valor do registrador E no registrador A (Acumulador)
    z.setHighAF()

    z.PC++ 
    z.M = 4
}

// 0x7C - LD A, H
func (z *Z80) LD_A_H() {
    z.A = z.H  // Carrega o valor do registrador H no registrador A (Acumulador)
    z.setHighAF()

    z.PC++ 
    z.M = 4
}

// 0x7D - LD A, L
func (z *Z80) LD_A_L() {
    z.A = z.L  // Carrega o valor do registrador L no registrador A (Acumulador)
    z.setHighAF()

    z.PC++ 
    z.M = 4
}

// 0x7E - LD A, (HL)
func (z *Z80) LD_A_HL() {
    address := z.HL
    z.A = z.Memory.ReadByte(address)
    z.setHighAF()

    z.M = 8
    z.PC++
}

// 0x7F - LD A, A
func (z *Z80) LD_A_A() {
    z.PC++ 
    z.M = 4
}

// 0x87 - RES 0, A (Clear bit 0 of A)
func (z *Z80) RES_0_A() {
    z.A &^= 0x01  // Zera o bit 0 do registrador A

    z.setHighAF() // Atualiza o par de registradores AF

    z.Z = z.A == 0   // Define a flag Z se A for zero
    z.N = false      // Reseta a flag N
    z.HF = false     // Reseta a flag HF
    z.CF = false     // Reseta a flag CF

    z.PC++  // Incrementa o contador de programa (PC)
    z.M = 8 // Define o número de ciclos de máquina
}


// 0xA1 - AND C
func (z *Z80) AND_C() {
    z.A &= z.C 
    z.setHighAF() 

    z.Z = z.A == 0   
    z.N = false      
    z.HF = true      
    z.CF = false     

    z.PC++ 
    z.M = 4 
}

//0xA9
func (z *Z80) XOR_C() {
    z.C = 0

    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.setLowBC()

    z.PC++
    z.M = 4
}


//0xAF
func (z *Z80) XOR_A() {
    z.A = 0

    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.setHighAF()
    //.setFlagZ()

    z.PC++
    z.M = 4
}


// 0xB1 - OR C
func (z *Z80) OR_C() {
    z.A |= z.C

    z.setHighAF()

    z.Z = z.A == 0   
    z.N = false      
    z.HF = false     
    z.CF = false     

    z.PC++
    z.M = 4
}

// 0xB0 - OR B
func (z *Z80) OR_B() {
    z.A |= z.B 
    z.setHighAF() 

    z.Z = z.A == 0   
    z.N = false      
    z.HF = false     
    z.CF = false     

    z.PC++ 
    z.M = 4 
}


// 0xC3 - JP nn
func (z *Z80) JP_nn() {
    lowByte := uint16(z.readMemory(z.PC + 1))
    highByte := uint16(z.readMemory(z.PC + 2))
    targetAddress := (highByte << 8) | lowByte

    z.PC = targetAddress
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


//0xCD
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
    z.E = z.Memory.ReadByte(z.SP)
    z.D = z.Memory.ReadByte(z.SP + 1)
    z.setDE() 

    z.SP += 2 

    z.M = 12 
}

// 0xD5 - PUSH DE
func (z *Z80) PUSH_DE() {
    z.SP--
    z.Memory.WriteByte(z.SP, byte(z.DE>>8)) 

    z.SP--
    z.Memory.WriteByte(z.SP, byte(z.DE)) 

    z.PC++
    z.M = 16
}


//0xE0
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
    z.L = z.Memory.ReadByte(z.SP)   
    z.H = z.Memory.ReadByte(z.SP+1) 

    z.setHL()

    z.SP += 2

    z.PC++  
    z.M = 12
}

//0xE2
func (z *Z80) LD_c_A() {
    address := uint16(0xFF00) + uint16(z.C) // Calcular o endereço baseado em 0xFF00 + valor do registrador C

    z.Memory.WriteByte(address, z.A) // Armazenar o valor do registrador A no endereço calculado

    z.PC++ 
    z.M = 8 
}

// 0xE5 - PUSH HL
func (z *Z80) PUSH_HL() {
    z.SP--
    z.Memory.WriteByte(z.SP, byte(z.HL>>8))

    z.SP--
    z.Memory.WriteByte(z.SP, byte(z.HL))

    z.PC++
    z.M = 16
}


// 0xE6 - AND n
func (z *Z80) AND_n() {
    immediate := z.readMemory(z.PC + 1)
   
    z.A &= immediate
    z.setHighAF()

    z.Z = z.A == 0         
    z.N = false             
    z.HF = true             
    z.CF = false            

    z.PC += 2               
    z.M = 8                 
}

//0xEA
func (z *Z80) LD_nn_A() {
    address := uint16(z.readMemory(z.PC + 1)) | (uint16(z.readMemory(z.PC + 2)) << 8)

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

//0xF0
func (z *Z80) LDH_A_a8() {
    address := uint16(0xFF00) + uint16(z.readMemory(z.PC + 1))

    z.A = z.readMemory(address)

    z.setHighAF()

    z.PC += 2
    z.M = 12
}

//0xF3
func (z *Z80) DI() {
    z.IME = false
    z.PC++
    z.M = 4 
}

// 0xFB - EI (Enable Interrupts)
func (z *Z80) EI() {
    z.IME = true  // Habilita as interrupções após o próximo ciclo de instrução

    z.PC++ 
    z.M = 4 
}

//0xFE
func (z *Z80) CP_n() {
    immediate := z.readMemory(z.PC + 1)

    result := z.A - immediate

    z.Z = result == 0   // Zero flag (definido se resultado é zero)
    z.N = true          // Flag Negativo é sempre definido
    z.HF = (z.A & 0x0F) < (immediate & 0x0F)  // Half-Carry flag
    z.CF = z.A < immediate                    // Carry flag (definido se A < n)

    z.PC += 2
    z.M = 8
}


//0xFF
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
    case 0x16:
        z.LD_D_d8()
    case 0x18:
        z.JR_e()
    case 0x19:
        z.ADD_HL_DE()
    case 0x20:
        z.JR_nz_e()
    case 0x21:
        z.LD_HL_nn()
    case 0x22:
        z.LD_HL_inc_A()
    case 0x23:
        z.INC_HL()
    case 0x27:
        z.DAA()
    case 0x28:
        z.JR_Z_e()
    case 0x2A:
        z.LDI_A_HL()
    case 0x2F:
        z.CPL()
    case 0x31:
        z.LD_SP_nn()
    case 0x32:
        z.LD_HL_dec_A()
    case 0x36:
        z.LD_HL_d8()
    case 0x3C:
        z.INC_A()
    case 0x3E:
        z.LD_A_d8()
    case 0x47:
        z.LD_B_A()
    case 0x4F:
        z.LD_C_A()
    case 0x57:
        z.LD_D_A()
    case 0x5F:
        z.LD_E_A()
    case 0x78:
        z.LD_A_B()
    case 0x79:
        z.LD_A_C()
    case 0x7A:
        z.LD_A_C()
    case 0x7B:
        z.LD_A_D()
    case 0x7C:
        z.LD_A_E()
    case 0x7D:
        z.LD_A_H()
    case 0x7E:
        z.LD_A_HL()
    case 0x7F:
        z.LD_A_A()
    case 0x87:
        z.RES_0_A()
    case 0xA1:
        z.AND_C()
    case 0xA9:
        z.XOR_C()
    case 0xAF:
        z.XOR_A()
    case 0xB0:
        z.OR_B()
    case 0xB1:
        z.OR_C()
    case 0xC3:
        z.JP_nn()
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
    case 0xEA:
        z.LD_nn_A()
    case 0xEF:
        z.RST_28H()
    case 0xF0:
        z.LDH_A_a8()
    case 0xF3:
        z.DI()
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
            z.Z = z.A == 0
            z.N = false
            z.HF = false
            z.CF = false
            z.setHighAF()
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
            z.setHighAF()
            z.Z = z.A == 0
            z.N = false
            z.HF = false
            z.CF = false
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


func (z *Z80) EmulateCycle() int{
    opcode := z.readMemory(z.PC)
    fmt.Println("\n================ Registros ================")
    fmt.Printf("PC: 0x%X SP: 0x%X OP: 0x%X\n", z.PC, z.SP, opcode)
    fmt.Printf("A:  0x%X F:  0x%X AF: 0x%X\n", z.A, z.F, z.AF)
    fmt.Printf("B:  0x%X C:  0x%X BC: 0x%X\n", z.B, z.C, z.BC)
    fmt.Printf("D:  0x%X E:  0x%X DE: 0x%X\n", z.D, z.E, z.DE)
    fmt.Printf("H:  0x%X L:  0x%X HL: 0x%X\n", z.H, z.L, z.HL)
    fmt.Println("=============================================")
    //fmt.Scanln() // Aguardar entrada do usuário (pressionar Enter)
    // fmt.Printf("Hram 0x44:  0x%X\n", z.Memory.Hram[0x44])

    z.ExecuteInstruction(opcode)

    return z.M

}