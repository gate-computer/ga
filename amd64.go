// Copyright (c) 2021 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ga

import (
	"fmt"
)

const headerAMD64 = header1 + `
.intel_syntax noprefix
` + header2

type RegAMD64 uint8

const (
	RAX RegAMD64 = iota
	RCX
	RDX
	RBX
	RSP
	RBP
	RSI
	RDI
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
)

func (r RegAMD64) String() string {
	return r.reg()
}

func (r RegAMD64) reg() string {
	switch r {
	case RAX:
		return "rax"
	case RCX:
		return "rcx"
	case RDX:
		return "rdx"
	case RBX:
		return "rbx"
	case RSP:
		return "rsp"
	case RBP:
		return "rbp"
	case RSI:
		return "rsi"
	case RDI:
		return "rdi"
	}

	if r < 16 {
		return fmt.Sprintf("r%d", r)
	}

	panic(r)
}

func (r RegAMD64) reg4() string {
	switch r {
	case RAX:
		return "eax"
	case RCX:
		return "ecx"
	case RDX:
		return "edx"
	case RBX:
		return "ebx"
	case RSP:
		panic(r)
	case RBP:
		return "ebp"
	case RSI:
		return "esi"
	case RDI:
		return "edi"
	}

	if r < 16 {
		return fmt.Sprintf("r%dd", r)
	}

	panic(r)
}

func (r RegAMD64) reg1() string {
	switch r {
	case RAX:
		return "al"
	case RCX:
		return "cl"
	case RDX:
		return "dl"
	case RBX:
		return "bl"
	case RSP:
		panic(r)
	case RBP:
		return "bpl"
	case RSI:
		return "sil"
	case RDI:
		return "dil"
	}

	if r < 16 {
		return fmt.Sprintf("r%ddb", r)
	}

	panic(r)
}

var AMD64 = &ArchAMD64{
	ClearableRegs: []RegAMD64{
		RAX,
		RCX,
		RDX,
		RBX,
		// RSP
		RBP,
		RSI,
		RDI,
		R8,
		R9,
		R10,
		R11,
		R12,
		R13,
		R14,
		R15,
	},
}

type ArchAMD64 struct {
	ClearableRegs []RegAMD64
}

func (*ArchAMD64) Machine() string {
	return "x86_64"
}

func (*ArchAMD64) Specify(x Specific) int {
	return x.AMD64
}

func (arch *ArchAMD64) ClearReg(a *Assembly, r RegAMD64) {
	if a.Arch != arch {
		panic(a.Arch)
	}
	a.insn("xor", r.reg4(), r.reg4())
}

func (arch *ArchAMD64) OrMem4BytesImm(a *Assembly, base RegAMD64, offset, value int) {
	if a.Arch != arch {
		panic(a.Arch)
	}
	switch {
	case offset == 0:
		a.insnf("or dword ptr [%s], %d", base.reg(), value)
	case offset > 0:
		a.insnf("or dword ptr [%s + %d], %d", base.reg(), offset, value)
	default:
		a.insnf("or dword ptr [%s - %d], %d", base.reg(), -offset, value)
	}
}

func (arch *ArchAMD64) ExchangeMem4BytesReg(a *Assembly, base RegAMD64, offset int, r RegAMD64) {
	if a.Arch != arch {
		panic(a.Arch)
	}
	switch {
	case offset == 0:
		a.insnf("xchg [%s], %s", base.reg(), r.reg4())
	case offset > 0:
		a.insnf("xchg [%s + %d], %s", base.reg(), offset, r.reg4())
	default:
		a.insnf("xchg [%s - %d], %s", base.reg(), -offset, r.reg4())
	}
}

func (*ArchAMD64) newAssembly(sys *System, buf *buffer) ArchAssembly {
	buf.WriteString(headerAMD64)
	return &amd64{
		System: sys,
		buffer: buf,
	}
}

type amd64 struct {
	*System
	*buffer
}

