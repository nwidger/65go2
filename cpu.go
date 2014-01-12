// Package m65go2 simulates the MOS 6502 CPU
package m65go2

import (
	"fmt"
	"os"
	"time"
)

// Flags used by P (Status) register
type Status uint8

const (
	C Status = 1 << iota // carry flag
	Z                    // zero flag
	I                    // interrupt disable
	D                    // decimal mode
	B                    // break command
	_                    // -UNUSED-
	V                    // overflow flag
	N                    // negative flag
)

// The 6502's registers, all registers are 8-bit values except for PC
// which is 16-bits.
type Registers struct {
	A  uint8  // accumulator
	X  uint8  // index register X
	Y  uint8  // index register Y
	P  Status // processor status
	SP uint8  // stack pointer
	PC uint16 // program counter
}

// Creates a new set of Registers.  All registers are initialized to
// 0.
func NewRegisters() Registers {
	return Registers{}
}

// Resets all registers.  Register P is initialized with only the I
// bit set, SP is initialized to 0xfd, PC is initialized to 0xfffc
// (the RESET vector) and all other registers are initialized to 0.
func (reg *Registers) Reset() {
	reg.A = 0
	reg.X = 0
	reg.Y = 0
	reg.P = I
	reg.SP = 0xfd
	reg.PC = 0xfffc
}

// Prints the values of each register to os.Stderr.
func (reg *Registers) String() {
	fmt.Fprintf(os.Stderr, "A:  %#02x (%03dd) (%08bb)\n", reg.A, reg.A, reg.A)
	fmt.Fprintf(os.Stderr, "X:  %#02x (%03dd) (%08bb)\n", reg.X, reg.X, reg.X)
	fmt.Fprintf(os.Stderr, "Y:  %#02x (%03dd) (%08bb)\n", reg.Y, reg.Y, reg.Y)
	fmt.Fprintf(os.Stderr, "SP: %#02x (%03dd) (%08bb)\n", reg.SP, reg.SP, reg.SP)

	f := ""

	getFlag := func(flag Status, set string) string {
		if reg.P&flag != 0 {
			return set
		}

		return "-"
	}

	f += getFlag(N, "N")
	f += getFlag(V, "V")
	f += "-" // -UNUSED-
	f += getFlag(B, "B")
	f += getFlag(D, "D")
	f += getFlag(I, "I")
	f += getFlag(Z, "Z")
	f += getFlag(C, "C")

	fmt.Fprintf(os.Stderr, "P:  %08bb (%s)\n", reg.P, f)
	fmt.Fprintf(os.Stderr, "PC: %#04x (%05dd) (%016bb)\n", reg.PC, reg.PC, reg.PC)
}

type Index uint8

const (
	X Index = iota
	Y
)

type CPUer interface {
	Reset()
	Execute() (cycles uint16, error error)
	Run() (error error)

	immediateAddress() (result uint16)
	zeroPageAddress() (result uint16)
	zeroPageIndexedAddress(index Index) (result uint16)
	relativeAddress() (result uint16)
	absoluteAddress() (result uint16)
	indirectAddress() (result uint16)
	absoluteIndexedAddress(index Index, cycles *uint16) (result uint16)
	indexedIndirectAddress() (result uint16)
	indirectIndexedAddress(cycles *uint16) (result uint16)

	Lda(address uint16)
	Ldx(address uint16)
	Ldy(address uint16)
	Sta(address uint16)
	Stx(address uint16)
	Sty(address uint16)
	Tax()
	Tay()
	Txa()
	Tya()
	Tsx()
	Txs()
	Pha()
	Php()
	Pla()
	Plp()
	And(address uint16)
	Eor(address uint16)
	Ora(address uint16)
	Bit(address uint16)
	Adc(address uint16)
	Sbc(address uint16)
	Cmp(address uint16)
	Cpx(address uint16)
	Cpy(address uint16)
	Inc(address uint16)
	Inx()
	Iny()
	Dec(address uint16)
	Dex()
	Dey()
	AslA()
	Asl(address uint16)
	LsrA()
	Lsr(address uint16)
	RolA()
	Rol(address uint16)
	RorA()
	Ror(address uint16)
	Jmp(address uint16)
	Jsr(address uint16)
	Rts()
	Bcc(address uint16, cycles *uint16)
	Bcs(address uint16, cycles *uint16)
	Beq(address uint16, cycles *uint16)
	Bmi(address uint16, cycles *uint16)
	Bne(address uint16, cycles *uint16)
	Bpl(address uint16, cycles *uint16)
	Bvc(address uint16, cycles *uint16)
	Bvs(address uint16, cycles *uint16)
	Clc()
	Cld()
	Cli()
	Clv()
	Sec()
	Sed()
	Sei()
	Brk()
	Rti()
}

