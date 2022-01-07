// Copyright (c) 2021 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ga

import (
	"fmt"
)

const headerARM64 = header1 + header2

type RegARM64 uint8

const (
	X0 RegARM64 = iota
	X1
	X2
	X3
	X4
	X5
	X6
	X7
	X8
	X9
	X10
	X11
	X12
	X13
	X14
	X15
	X16
	X17
	X18
	X19
	X20
	X21
	X22
	X23
	X24
	X25
	X26
	X27
	X28
	X29
	X30
	X31

	XLR = X30
	XSP = X31
	XZR = X31
)

func (r RegARM64) String() string {
	return r.reg()
}

func (r RegARM64) reg() string {
	if r < 32 {
		return fmt.Sprintf("x%d", r)
	}

	panic(r)
}

func (r RegARM64) reg4() string {
	if r < 32 {
		return fmt.Sprintf("w%d", r)
	}

	panic(r)
}

var ARM64 = &ArchARM64{
	ClearableRegs: []RegARM64{
		X0,
		X1,
		X2,
		X3,
		X4,
		X5,
		X6,
		X7,
		X8,
		X9,
		X10,
		X11,
		X12,
		X13,
		X14,
		X15,
		X16,
		X17,
		X18,
		X19,
		X20,
		X21,
		X22,
		X23,
		X24,
		X25,
		X26,
		X27,
		X28,
		X29,
		// X30
		// X31
	},
}

type ArchARM64 struct {
	ClearableRegs []RegARM64
}

func (*ArchARM64) Machine() string {
	return "aarch64"
}

func (*ArchARM64) Specify(x Specific) int {
	return x.ARM64
}

func (arch *ArchARM64) ClearReg(a *Assembly, r RegARM64) {
	if a.Arch != arch {
		panic(a.Arch)
	}
	a.insn("mov", r.reg(), "0")
}

func (*ArchARM64) newAssembly(sys *System, buf *buffer) ArchAssembly {
	buf.WriteString(headerARM64)
	return &arm64{
		System: sys,
		buffer: buf,
	}
}

type arm64 struct {
	*System
	*buffer
}

func (a *arm64) check(r Reg) {
	a.checkUsage(uint8(r.ARM64), r.Use)
}

func (a *arm64) Set(r Reg) {
	a.buffer.regUsage[r.ARM64] = r.Use
}

func (a *arm64) Label(name string) {
	if global(name) {
		a.printf("")
		a.printf(".globl %s", symbol(name))
		a.printf(".type  %s,@function", symbol(name))
	}
	a.printf("")
	a.label(name)
}

func (a *arm64) FunctionEpilogue() {
	a.insnf("ldr lr, [%s], 8", a.reg(a.StackPtr))
}

func (a *arm64) Function(name string) {
	if global(name) {
		a.printf("")
		a.printf(".globl %s", symbol(name))
		a.printf(".type  %s,@function", symbol(name))
	}
	a.printf("")
	a.label(name)
	a.insnf("str lr, [%s, -8]!", a.reg(a.StackPtr))
}

func (a *arm64) FunctionWithoutPrologue(name string) {
	if global(name) {
		a.printf("")
		a.printf(".globl %s", symbol(name))
	}
	a.printf("")
	a.label(name)
}

func (a *arm64) Return() {
	a.FunctionEpilogue()
	a.ReturnWithoutEpilogue()
}

func (a *arm64) ReturnWithoutEpilogue() {
	a.insn("ret")
	a.speculationBarrier()
}

func (a *arm64) Address(dest Reg, name string) {
	a.insn("adr", a.reg(dest), symbol(name))
	a.Set(dest)
}

func (a *arm64) MoveDef(dest Reg, name string) {
	a.insn("mov", a.reg(dest), symbol(name))
	a.Set(dest)
}

func (a *arm64) MoveImm(dest Reg, value int) {
	a.MoveImm64(dest, uint64(int64(value)))
	a.Set(dest)
}