func (a *amd64) check(r Reg) {
	a.checkUsage(uint8(r.AMD64), r.Use)
}

func (a *amd64) Set(r Reg) {
	a.buffer.regUsage[r.AMD64] = r.Use
}

func (a *amd64) Label(name string) {
	if global(name) {
		a.printf("")
		a.printf(".align 16,0x90") // nop
		a.printf(".globl %s", symbol(name))
		a.printf(".type  %s,@function", symbol(name))
	}
	a.printf("")
	a.label(name)
}

func (a *amd64) FunctionEpilogue() {
}

func (a *amd64) Function(name string) {
	a.printf("")
	a.printf(".align 16,0xcc") // int3
	if global(name) {
		a.printf(".globl %s", symbol(name))
		a.printf(".type  %s,@function", symbol(name))
	}
	a.printf("")
	a.label(name)
}

func (a *amd64) FunctionWithoutPrologue(name string) {
	a.printf("")
	a.printf(".align 16,0xcc") // int3
	if global(name) {
		a.printf(".globl %s", symbol(name))
	}
	a.printf("")
	a.label(name)
}

func (a *amd64) Return() {
	a.ReturnWithoutEpilogue()
}

func (a *amd64) ReturnWithoutEpilogue() {
	a.insn("ret")
	a.speculationBarrier()
}

func (a *amd64) Address(dest Reg, name string) {
	a.insnf("lea %s, [rip + %s]", a.reg(dest), symbol(name))
	a.Set(dest)
}

func (a *amd64) MoveDef(dest Reg, name string) {
	a.insnf("mov %s, %s", a.reg(dest), symbol(name))
	a.Set(dest)
}

func (a *amd64) MoveImm(dest Reg, value int) {
	switch {
	case value == 0:
		a.insn("xor", a.reg4(dest), a.reg4(dest))
	case value >= -0x80000000 && value <= 0x7fffffff:
		a.insn("mov", a.reg4(dest), a.imm(value))
	default:
		a.insn("mov", a.reg(dest), a.imm(value))
	}
	a.Set(dest)
}

func (a *amd64) MoveImm64(dest Reg, value uint64) {
	switch {
	case value == 0:
		a.insn("xor", a.reg4(dest), a.reg4(dest))
	case int64(value) >= -0x80000000 && int64(value) <= 0x7fffffff:
		a.insn("mov", a.reg4(dest), a.imm(int(value)))
	default:
		a.insn("mov", a.reg(dest), a.imm64(value))
	}
	a.Set(dest)
}

func (a *amd64) MoveReg(dest, src Reg) {
	a.check(src)
	if !dest.Is(src) {
		a.insn("mov", a.reg(dest), a.reg(src))
	}
	a.Set(dest)
}

func (a *amd64) MoveRegFloat(dest Reg, src FloatReg) {
	a.insn("movq", a.reg(dest), a.floatreg(src))
	a.Set(dest)
}

func (a *amd64) AddImm(dest, src Reg, value int) {
	a.check(src)
	switch {
	case value == 0:
		a.MoveReg(dest, src)
	case dest.Is(src):
		a.insn("add", a.reg(dest), a.imm(value))
	case value > 0:
		a.insnf("lea %s, [%s + %s]", a.reg(dest), a.reg(src), a.imm(value))
	default:
		a.insn("mov", a.reg(dest), a.reg(src))
		a.insn("add", a.reg(dest), a.imm(value))
	}
	a.Set(dest)
}

func (a *amd64) AddReg(dest, src1, src2 Reg) {
	a.check(src1)
	a.check(src2)
	switch {
	case dest.Is(src1):
		a.insn("add", a.reg(dest), a.reg(src2))
	case dest.Is(src2):
		a.insn("add", a.reg(dest), a.reg(src1))
	default:
		a.insnf("lea %s, [%s + %s]", a.reg(dest), a.reg(src1), a.reg(src2))
	}
	a.Set(dest)
}