// Represents the 6502 CPU.
type M6502 struct {
	decode       bool
	clock        Clocker
	Registers    Registers
	Memory       Memory
	Instructions InstructionTable
}

const DEFAULT_MASTER_RATE time.Duration = 46 * time.Nanosecond // 21.477272Mhz
const DEFAULT_CLOCK_DIVISOR uint64 = 12                        // 1.789733Mhz

// Returns a pointer to a new CPU with the given Memory and clock.
func NewM6502(mem Memory, clock Clocker) *M6502 {
	instructions := NewInstructionTable()
	instructions.InitInstructions()

	return &M6502{decode: false, clock: clock, Registers: NewRegisters(), Memory: mem, Instructions: instructions}
}

// Resets the CPU by resetting both the registers and memory.
func (cpu *M6502) Reset() {
	cpu.Registers.Reset()
	cpu.Memory.Reset()
}

// Error type used to indicate that the CPU attempted to execute an
// invalid opcode
type BadOpCodeError OpCode

func (b BadOpCodeError) Error() string {
	return fmt.Sprintf("No such opcode %#02x", b)
}

// Executes the instruction pointed to by the PC register in the
// number of clock cycles as returned by the instruction's Exec
// function.  Returns the number of cycles executed and any error
// (such as BadOpCodeError).
func (cpu *M6502) Execute() (cycles uint16, error error) {
	ticks := cpu.clock.Ticks()

	// fetch
	opcode := OpCode(cpu.Memory.Fetch(cpu.Registers.PC))
	inst, ok := cpu.Instructions[opcode]

	if !ok {
		return 0, BadOpCodeError(opcode)
	}

	// execute
	cpu.Registers.PC++
	cycles = inst.Exec(cpu)

	// count cycles
	cpu.clock.Await(ticks + uint64(cycles))

	return cycles, nil
}

// Executes instruction until Execute() returns an error.
func (cpu *M6502) Run() (error error) {
	for {
		if _, error := cpu.Execute(); error != nil {
			return error
		}
	}

	return nil
}

func (cpu *M6502) setZFlag(value uint8) uint8 {
	if value == 0 {
		cpu.Registers.P |= Z
	} else {
		cpu.Registers.P &= ^Z
	}

	return value
}

func (cpu *M6502) setNFlag(value uint8) uint8 {
	cpu.Registers.P = (cpu.Registers.P & ^N) | Status(value&uint8(N))
	return value
}

func (cpu *M6502) setZNFlags(value uint8) uint8 {
	cpu.setZFlag(value)
	cpu.setNFlag(value)
	return value
}

func (cpu *M6502) setCFlagAddition(value uint16) uint16 {
	cpu.Registers.P = (cpu.Registers.P & ^C) | Status(value>>8&uint16(C))
	return value
}

func (cpu *M6502) setVFlagAddition(term1 uint16, term2 uint16, result uint16) uint16 {
	cpu.Registers.P = (cpu.Registers.P & ^V) | Status((^(term1^term2)&(term1^result)&uint16(N))>>1)
	return result
}

func (cpu *M6502) load(address uint16, register *uint8) {
	*register = cpu.setZNFlags(cpu.Memory.Fetch(address))
}

