package gb

import "fmt"

func (z *Z80) ExecuteCBInstruction() {
	opcodeCB := z.readMemory(z.PC + 1)
	//z.PC += 2

	switch opcodeCB {
	case 0x19:
		//RR C
		carry := (z.C & 0x01) != 0
		result := z.C >> 1

		if z.CF {
			result |= 0x80
		}

		z.C = result
		z.setBC()

		z.Z = z.C == 0
		z.N = false
		z.HF = false
		z.CF = carry
		z.setFlags()

		z.PC += 2
		z.M = 8
		fmt.Println("CB19")
	case 0x1A:
		//RR D
		carry := (z.D & 0x01) != 0
		result := z.D >> 1

		if z.CF {
			result |= 0x80
		}
		z.D = result
		z.setDE()

		z.Z = z.D == 0
		z.N = false
		z.HF = false
		z.CF = carry
		z.setFlags()

		z.PC += 2
		z.M = 8
		fmt.Println("CB1A")
	case 0x37:
		// SWAP A
		z.A = (z.A >> 4) | (z.A << 4)
		z.Z = z.A == 0
		z.N = false
		z.HF = false
		z.CF = false
		z.setFlags()

		z.PC += 2
		z.M = 8
	case 0x38:
		//SRL B
		carry := (z.B & 0x01) != 0
		result := z.B >> 1
		result &= 0x7F

		z.B = result
		z.setBC()

		z.Z = z.B == 0
		z.N = false
		z.HF = false
		z.CF = carry

		z.setFlags()

		z.PC += 2
		z.M = 8
		fmt.Println("CB38")
	case 0x3F:
		//SRL A
		carry := (z.A & 0x01) != 0
		result := z.A >> 1
		result &= 0x7F

		z.A = result

		z.Z = z.A == 0
		z.N = false
		z.HF = false
		z.CF = carry

		z.setFlags()

		z.PC += 2
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
		// z.Z = z.A == 0
		// z.N = false
		// z.HF = false
		// z.CF = false
		z.setFlags()
		z.PC += 2
		z.M = 8
		fmt.Println("CB87")
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