func (a *amd64) SubtractImm(dest Reg, value int) {
	a.check(dest)
	if value != 0 {
		a.insn("sub", a.reg(dest), a.imm(value))
	}
	a.Set(dest)
}

func (a *amd64) SubtractReg(dest, src Reg) {
	a.check(dest)
	a.check(src)
	a.insn("sub", a.reg(dest), a.reg(src))
	a.Set(dest)
}

func (a *amd64) MultiplyImm(dest, src Reg, value int, temp Reg) {
	a.check(dest)
	a.check(src)
	a.Set(temp)
	a.insn("imul", a.reg(dest), a.reg(src), a.imm(value))
	a.Set(dest)
}

func (a *amd64) AndImm(dest Reg, value int) {
	a.check(dest)
	switch {
	case value == 0:
		a.insn("xor", a.reg4(dest), a.reg4(dest))
	case value > 0 && value <= 0x7fffffff:
		a.insn("and", a.reg4(dest), a.imm(value))
	default:
		a.insn("and", a.reg(dest), a.imm(value))
	}
	a.Set(dest)
}

func (a *amd64) AndReg(dest, src Reg) {
	a.check(dest)
	a.check(src)
	a.insn("and", a.reg(dest), a.reg(src))
	a.Set(dest)
}

func (a *amd64) OrImm(dest Reg, value int) {
	a.check(dest)
	a.insn("or", a.reg(dest), a.imm(value))
	a.Set(dest)
}

func (a *amd64) OrReg(dest, src Reg) {
	a.check(dest)
	a.check(src)
	a.insn("or", a.reg(dest), a.reg(src))
	a.Set(dest)
}

func (a *amd64) ShiftImm(s Shift, r Reg, count int) {
	a.check(r)
	if count != 0 {
		a.insn(a.shift(s), a.reg(r), a.imm(count))
	}
	a.Set(r)
}

func (a *amd64) Load(dest, base Reg, offset int) {
	a.check(base)
	switch {
	case offset == 0:
		a.insnf("mov %s, [%s]", a.reg(dest), a.reg(base))
	case offset > 0:
		a.insnf("mov %s, [%s + %d]", a.reg(dest), a.reg(base), offset)
	default:
		a.insnf("mov %s, [%s - %d]", a.reg(dest), a.reg(base), -offset)
	}
	a.Set(dest)
}

func (a *amd64) Load4Bytes(dest, base Reg, offset int) {
	a.check(base)
	switch {
	case offset == 0:
		a.insnf("mov %s, [%s]", a.reg4(dest), a.reg(base))
	case offset > 0:
		a.insnf("mov %s, [%s + %d]", a.reg4(dest), a.reg(base), offset)
	default:
		a.insnf("mov %s, [%s - %d]", a.reg4(dest), a.reg(base), -offset)
	}
	a.Set(dest)
}

func (a *amd64) LoadByte(dest, base Reg, offset int) {
	a.check(base)
	switch {
	case offset == 0:
		a.insnf("mov %s, [%s]", a.reg1(dest), a.reg(base))
	case offset > 0:
		a.insnf("mov %s, [%s + %d]", a.reg1(dest), a.reg(base), offset)
	default:
		a.insnf("mov %s, [%s - %d]", a.reg1(dest), a.reg(base), -offset)
	}
	a.Set(dest)
}

func (a *amd64) Store(base Reg, offset int, src Reg) {
	a.check(base)
	a.check(src)
	switch {
	case offset == 0:
		a.insnf("mov [%s], %s", a.reg(base), a.reg(src))
	case offset > 0:
		a.insnf("mov [%s + %d], %s", a.reg(base), offset, a.reg(src))
	default:
		a.insnf("mov [%s - %d], %s", a.reg(base), -offset, a.reg(src))
	}
}