func (cpu *M6502) immediateAddress() (result uint16) {
	result = cpu.Registers.PC
	cpu.Registers.PC++
	return
}

func (cpu *M6502) zeroPageAddress() (result uint16) {
	result = uint16(cpu.Memory.Fetch(cpu.Registers.PC))
	cpu.Registers.PC++
	return
}

func (cpu *M6502) IndexToRegister(which Index) uint8 {
	var index uint8

	switch which {
	case X:
		index = cpu.Registers.X
	case Y:
		index = cpu.Registers.Y
	}

	return index
}

func (cpu *M6502) zeroPageIndexedAddress(index Index) (result uint16) {
	result = uint16(cpu.Memory.Fetch(cpu.Registers.PC) + cpu.IndexToRegister(index))
	cpu.Registers.PC++
	return
}

func (cpu *M6502) relativeAddress() (result uint16) {
	value := uint16(cpu.Memory.Fetch(cpu.Registers.PC))
	cpu.Registers.PC++

	if value > 0x7f {
		result = cpu.Registers.PC - (0x0100 - value)
	} else {
		result = cpu.Registers.PC + value
	}

	return
}

func (cpu *M6502) absoluteAddress() (result uint16) {
	low := cpu.Memory.Fetch(cpu.Registers.PC)
	high := cpu.Memory.Fetch(cpu.Registers.PC + 1)
	cpu.Registers.PC += 2

	result = (uint16(high) << 8) | uint16(low)
	return
}

func (cpu *M6502) indirectAddress() (result uint16) {
	low := cpu.Memory.Fetch(cpu.Registers.PC)
	high := cpu.Memory.Fetch(cpu.Registers.PC + 1)
	cpu.Registers.PC += 2

	// XXX: The 6502 had a bug in which it incremented only the
	// high byte instead of the whole 16-bit address when
	// computing the address.
	//
	// See http://www.obelisk.demon.co.uk/6502/reference.html#JMP
	// and http://www.6502.org/tutorials/6502opcodes.html#JMP for
	// details
	aHigh := (uint16(high) << 8) | uint16(low+1)
	aLow := (uint16(high) << 8) | uint16(low)

	low = cpu.Memory.Fetch(aLow)
	high = cpu.Memory.Fetch(aHigh)

	result = (uint16(high) << 8) | uint16(low)
	return
}

func (cpu *M6502) absoluteIndexedAddress(index Index, cycles *uint16) (result uint16) {
	low := cpu.Memory.Fetch(cpu.Registers.PC)
	high := cpu.Memory.Fetch(cpu.Registers.PC + 1)
	cpu.Registers.PC += 2

	address := (uint16(high) << 8) | uint16(low)
	result = address + uint16(cpu.IndexToRegister(index))

	if cycles != nil && !SamePage(address, result) {
		*cycles++
	}

	return
}

func (cpu *M6502) indexedIndirectAddress() (result uint16) {
	address := uint16(cpu.Memory.Fetch(cpu.Registers.PC) + cpu.Registers.X)
	cpu.Registers.PC++

	low := cpu.Memory.Fetch(address)
	high := cpu.Memory.Fetch(address + 1)

	result = (uint16(high) << 8) | uint16(low)
	return
}

func (cpu *M6502) indirectIndexedAddress(cycles *uint16) (result uint16) {
	address := uint16(cpu.Memory.Fetch(cpu.Registers.PC))
	cpu.Registers.PC++

	low := cpu.Memory.Fetch(address)
	high := cpu.Memory.Fetch(address + 1)

	address = (uint16(high) << 8) | uint16(low)

	result = address + uint16(cpu.Registers.Y)

	if cycles != nil && !SamePage(address, result) {
		*cycles++
	}

	return
}

// Loads a byte of memory into the accumulator setting the zero and
// negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Lda(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: LDA $%04x\n", cpu.Registers.PC, address)
	}

	cpu.load(address, &cpu.Registers.A)
}

// Loads a byte of memory into the X register setting the zero and
// negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Ldx(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: LDX $%04x\n", cpu.Registers.PC, address)
	}

	cpu.load(address, &cpu.Registers.X)
}

