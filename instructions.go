package _65go2

type OpCode uint8

type Instruction struct {
	opcode OpCode
	exec   func(*Cpu) (cycles uint16)
}

type InstructionTable map[OpCode]Instruction

func NewInstructionTable() InstructionTable {
	instructions := make(map[OpCode]Instruction)
	InstructionTable(instructions).InitInstructions()
	return instructions
}

func (instructions InstructionTable) AddInstruction(inst Instruction) {
	instructions[inst.opcode] = inst
}

func (instructions InstructionTable) RemoveInstruction(opcode OpCode) {
	delete(instructions, opcode)
}

func (instructions InstructionTable) InitInstructions() {
	// LDA

	//     Immediate
	instructions.AddInstruction(Instruction{
		opcode: 0xa9,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Lda(cpu.immediateAddress())
			return
		}})

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0xa5,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Lda(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(Instruction{
		opcode: 0xb5,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Lda(cpu.zeroPageIndexedAddress(cpu.registers.X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0xad,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Lda(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(Instruction{
		opcode: 0xbd,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Lda(cpu.absoluteIndexedAddress(cpu.registers.X, &cycles))
			return
		}})

	//     Absolute,Y
	instructions.AddInstruction(Instruction{
		opcode: 0xb9,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Lda(cpu.absoluteIndexedAddress(cpu.registers.Y, &cycles))
			return
		}})

	//     (Indirect,X)
	instructions.AddInstruction(Instruction{
		opcode: 0xa1,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 6
			cpu.Lda(cpu.indexedIndirectAddress())
			return
		}})

	//     (Indirect),Y
	instructions.AddInstruction(Instruction{
		opcode: 0xb1,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 5
			cpu.Lda(cpu.indirectIndexedAddress(&cycles))
			return
		}})

	// LDX

	//     Immediate
	instructions.AddInstruction(Instruction{
		opcode: 0xa2,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Ldx(cpu.immediateAddress())
			return
		}})

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0xa6,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Ldx(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,Y
	instructions.AddInstruction(Instruction{
		opcode: 0xb6,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Ldx(cpu.zeroPageIndexedAddress(cpu.registers.Y))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0xae,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Ldx(cpu.absoluteAddress())
			return
		}})

	//     Absolute,Y
	instructions.AddInstruction(Instruction{
		opcode: 0xbe,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Ldx(cpu.absoluteIndexedAddress(cpu.registers.Y, &cycles))
			return
		}})

	// LDY

	//     Immediate
	instructions.AddInstruction(Instruction{
		opcode: 0xa0,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Ldy(cpu.immediateAddress())
			return
		}})

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0xa4,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Ldy(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(Instruction{
		opcode: 0xb4,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Ldy(cpu.zeroPageIndexedAddress(cpu.registers.X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0xac,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Ldy(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(Instruction{
		opcode: 0xbc,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Ldy(cpu.absoluteIndexedAddress(cpu.registers.X, &cycles))
			return
		}})

	// STA

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0x85,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Sta(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(Instruction{
		opcode: 0x95,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Sta(cpu.zeroPageIndexedAddress(cpu.registers.X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0x8d,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Sta(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(Instruction{
		opcode: 0x9d,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 5
			cpu.Sta(cpu.absoluteIndexedAddress(cpu.registers.X, nil))
			return
		}})

	//     Absolute,Y
	instructions.AddInstruction(Instruction{
		opcode: 0x99,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 5
			cpu.Sta(cpu.absoluteIndexedAddress(cpu.registers.Y, nil))
			return
		}})

	//     (Indirect,X)
	instructions.AddInstruction(Instruction{
		opcode: 0x81,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 6
			cpu.Sta(cpu.indexedIndirectAddress())
			return
		}})

	//     (Indirect),Y
	instructions.AddInstruction(Instruction{
		opcode: 0x91,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 6
			cpu.Sta(cpu.indirectIndexedAddress(nil))
			return
		}})

	// STX

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0x86,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Stx(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,Y
	instructions.AddInstruction(Instruction{
		opcode: 0x96,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Stx(cpu.zeroPageIndexedAddress(cpu.registers.Y))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0x8e,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Stx(cpu.absoluteAddress())
			return
		}})

	// STY

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0x84,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Sty(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(Instruction{
		opcode: 0x94,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Sty(cpu.zeroPageIndexedAddress(cpu.registers.X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0x8c,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Sty(cpu.absoluteAddress())
			return
		}})

	// TAX

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0xaa,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Tax()
			return
		}})

	// TAY

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0xa8,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Tay()
			return
		}})

	// TXA

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x8a,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Txa()
			return
		}})

	// TYA

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x98,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Tya()
			return
		}})

	// TSX

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0xba,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Tsx()
			return
		}})

	// TXS

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x9a,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Txs()
			return
		}})

	// PHA

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x48,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Pha()
			return
		}})

	// PHP

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x08,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Php()
			return
		}})

	// PLA

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x68,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Pla()
			return
		}})

	// PLP

	//     Implied
	instructions.AddInstruction(Instruction{
		opcode: 0x28,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Plp()
			return
		}})

	// AND

	//     Immediate
	instructions.AddInstruction(Instruction{
		opcode: 0x29,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.And(cpu.immediateAddress())
			return
		}})

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0x25,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.And(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(Instruction{
		opcode: 0x35,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.And(cpu.zeroPageIndexedAddress(cpu.registers.X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0x2d,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.And(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(Instruction{
		opcode: 0x3d,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.And(cpu.absoluteIndexedAddress(cpu.registers.X, &cycles))
			return
		}})

	//     Absolute,Y
	instructions.AddInstruction(Instruction{
		opcode: 0x39,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.And(cpu.absoluteIndexedAddress(cpu.registers.Y, &cycles))
			return
		}})

	//     (Indirect,X)
	instructions.AddInstruction(Instruction{
		opcode: 0x21,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 6
			cpu.And(cpu.indexedIndirectAddress())
			return
		}})

	//     (Indirect),Y
	instructions.AddInstruction(Instruction{
		opcode: 0x31,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 5
			cpu.And(cpu.indirectIndexedAddress(&cycles))
			return
		}})

	// EOR

	//     Immediate
	instructions.AddInstruction(Instruction{
		opcode: 0x49,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 2
			cpu.Eor(cpu.immediateAddress())
			return
		}})

	//     Zero Page
	instructions.AddInstruction(Instruction{
		opcode: 0x45,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 3
			cpu.Eor(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(Instruction{
		opcode: 0x55,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Eor(cpu.zeroPageIndexedAddress(cpu.registers.X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(Instruction{
		opcode: 0x4d,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Eor(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(Instruction{
		opcode: 0x5d,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Eor(cpu.absoluteIndexedAddress(cpu.registers.X, &cycles))
			return
		}})

	//     Absolute,Y
	instructions.AddInstruction(Instruction{
		opcode: 0x59,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 4
			cpu.Eor(cpu.absoluteIndexedAddress(cpu.registers.Y, &cycles))
			return
		}})

	//     (Indirect,X)
	instructions.AddInstruction(Instruction{
		opcode: 0x41,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 6
			cpu.Eor(cpu.indexedIndirectAddress())
			return
		}})

	//     (Indirect),Y
	instructions.AddInstruction(Instruction{
		opcode: 0x51,
		exec: func(cpu *Cpu) (cycles uint16) {
			cycles = 5
			cpu.Eor(cpu.indirectIndexedAddress(&cycles))
			return
		}})

}