func (a *arm64) MoveImm64(dest Reg, value uint64) {
	defer a.Set(dest)

	// These cases are not supported by the loops below.
	if value == 0 || int64(value) == -1 {
		a.insn("mov", a.reg(dest), a.imm(int(int64(value))))
		return
	}

	var zbased int
	var nbased int
	for i := uint(0); i < 64; i += 16 {
		if uint16(value>>i) != 0 {
			zbased++
		}
		if uint16(value>>i) != 0xffff {
			nbased++
		}
	}

	if zbased <= nbased {
		op := "movz" // First instruction clears surroundings.
		for i := uint(0); i < 64; i += 16 {
			v := uint16(value >> i)
			if v != 0 {
				a.insn(op, a.reg(dest), a.imm(int(v)), fmt.Sprintf("lsl #%d", i))
				op = "movk" // Secondary instructions keep surroundings.
			}
		}
	} else {
		first := true
		for i := uint(0); i < 64; i += 16 {
			v := uint16(value >> i)
			if v != 0xffff {
				if first {
					a.insn("movn", a.reg(dest), a.imm(int(^v)), fmt.Sprintf("lsl #%d", i))
				} else {
					a.insn("movk", a.reg(dest), a.imm(int(v)), fmt.Sprintf("lsl #%d", i))
				}
				first = false
			}
		}
	}
}

func (a *arm64) MoveReg(dest, src Reg) {
	a.check(src)
	if a.reg(dest) != a.reg(src) {
		a.insn("mov", a.reg(dest), a.reg(src))
	}
	a.Set(dest)
}

func (a *arm64) MoveRegFloat(dest Reg, src FloatReg) {
	a.insn("fmov", a.reg(dest), a.floatreg(src))
	a.Set(dest)
}

func (a *arm64) AddImm(dest, src Reg, value int) {
	a.check(src)
	a.insn("add", a.reg(dest), a.reg(src), a.imm(value))
	a.Set(dest)
}

func (a *arm64) AddReg(dest, src1, src2 Reg) {
	a.check(src1)
	a.check(src2)
	a.insn("add", a.reg(dest), a.reg(src1), a.reg(src2))
	a.Set(dest)
}

func (a *arm64) SubtractImm(dest Reg, value int) {
	a.check(dest)
	if value != 0 {
		a.insn("sub", a.reg(dest), a.reg(dest), a.imm(value))
	}
	a.Set(dest)
}

func (a *arm64) SubtractReg(dest, src Reg) {
	a.check(dest)
	a.check(src)
	a.insn("sub", a.reg(dest), a.reg(dest), a.reg(src))
	a.Set(dest)
}

func (a *arm64) MultiplyImm(dest, src Reg, value int, temp Reg) {
	a.check(src)
	a.MoveImm(temp, value)
	a.insn("mul", a.reg(dest), a.reg(src), a.reg(temp))
	a.Set(dest)
	a.Set(temp.As(""))
}

func (a *arm64) AndImm(dest Reg, value int) {
	a.check(dest)
	a.insn("and", a.reg(dest), a.reg(dest), a.imm(value))
	a.Set(dest)
}

func (a *arm64) AndReg(dest, src Reg) {
	a.check(dest)
	a.check(src)
	a.insn("and", a.reg(dest), a.reg(dest), a.reg(src))
	a.Set(dest)
}

func (a *arm64) OrImm(dest Reg, value int) {
	a.check(dest)
	a.insn("orr", a.reg(dest), a.reg(dest), a.imm(value))
	a.Set(dest)
}

func (a *arm64) OrReg(dest, src Reg) {
	a.check(dest)
	a.check(src)
	a.insn("orr", a.reg(dest), a.reg(dest), a.reg(src))
	a.Set(dest)
}

func (a *arm64) ShiftImm(s Shift, r Reg, count int) {
	a.check(r)
	if count != 0 {
		a.insn(a.shift(s), a.reg(r), a.reg(r), a.imm(count))
	}
	a.Set(r)
}