// Loads a byte of memory into the Y register setting the zero and
// negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Ldy(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: LDY $%04x\n", cpu.Registers.PC, address)
	}

	cpu.load(address, &cpu.Registers.Y)
}

func (cpu *M6502) store(address uint16, value uint8) {
	cpu.Memory.Store(address, value)
}

// Stores the contents of the accumulator into memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sta(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: STA $%04x\n", cpu.Registers.PC, address)
	}

	cpu.store(address, cpu.Registers.A)
}

// Stores the contents of the X register into memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Stx(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: STX $%04x\n", cpu.Registers.PC, address)
	}

	cpu.store(address, cpu.Registers.X)
}

// Stores the contents of the Y register into memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sty(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: STY $%04x\n", cpu.Registers.PC, address)
	}

	cpu.store(address, cpu.Registers.Y)
}

func (cpu *M6502) transfer(from uint8, to *uint8) {
	*to = cpu.setZNFlags(from)
}

// Copies the current contents of the accumulator into the X register
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Tax() {
	if cpu.decode {
		fmt.Printf("  %04x: TAX\n", cpu.Registers.PC)
	}

	cpu.transfer(cpu.Registers.A, &cpu.Registers.X)
}

// Copies the current contents of the accumulator into the Y register
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Tay() {
	if cpu.decode {
		fmt.Printf("  %04x: TAY\n", cpu.Registers.PC)
	}

	cpu.transfer(cpu.Registers.A, &cpu.Registers.Y)
}

// Copies the current contents of the X register into the accumulator
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Txa() {
	if cpu.decode {
		fmt.Printf("  %04x: TXA\n", cpu.Registers.PC)
	}

	cpu.transfer(cpu.Registers.X, &cpu.Registers.A)
}

// Copies the current contents of the Y register into the accumulator
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Tya() {
	if cpu.decode {
		fmt.Printf("  %04x: TYA\n", cpu.Registers.PC)
	}

	cpu.transfer(cpu.Registers.Y, &cpu.Registers.A)
}

// Copies the current contents of the stack register into the X
// register and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Tsx() {
	if cpu.decode {
		fmt.Printf("  %04x: TSX\n", cpu.Registers.PC)
	}

	cpu.transfer(cpu.Registers.SP, &cpu.Registers.X)
}

// Copies the current contents of the X register into the stack
// register.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Txs() {
	if cpu.decode {
		fmt.Printf("  %04x: TXS\n", cpu.Registers.PC)
	}

	cpu.transfer(cpu.Registers.X, &cpu.Registers.SP)
}

func (cpu *M6502) push(value uint8) {
	cpu.Memory.Store(0x0100|uint16(cpu.Registers.SP), value)
	cpu.Registers.SP--
}

func (cpu *M6502) push16(value uint16) {
	cpu.push(uint8(value >> 8))
	cpu.push(uint8(value))
}

func (cpu *M6502) pull() (value uint8) {
	cpu.Registers.SP++
	value = cpu.Memory.Fetch(0x0100 | uint16(cpu.Registers.SP))
	return
}

func (cpu *M6502) pull16() (value uint16) {
	low := cpu.pull()
	high := cpu.pull()

	value = (uint16(high) << 8) | uint16(low)
	return
}

// Pushes a copy of the accumulator on to the stack.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Pha() {
	if cpu.decode {
		fmt.Printf("  %04x: PHA\n", cpu.Registers.PC)
	}

	cpu.push(cpu.Registers.A)
}

// Pushes a copy of the status flags on to the stack.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Php() {
	if cpu.decode {
		fmt.Printf("  %04x: PHP\n", cpu.Registers.PC)
	}

	cpu.push(uint8(cpu.Registers.P | B))
}

// Pulls an 8 bit value from the stack and into the accumulator. The
// zero and negative flags are set as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Pla() {
	if cpu.decode {
		fmt.Printf("  %04x: PLA\n", cpu.Registers.PC)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.pull())
}

