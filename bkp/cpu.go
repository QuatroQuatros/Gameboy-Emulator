package gb

import (
    "fmt"
    "os"
)

type Z80 struct {
    A, B, C, D, E, H, L byte // Registradores de 8 bits
    AF, BC, DE, HL uint16   // Pares de registradores de 16 bits
    PC, SP uint16           // Contador de programa (PC) e ponteiro de pilha (SP)
    
    Z, N, HF, CF, IME bool // Flags de status (Zero, Negativo, Meio-carro, Carry, Interrupt Master Enable)
    
    M int // Contador de ciclos de máquina

    Divider int

    Memory *Memory
}

func (cpu *Z80) SpHiLo() uint16{
    return cpu.SP
}

func (cpu *Z80) Init(memory *Memory) {
	cpu.PC = 0x0100
    cpu.SP = 0xFFFE
	// if cgb {
	// 	cpu.AF.Set(0x1180)
	// } else {
	// 	cpu.AF.Set(0x01B0)
	// }

//  8 BITS
    cpu.A = 0x01
    cpu.B = 0x00
    cpu.C = 0x13
    cpu.D = 0x00
    cpu.E = 0xD8
    cpu.H = 0x01
    cpu.L = 0x4D


//  16 BITS
    cpu.AF = 0x01B0
	cpu.BC = 0x0000
	cpu.DE = 0xFF56
	cpu.HL = 0x000D
    cpu.Memory = memory

    //cpu.AF.mask = 0xFFF0
}

func (z *Z80) readMemory(addr uint16) byte {
    if z.Memory == nil {
        fmt.Println("Erro: Memória não inicializada")
        return 0xFF // Retornar valor padrão em caso de memória não inicializada
    }
    
    return z.Memory.ReadByte(addr)
}

func (z *Z80) setHL(){
    // Combina os registradores H e L em um valor de 16 bits (HL)
    z.HL = (uint16(z.H) << 8) | uint16(z.L)
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

    z.PC += 3
    z.M = 12
}

// 0x05 - DEC B
func (z *Z80) DEC_B() {
    z.B-- 

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

    z.M = 8
    z.PC += 2
}

//0x0B - DEC BC
func (z *Z80) DEC_BC() {
    z.BC--

    z.M = 8
    z.PC++
}

//0x0C
func (z *Z80) INC_C() {
    z.C++

    z.updateFlagsInc(z.C)

    z.M = 4
    z.PC++

}

//0x0D
func (z *Z80) DEC_C() {
    z.C--

    z.updateFlagsDec(z.C)

    z.M = 4
    z.PC++

}


// 0x0E - LD C, d8
func (z *Z80) LD_C_d8() {
    // Lê o byte imediatamente seguinte ao PC para obter o valor de 8 bits (d8)
    immediate := z.Memory.ReadByte(z.PC + 1)

    z.C = immediate

    z.M = 8
    z.PC += 2
}

// 0x11 - LD DE, nn
func (z *Z80) LD_DE_nn() {
    // Lê os bytes imediatos seguintes ao PC para formar um valor de 16 bits (little-endian)
    lowByte := uint16(z.Memory.ReadByte(z.PC + 1))
    highByte := uint16(z.Memory.ReadByte(z.PC + 2))
    value := (highByte << 8) | lowByte

    z.DE = value

    z.M = 12
    z.PC += 3
}

// 0x12 - LD (DE), A
func (z *Z80) LD_DE_A() {
    // Obter o endereço apontado pelo par de registradores DE
    address := z.DE

    // Armazenar o valor do registrador A na memória no endereço DE
    z.Memory.WriteByte(address, z.A)

    z.M = 8
    z.PC++
}

//0x14
func (z *Z80) INC_D() {
    z.D++

    z.updateFlagsInc(z.D)

    z.M = 4
    z.PC++

}

//0x18
func (z *Z80) JR_e() {
    displacement := int8(z.readMemory(z.PC + 1))

    newAddress := uint16(int(z.PC) + 2 + int(displacement))

    z.PC = newAddress
    z.M = 12
}

//0x1C
func (z *Z80) INC_E() {
    z.E++

    z.updateFlagsInc(z.E)

    z.M = 4
    z.PC++

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
    z.PC += 3
    z.M = 12
}

// 0x22 - LD (HL+), A
func (z *Z80) LD_HL_inc_A() {
    address := z.HL

    z.Memory.WriteByte(address, z.A)

    z.HL++

    z.M = 8
    z.PC++
}

// 0x23 - INC HL
func (z *Z80) INC_HL() {
    z.HL++

    z.M = 8
    z.PC++
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

    z.HL++

    z.M = 8
    z.PC++
}