func (a *arm64) Load(dest, base Reg, offset int) {
	a.check(base)
	a.insnf("ldr %s, [%s, %d]", a.reg(dest), a.reg(base), offset)
	a.Set(dest)
}

func (a *arm64) Load4Bytes(dest, base Reg, offset int) {
	a.check(base)
	a.insnf("ldr %s, [%s, %d]", a.reg4(dest), a.reg(base), offset)
	a.Set(dest)
}

func (a *arm64) LoadByte(dest, base Reg, offset int) {
	a.check(base)
	a.insnf("ldurb %s, [%s, %d]", a.reg4(dest), a.reg(base), offset)
	a.Set(dest)
}

func (a *arm64) Store(base Reg, offset int, src Reg) {
	a.check(base)
	a.check(src)
	a.insnf("str %s, [%s, %d]", a.reg(src), a.reg(base), offset)
}

func (a *arm64) Store4Bytes(base Reg, offset int, src Reg) {
	a.check(base)
	a.check(src)
	a.insnf("str %s, [%s, %d]", a.reg4(src), a.reg(base), offset)
}

func (a *arm64) Push(r Reg) {
	a.check(r)
	a.insnf("str %s, [%s, -8]!", a.reg(r), a.reg(a.StackPtr))
}

func (a *arm64) Pop(r Reg) {
	a.insnf("ldr %s, [%s], 8", a.reg(r), a.reg(a.StackPtr))
	a.Set(r)
}

func (a *arm64) Jump(name string) {
	a.insn("b", symbol(name))
}

func (a *arm64) JumpRegRoutine(r Reg, internalNamePrefix string) {
	a.check(r)
	a.insn("br", a.reg(r))
	a.speculationBarrier()
}

func (a *arm64) JumpIfBitSet(r Reg, bit uint, name string) {
	a.check(r)
	a.insn("tbnz", a.reg(r), a.imm(int(bit)), symbol(name))
}

func (a *arm64) JumpIfBitNotSet(r Reg, bit uint, name string) {
	a.check(r)
	a.insn("tbz", a.reg(r), a.imm(int(bit)), symbol(name))
}

func (a *arm64) JumpIfImm(c Cond, r Reg, value int, name string) {
	a.check(r)
	a.insn("cmp", a.reg(r), a.imm(value))
	a.insn("b."+a.cond(c), symbol(name))
}

func (a *arm64) JumpIfReg(c Cond, dest, src Reg, name string) {
	a.check(dest)
	a.check(src)
	a.insn("cmp", a.reg(dest), a.reg(src))
	a.insn("b."+a.cond(c), symbol(name))
}

func (a *arm64) Call(name string) {
	a.insn("bl", symbol(name))
}

func (a *arm64) Syscall(nr Syscall) {
	a.insn("mov", a.SyscallNr.ARM64.reg4(), a.imm(nr.ARM64))
	a.insn("svc", a.imm(0))
	a.Set(a.SysResult)
}

func (a *arm64) Unreachable() {
	a.insn("brk", a.imm(0))
}

func (a *arm64) speculationBarrier() {
	a.insn("dsb", "sy")
	a.insn("isb")
}

func (a *arm64) imm(x int) string {
	return fmt.Sprintf("%d", x)
}

func (a *arm64) reg(x Reg) string {
	return x.ARM64.reg()
}

func (a *arm64) reg4(x Reg) string {
	return x.ARM64.reg4()
}

func (a *arm64) floatreg(x FloatReg) string {
	return fmt.Sprintf("d%d", x)
}

func (a *arm64) cond(x Cond) string {
	switch x {
	case EQ:
		return "eq"
	case NE:
		return "ne"
	case LT:
		return "lt"
	case LE:
		return "le"
	case GT:
		return "gt"
	case GE:
		return "ge"
	}

	panic(x)
}

func (a *arm64) shift(x Shift) string {
	switch x {
	case Left:
		return "lsl"
	case RightLogical:
		return "lsr"
	case RightArithmetic:
		return "asr"
	}

	panic(x)
}