// Pulls an 8 bit value from the stack and into the processor
// flags. The flags will take on new states as determined by the value
// pulled.
//
//         C 	Carry Flag 	  Set from stack
//         Z 	Zero Flag 	  Set from stack
//         I 	Interrupt Disable Set from stack
//         D 	Decimal Mode Flag Set from stack
//         B 	Break Command 	  Set from stack
//         V 	Overflow Flag 	  Set from stack
//         N 	Negative Flag 	  Set from stack
func (cpu *M6502) Plp() {
	if cpu.decode {
		fmt.Printf("  %04x: PLP\n", cpu.Registers.PC)
	}

	cpu.Registers.P = Status(cpu.pull())
}

// A logical AND is performed, bit by bit, on the accumulator contents
// using the contents of a byte of memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) And(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: AND $%04x\n", cpu.Registers.PC, address)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.Registers.A & cpu.Memory.Fetch(address))
}

// An exclusive OR is performed, bit by bit, on the accumulator
// contents using the contents of a byte of memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Eor(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: EOR $%04x\n", cpu.Registers.PC, address)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.Registers.A ^ cpu.Memory.Fetch(address))
}

// An inclusive OR is performed, bit by bit, on the accumulator
// contents using the contents of a byte of memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Ora(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: ORA $%04x\n", cpu.Registers.PC, address)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.Registers.A | cpu.Memory.Fetch(address))
}

// This instructions is used to test if one or more bits are set in a
// target memory location. The mask pattern in A is ANDed with the
// value in memory to set or clear the zero flag, but the result is
// not kept. Bits 7 and 6 of the value from memory are copied into the
// N and V flags.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if the result if the AND is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set to bit 6 of the memory value
//         N 	Negative Flag 	  Set to bit 7 of the memory value
func (cpu *M6502) Bit(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BIT $%04x\n", cpu.Registers.PC, address)
	}

	value := cpu.Memory.Fetch(address)
	cpu.setZFlag(value & cpu.Registers.A)
	cpu.Registers.P = Status(uint8(cpu.Registers.P) | (value & 0xc0))
}

func (cpu *M6502) addition(value uint16) {
	orig := uint16(cpu.Registers.A)

	if cpu.Registers.P&D == 0 {
		result := cpu.setCFlagAddition(orig + value + uint16(cpu.Registers.P&C))
		cpu.Registers.A = cpu.setZNFlags(uint8(cpu.setVFlagAddition(orig, value, result)))
	} else {
		low := uint16(orig&0x000f) + uint16(value&0x000f) + uint16(cpu.Registers.P&C)
		high := uint16(orig&0x00f0) + uint16(value&0x00f0)

		if low >= 0x000a {
			low -= 0x000a
			high += 0x0010
		}

		if high >= 0x00a0 {
			high -= 0x00a0
		}

		result := cpu.setCFlagAddition(high | (low & 0x000f))
		cpu.Registers.A = cpu.setZNFlags(uint8(cpu.setVFlagAddition(orig, value, result)))
	}
}

// This instruction adds the contents of a memory location to the
// accumulator together with the carry bit. If overflow occurs the
// carry bit is set, this enables multiple byte addition to be
// performed.
//
//         C 	Carry Flag 	  Set if overflow in bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set if sign bit is incorrect
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Adc(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: ADC $%04x\n", cpu.Registers.PC, address)
	}

	value := uint16(cpu.Memory.Fetch(address))
	cpu.addition(value)
}

// This instruction subtracts the contents of a memory location to the
// accumulator together with the not of the carry bit. If overflow
// occurs the carry bit is clear, this enables multiple byte
// subtraction to be performed.
//
//         C 	Carry Flag 	  Clear if overflow in bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set if sign bit is incorrect
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Sbc(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: SBC $%04x\n", cpu.Registers.PC, address)
	}

	value := uint16(cpu.Memory.Fetch(address))

	if cpu.Registers.P&D == 0 {
		value ^= 0xff
	} else {
		value = 0x99 - value
	}

	cpu.addition(value)
}