func (a *amd64) Store4Bytes(base Reg, offset int, src Reg) {
	a.check(base)
	a.check(src)
	switch {
	case offset == 0:
		a.insnf("mov [%s], %s", a.reg(base), a.reg4(src))
	case offset > 0:
		a.insnf("mov [%s + %d], %s", a.reg(base), offset, a.reg4(src))
	default:
		a.insnf("mov [%s - %d], %s", a.reg(base), -offset, a.reg4(src))
	}
}

func (a *amd64) Push(r Reg) {
	a.check(r)
	a.insn("push", a.reg(r))
}

func (a *amd64) Pop(r Reg) {
	a.insn("pop", a.reg(r))
	a.Set(r)
}

func (a *amd64) Jump(name string) {
	a.insn("jmp", symbol(name))
}

func (a *amd64) JumpRegRoutine(r Reg, internalNamePrefix string) {
	a.check(r)
	a.Call(internalNamePrefix + "_setup")

	a.label(internalNamePrefix + "_capture")
	a.insn("pause")
	a.Jump(internalNamePrefix + "_capture")

	a.FunctionWithoutPrologue(internalNamePrefix + "_setup")
	a.Store(a.StackPtr, 0, r)
	a.MoveImm(r, 0)
	a.Return()
}

func (a *amd64) JumpIfBitSet(r Reg, bit uint, name string) {
	a.check(r)
	switch {
	case bit < 31:
		a.insn("test", a.reg4(r), a.imm(1<<bit))
	default:
		a.insn("test", a.reg(r), a.imm(1<<bit))
	}
	a.insn("jne", symbol(name))
}

func (a *amd64) JumpIfBitNotSet(r Reg, bit uint, name string) {
	a.check(r)
	switch {
	case bit < 31:
		a.insn("test", a.reg4(r), a.imm(1<<bit))
	default:
		a.insn("test", a.reg(r), a.imm(1<<bit))
	}
	a.insn("je", symbol(name))
}

func (a *amd64) JumpIfImm(c Cond, r Reg, value int, name string) {
	a.check(r)
	switch {
	case value == 0 && (c == EQ || c == NE):
		a.insn("test", a.reg(r), a.reg(r))
	default:
		a.insn("cmp", a.reg(r), a.imm(value))
	}
	a.insn("j"+a.cond(c), symbol(name))
}

func (a *amd64) JumpIfReg(c Cond, dest, src Reg, name string) {
	a.check(dest)
	a.check(src)
	a.insn("cmp", a.reg(dest), a.reg(src))
	a.insn("j"+a.cond(c), symbol(name))
}

func (a *amd64) Call(name string) {
	a.insn("call", symbol(name))
}

func (a *amd64) Syscall(nr Syscall) {
	a.MoveImm(a.SyscallNr, nr.AMD64)
	a.insn("syscall")
	a.Set(a.SysResult)
}

func (a *amd64) Unreachable() {
	a.insn("int3")
}

func (a *amd64) speculationBarrier() {
	a.insn("int3")
}

func (a *amd64) imm(x int) string {
	return fmt.Sprintf("%d", x)
}

func (a *amd64) imm64(x uint64) string {
	return fmt.Sprintf("%d", x)
}

func (a *amd64) reg(x Reg) string {
	return x.AMD64.reg()
}

func (a *amd64) reg4(x Reg) string {
	return x.AMD64.reg4()
}

func (a *amd64) reg1(x Reg) string {
	return x.AMD64.reg1()
}

func (a *amd64) floatreg(x FloatReg) string {
	return fmt.Sprintf("xmm%d", x)
}

func (a *amd64) cond(x Cond) string {
	switch x {
	case EQ:
		return "e"
	case NE:
		return "ne"
	case LT:
		return "l"
	case LE:
		return "le"
	case GT:
		return "g"
	case GE:
		return "ge"
	}

	panic(x)
}

func (a *amd64) shift(x Shift) string {
	switch x {
	case Left:
		return "shl"
	case RightLogical:
		return "shr"
	case RightArithmetic:
		return "sar"
	}

	panic(x)
}