//0x30
func (z *Z80) JR_nc_e() {
    // Verificar se o flag de carry (C) está limpo (NC)
    if !z.CF {
        displacement := int8(z.readMemory(z.PC + 1))

        newAddress := uint16(int(z.PC) + 2 + int(displacement))

        z.PC = newAddress
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
    address := z.HL

    z.Memory.WriteByte(address, z.A)

    z.HL--

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

    z.updateFlagsInc(z.A)

    z.M = 4
    z.PC++

}

//0x3E
func (z *Z80) LD_A_d8() {
    immediate := z.readMemory(z.PC + 1)

    z.A = immediate

    z.PC += 2
    z.M = 8
}

// 0x40 - LD B, B
func (z *Z80) LD_B_B() {
    z.M = 4
    z.PC++
}

// 0x41 - LD B, C
func (z *Z80) LD_B_C() {
    z.B = z.C // Carrega o valor do registrador C no registrador B

    z.M = 4
    z.PC++
}

// 0x42 - LD B, D
func (z *Z80) LD_B_D() {
    z.B = z.D // Carrega o valor do registrador D no registrador B

    z.M = 4
    z.PC++
}

// 0x43 - LD B, E
func (z *Z80) LD_B_E() {
    z.B = z.E // Carrega o valor do registrador E no registrador B

    z.M = 4
    z.PC++
}

// 0x44 - LD B, H
func (z *Z80) LD_B_H() {
    z.B = z.H // Carrega o valor do registrador H no registrador B

    z.M = 4
    z.PC++
}

// 0x45 - LD B, L
func (z *Z80) LD_B_L() {
    z.B = z.L // Carrega o valor do registrador L no registrador B

    z.M = 4
    z.PC++
}

// 0x46 - LD B, HL
func (z *Z80) LD_B_HL() {
    address := z.HL
    z.B = z.Memory.ReadByte(address) // Carrega o valor do registrador HL no registrador B

    z.M = 4
    z.PC++
}

// 0x47 - LD B, A
func (z *Z80) LD_B_A() {
    z.B = z.A // Carrega o valor do registrador A no registrador B

    z.M = 4
    z.PC++
}

// 0x57 - LD D, A
func (z *Z80) LD_D_A() {
    z.D = z.A  

    z.PC++  
    z.M = 4 
}

// 0x78 - LD A, B
func (z *Z80) LD_A_B() {
    z.A = z.B  // Carrega o valor do registrador B no registrador A (Acumulador)

    z.PC++ 
    z.M = 4
}

// 0x79 - LD A, C
func (z *Z80) LD_A_C() {
    z.A = z.C  // Carrega o valor do registrador C no registrador A (Acumulador)

    z.PC++ 
    z.M = 4
}

// 0x7A - LD A, D
func (z *Z80) LD_A_D() {
    z.A = z.D  // Carrega o valor do registrador D no registrador A (Acumulador)

    z.PC++ 
    z.M = 4
}

// 0x7B - LD A, E
func (z *Z80) LD_A_E() {
    z.A = z.E  // Carrega o valor do registrador E no registrador A (Acumulador)

    z.PC++ 
    z.M = 4
}

// 0x7C - LD A, H
func (z *Z80) LD_A_H() {
    z.A = z.H  // Carrega o valor do registrador H no registrador A (Acumulador)

    z.PC++ 
    z.M = 4
}

// 0x7D - LD A, L
func (z *Z80) LD_A_L() {
    z.A = z.L  // Carrega o valor do registrador L no registrador A (Acumulador)

    z.PC++ 
    z.M = 4
}

// 0x7E - LD A, (HL)
func (z *Z80) LD_A_HL() {
    address := z.HL
    z.A = z.Memory.ReadByte(address)

    z.M = 8
    z.PC++
}

// 0x7F - LD A, A
func (z *Z80) LD_A_A() {
    z.PC++ 
    z.M = 4
}

// 0xA7 - AND A
func (z *Z80) AND_A() {
    z.A &= z.A

    z.Z = z.A == 0     
    z.N = false        
    z.HF = true        
    z.CF = false       

    z.PC++             
    z.M = 4           
}


//0xA8
func (z *Z80) XOR_B() {
    z.B = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

//0xA9
func (z *Z80) XOR_C() {
    z.C = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

//0xAA
func (z *Z80) XOR_D() {
    z.D = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

//0xAB
func (z *Z80) XOR_E() {
    z.E = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

//0xAC
func (z *Z80) XOR_H() {
    z.H = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

//0xAD
func (z *Z80) XOR_L() {
    z.L = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

//0xAE
func (z *Z80) XOR_HL() {
    z.setHL()
    address := z.HL
    value := z.Memory.ReadByte(address)

    z.A ^= value

    z.Z = z.A == 0 
    z.N = false   
    z.HF = false   
    z.CF = false   

    z.PC++ 
    z.M = 8
}


//0xAF
func (z *Z80) XOR_A() {
    z.A = 0

    // Definir flags de status
    z.Z = true 
    z.N = false 
    z.HF = false 
    z.CF = false

    z.PC++
    z.M = 4
}

// 0xB1 - OR C
func (z *Z80) OR_C() {
    z.A |= z.C

    z.Z = z.A == 0   
    z.N = false      
    z.HF = false     
    z.CF = false     

    z.PC++
    z.M = 4
}

// 0xC0 - RET NZ (Return if Not Zero)
func (z *Z80) RET_NZ() {
    if !z.Z { 
        returnAddress := z.Memory.ReadWord(z.SP) 
        z.SP += 2 

        z.PC = returnAddress
        z.M = 20
    } else {
        z.PC++ 
        z.M = 8 
    }
}

//0xC3
func (z *Z80) JP_nn() {
    lowByte := uint16(z.readMemory(z.PC + 1))
    highByte := uint16(z.readMemory(z.PC + 2))
    address := (highByte << 8) | lowByte

    z.PC = address

    z.M = 16
}

// 0xC5 - PUSH BC
func (z *Z80) PUSH_BC() {
    z.SP-- 
    z.Memory.WriteByte(z.SP, byte(z.BC>>8)) 
    z.SP-- 
    z.Memory.WriteByte(z.SP, byte(z.BC))

    z.PC++ 
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
    z.DE = z.Memory.ReadWord(z.SP)
    z.SP += 2

    z.M = 10
    z.PC++
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

//0xE2
func (z *Z80) LD_C_A() {
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

//0xF0
func (z *Z80) LDH_A_a8() {
    address := uint16(0xFF00) + uint16(z.readMemory(z.PC + 1))

    z.A = z.readMemory(address)

    z.PC += 2
    z.M = 12
}

//0xF3
func (z *Z80) DI() {
    z.IME = false
    z.PC++
    z.M = 4 
}

// 0xF5 - PUSH AF
func (z *Z80) PUSH_AF() {
    z.SP-- 
    z.Memory.WriteByte(z.SP, byte(z.AF>>8)) 
    z.SP-- 
    z.Memory.WriteByte(z.SP, byte(z.AF)) 

    z.PC++ 
    z.M = 16 

}

//0xFA
func (z *Z80) LD_A_nn() {
    lowByte := uint16(z.readMemory(z.PC + 1))
    highByte := uint16(z.readMemory(z.PC + 2))
    address := (highByte << 8) | lowByte

    z.A = z.readMemory(address)

    z.PC += 3
    z.M = 13
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
    case 0x11:
        z.LD_DE_nn()
    case 0x12:
        z.LD_DE_A()
    case 0x14:
        z.INC_D()
    case 0x18:
        z.JR_e()
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
    case 0x28:
        z.JR_Z_e()
    case 0x2A:
        z.LDI_A_HL()
    case 0x30:
        z.JR_nc_e()
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
        z.LD_B_HL()
    case 0x47:
        z.LD_B_A()
    case 0x57:
        z.LD_D_A()
    case 0x78:
        z.LD_A_B()
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
        z.XOR_HL()
    case 0xAF:
        z.XOR_A()
    case 0xB1:
        z.OR_C()
    case 0xC0:
        z.RET_NZ()
    case 0xC3:
        z.JP_nn()
    case 0xC5:
        z.PUSH_BC()
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
    case 0xE2:
        z.LD_C_A()
    case 0xE5:
        z.PUSH_HL()
    case 0xE6:
        z.AND_n()
    case 0xEA:
        z.LD_nn_A()
    case 0xF0:
        z.LDH_A_a8()
    case 0xF3:
        z.DI()
    case 0xF5:
        z.PUSH_AF()
    case 0xFA:
        z.LD_A_nn()
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
            z.M = 8

        case 0x86:
            // RES 0, (HL)
            address := z.HL
            value := z.Memory.ReadByte(address)
            value &^= (1 << 0) // Clear bit 0
            z.Memory.WriteByte(address, value)
            z.M = 16
        case 0x87:
            // RES 0, A
            z.A &^= (1 << 0) // Clear bit 0 of A
            z.Z = z.A == 0
            z.N = false
            z.HF = false
            z.CF = false
            z.M = 8
        case 0x88:
            // RES 0, B
            z.B &^= (1 << 0) // Clear bit 0 of B
            z.Z = z.B == 0
            z.N = false
            z.HF = false
            z.CF = false
            z.M = 8
        case 0x89:
            // RES 0, C
            z.C &^= (1 << 0) // Clear bit 0 of C
            z.Z = z.C == 0
            z.N = false
            z.HF = false
            z.CF = false
            z.M = 8
    

        // Outras instruções CB...

        default:
            fmt.Printf("Opcode CB não suportado: 0xCB%x\n", opcodeCB)
            fmt.Scanln() 
            z.M = 4
        }
}


func (z *Z80) EmulateCycle() int{
    opcode := z.readMemory(z.PC)
    //incrementar o PC?
    fmt.Printf("PC:  0x%X\n", z.PC)
    fmt.Printf("Hram 0x44:  0x%X\n", z.Memory.Hram[0x44])
    fmt.Printf("excutando opcode:  0x%X\n", opcode)
    // if opcode == 0xE2{
    //     fmt.Println("Pressione Enter para continuar...")
    //     fmt.Scanln() // Aguardar entrada do usuário (pressionar Enter)
    // }

    z.ExecuteInstruction(opcode)


    //z.M--
    //return z.M *4
    //fmt.Println(z.M)
    //fmt.Scanln()
    return z.M

}