func (cpu *M6502) compare(address uint16, register uint8) {
	value := uint16(cpu.Memory.Fetch(address)) ^ 0xff + 1
	cpu.setZNFlags(uint8(cpu.setCFlagAddition(uint16(register) + value)))
}

// This instruction compares the contents of the accumulator with
// another memory held value and sets the zero and carry flags as
// appropriate.
//
//         C 	Carry Flag 	  Set if A >= M
//         Z 	Zero Flag 	  Set if A = M
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Cmp(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: CMP $%04x\n", cpu.Registers.PC, address)
	}

	cpu.compare(address, cpu.Registers.A)
}

// This instruction compares the contents of the X register with
// another memory held value and sets the zero and carry flags as
// appropriate.
//
//         C 	Carry Flag 	  Set if X >= M
//         Z 	Zero Flag 	  Set if X = M
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Cpx(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: CPX $%04x\n", cpu.Registers.PC, address)
	}

	cpu.compare(address, cpu.Registers.X)
}

// This instruction compares the contents of the Y register with
// another memory held value and sets the zero and carry flags as
// appropriate.
//
//         C 	Carry Flag 	  Set if Y >= M
//         Z 	Zero Flag 	  Set if Y = M
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Cpy(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: CPY $%04x\n", cpu.Registers.PC, address)
	}

	cpu.compare(address, cpu.Registers.Y)
}

// Adds one to the value held at a specified memory location setting
// the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if result is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Inc(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: INC $%04x\n", cpu.Registers.PC, address)
	}

	cpu.Memory.Store(address, cpu.setZNFlags(cpu.Memory.Fetch(address)+1))
}

func (cpu *M6502) increment(register *uint8) {
	*register = cpu.setZNFlags(*register + 1)
}

// Adds one to the X register setting the zero and negative flags as
// appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Inx() {
	if cpu.decode {
		fmt.Printf("  %04x: INX\n", cpu.Registers.PC)
	}

	cpu.increment(&cpu.Registers.X)
}

// Adds one to the Y register setting the zero and negative flags as
// appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Iny() {
	if cpu.decode {
		fmt.Printf("  %04x: INY\n", cpu.Registers.PC)
	}

	cpu.increment(&cpu.Registers.Y)
}

// Subtracts one from the value held at a specified memory location
// setting the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if result is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Dec(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: DEC $%04x\n", cpu.Registers.PC, address)
	}

	cpu.Memory.Store(address, cpu.setZNFlags(cpu.Memory.Fetch(address)-1))
}

func (cpu *M6502) decrement(register *uint8) {
	*register = cpu.setZNFlags(*register - 1)
}

// Subtracts one from the X register setting the zero and negative
// flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Dex() {
	if cpu.decode {
		fmt.Printf("  %04x: DEX\n", cpu.Registers.PC)
	}

	cpu.decrement(&cpu.Registers.X)
}

// Subtracts one from the Y register setting the zero and negative
// flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Dey() {
	if cpu.decode {
		fmt.Printf("  %04x: DEY\n", cpu.Registers.PC)
	}

	cpu.decrement(&cpu.Registers.Y)
}

type direction int

const (
	left direction = iota
	right
)

func (cpu *M6502) shift(direction direction, value uint8, store func(uint8)) {
	c := Status(0)

	switch direction {
	case left:
		c = Status((value & uint8(N)) >> 7)
		value <<= 1
	case right:
		c = Status(value & uint8(C))
		value >>= 1
	}

	cpu.Registers.P &= ^C
	cpu.Registers.P |= c

	store(cpu.setZNFlags(value))
}

// This operation shifts all the bits of the accumulator one bit
// left. Bit 0 is set to 0 and bit 7 is placed in the carry flag. The
// effect of this operation is to multiply the memory contents by 2
// (ignoring 2's complement considerations), setting the carry if the
// result will not fit in 8 bits.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) AslA() {
	if cpu.decode {
		fmt.Printf("  %04x: ASL A\n", cpu.Registers.PC)
	}

	cpu.shift(left, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })
}

// This operation shifts all the bits of the memory contents one bit
// left. Bit 0 is set to 0 and bit 7 is placed in the carry flag. The
// effect of this operation is to multiply the memory contents by 2
// (ignoring 2's complement considerations), setting the carry if the
// result will not fit in 8 bits.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Asl(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: ASL $%04x\n", cpu.Registers.PC, address)
	}

	cpu.shift(left, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

// Each of the bits in A is shift one place to the right. The bit that
// was in bit 0 is shifted into the carry flag. Bit 7 is set to zero.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if result = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) LsrA() {
	if cpu.decode {
		fmt.Printf("  %04x: LSR A\n", cpu.Registers.PC)
	}

	cpu.shift(right, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })
}

// Each of the bits in M is shift one place to the right. The bit that
// was in bit 0 is shifted into the carry flag. Bit 7 is set to zero.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if result = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Lsr(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: LSR $%04x\n", cpu.Registers.PC, address)
	}

	cpu.shift(right, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

func (cpu *M6502) rotate(direction direction, value uint8, store func(uint8)) {
	c := Status(0)

	switch direction {
	case left:
		c = Status(value & uint8(N) >> 7)
		value = ((value << 1) & uint8(^C)) | uint8(cpu.Registers.P&C)
	case right:
		c = Status(value & uint8(C))
		value = ((value >> 1) & uint8(^N)) | uint8((cpu.Registers.P&C)<<7)
	}

	cpu.Registers.P &= ^C
	cpu.Registers.P |= c

	store(cpu.setZNFlags(value))
}

// Move each of the bits in A one place to the left. Bit 0 is filled
// with the current value of the carry flag whilst the old bit 7
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) RolA() {
	if cpu.decode {
		fmt.Printf("  %04x: ROL A\n", cpu.Registers.PC)
	}

	cpu.rotate(left, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })
}

// Move each of the bits in A one place to the left. Bit 0 is filled
// with the current value of the carry flag whilst the old bit 7
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Rol(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: ROL $%04x\n", cpu.Registers.PC, address)
	}

	cpu.rotate(left, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

// Move each of the bits in A one place to the right. Bit 7 is filled
// with the current value of the carry flag whilst the old bit 0
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) RorA() {
	if cpu.decode {
		fmt.Printf("  %04x: ROR A\n", cpu.Registers.PC)
	}

	cpu.rotate(right, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })
}

// Move each of the bits in M one place to the right. Bit 7 is filled
// with the current value of the carry flag whilst the old bit 0
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Ror(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: ROR $%04x\n", cpu.Registers.PC, address)
	}

	cpu.rotate(right, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

// Sets the program counter to the address specified by the operand.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Jmp(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: JMP $%04x\n", cpu.Registers.PC, address)
	}

	cpu.Registers.PC = address
}

// The JSR instruction pushes the address (minus one) of the return
// point on to the stack and then sets the program counter to the
// target memory address.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Jsr(address uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: JSR $%04x\n", cpu.Registers.PC, address)
	}

	value := cpu.Registers.PC - 1

	cpu.push16(value)

	cpu.Registers.PC = address
}

// The RTS instruction is used at the end of a subroutine to return to
// the calling routine. It pulls the program counter (minus one) from
// the stack.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Rts() {
	if cpu.decode {
		fmt.Printf("  %04x: RTS\n", cpu.Registers.PC)
	}

	cpu.Registers.PC = cpu.pull16() + 1
}

func (cpu *M6502) branch(address uint16, condition func() bool, cycles *uint16) {
	if condition() {
		*cycles++

		if !SamePage(cpu.Registers.PC, address) {
			*cycles++
		}

		cpu.Registers.PC = address
	}
}

// If the carry flag is clear then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bcc(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BCC $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&C == 0 }, cycles)
}

// If the carry flag is set then add the relative displacement to the
// program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bcs(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BCS $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&C != 0 }, cycles)
}

// If the zero flag is set then add the relative displacement to the
// program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Beq(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BEQ $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&Z != 0 }, cycles)
}

// If the negative flag is set then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bmi(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BMI $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&N != 0 }, cycles)
}

// If the zero flag is clear then add the relative displacement to the
// program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bne(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BNE $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&Z == 0 }, cycles)
}

// If the negative flag is clear then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bpl(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BPL $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&N == 0 }, cycles)
}

// If the overflow flag is clear then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bvc(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BVC $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&V == 0 }, cycles)
}

// If the overflow flag is set then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bvs(address uint16, cycles *uint16) {
	if cpu.decode {
		fmt.Printf("  %04x: BVS $%04x\n", cpu.Registers.PC, address)
	}

	cpu.branch(address, func() bool { return cpu.Registers.P&V != 0 }, cycles)
}

// Set the carry flag to zero.
//
//         C 	Carry Flag 	  Set to 0
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Clc() {
	if cpu.decode {
		fmt.Printf("  %04x: CLC\n", cpu.Registers.PC)
	}

	cpu.Registers.P &^= C
}

// Set the decimal mode flag to zero.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Set to 0
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Cld() {
	if cpu.decode {
		fmt.Printf("  %04x: CLD\n", cpu.Registers.PC)
	}

	cpu.Registers.P &^= D
}

// Clears the interrupt disable flag allowing normal interrupt
// requests to be serviced.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Set to 0
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Cli() {
	if cpu.decode {
		fmt.Printf("  %04x: CLI\n", cpu.Registers.PC)
	}

	cpu.Registers.P &^= I
}

// Clears the interrupt disable flag allowing normal interrupt
// requests to be serviced.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set to 0
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Clv() {
	if cpu.decode {
		fmt.Printf("  %04x: CLV\n", cpu.Registers.PC)
	}

	cpu.Registers.P &^= V
}

// Set the carry flag to one.
//
//         C 	Carry Flag 	  Set to 1
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sec() {
	if cpu.decode {
		fmt.Printf("  %04x: SEC\n", cpu.Registers.PC)
	}

	cpu.Registers.P |= C
}

// Set the decimal mode flag to one.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Set to 1
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sed() {
	if cpu.decode {
		fmt.Printf("  %04x: SED\n", cpu.Registers.PC)
	}

	cpu.Registers.P |= D
}

// Set the interrupt disable flag to one.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Set to 1
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sei() {
	if cpu.decode {
		fmt.Printf("  %04x: SEI\n", cpu.Registers.PC)
	}

	cpu.Registers.P |= I
}

// The BRK instruction forces the generation of an interrupt
// request. The program counter and processor status are pushed on the
// stack then the IRQ interrupt vector at $FFFE/F is loaded into the
// PC and the break flag in the status set to one.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Set to 1
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Brk() {
	if cpu.decode {
		fmt.Printf("  %04x: BRK\n", cpu.Registers.PC)
	}

	cpu.Registers.PC++

	cpu.push16(cpu.Registers.PC)
	cpu.push(uint8(cpu.Registers.P | B))

	cpu.Registers.P |= I

	low := cpu.Memory.Fetch(0xfffe)
	high := cpu.Memory.Fetch(0xffff)

	cpu.Registers.PC = (uint16(high) << 8) | uint16(low)
}

// The RTI instruction is used at the end of an interrupt processing
// routine. It pulls the processor flags from the stack followed by
// the program counter.
//
//         C 	Carry Flag 	  Set from stack
//         Z 	Zero Flag 	  Set from stack
//         I 	Interrupt Disable Set from stack
//         D 	Decimal Mode Flag Set from stack
//         B 	Break Command 	  Set from stack
//         V 	Overflow Flag 	  Set from stack
//         N 	Negative Flag 	  Set from stack
func (cpu *M6502) Rti() {
	if cpu.decode {
		fmt.Printf("  %04x: RTI\n", cpu.Registers.PC)
	}

	cpu.Registers.P = Status(cpu.pull())
	cpu.Registers.PC = cpu.pull16()
}
