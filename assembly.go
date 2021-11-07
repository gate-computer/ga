// Copyright (c) 2021 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ga

import (
	"bytes"
	"fmt"
	"strings"
)

const header1 = `// Generated by gate.computer/ga, DO NOT EDIT!
`

const header2 = `
.section .note.GNU-stack,"",%progbits
.section .text
`

type Cond uint8

const (
	EQ Cond = iota
	NE
	LT
	LE
	GT
	GE
)

type Shift uint8

const (
	Left Shift = iota
	RightLogical
	RightArithmetic
)

type FloatReg uint8

func global(name string) bool {
	return !strings.HasPrefix(name, ".")
}

func symbol(name string) string {
	if global(name) {
		return name
	}
	return strings.Replace(name, ".", ".L", 1)
}

type Assembly struct {
	Arch
	ArchAssembly
	*System
	*buffer
}

func NewAssembly(arch Arch, sys *System) *Assembly {
	buf := new(buffer)
	return &Assembly{
		Arch:         arch,
		ArchAssembly: arch.newAssembly(sys, buf),
		System:       sys,
		buffer:       buf,
	}
}

func (a *Assembly) String() string {
	return a.buffer.String()
}

type ArchAssembly interface {
	Label(name string)
	FunctionEpilogue()
	Function(name string)
	FunctionWithoutPrologue(name string)
	Return()
	ReturnWithoutEpilogue()
	Address(dest Reg, name string)
	MoveDef(dest Reg, name string)
	MoveImm(dest Reg, value int)
	MoveImm64(dest Reg, value uint64)
	MoveReg(dest, src Reg)
	MoveRegFloat(dest Reg, src FloatReg)
	AddImm(dest, src Reg, value int)
	AddReg(dest, src1, src2 Reg)
	SubtractImm(dest Reg, value int)
	SubtractReg(dest, src Reg)
	MultiplyImm(dest, src Reg, value int, temp Reg)
	AndImm(dest Reg, value int)
	AndReg(dest, src Reg)
	OrImm(dest Reg, value int)
	OrReg(dest, src Reg)
	ShiftImm(s Shift, r Reg, count int)
	Load(dest, base Reg, offset int)
	Load4Bytes(dest, base Reg, offset int)
	LoadByte(dest, base Reg, offset int)
	Store(base Reg, offset int, src Reg)
	Store4Bytes(base Reg, offset int, src Reg)
	Push(Reg)
	Pop(Reg)
	Call(name string)
	Jump(name string)
	JumpRegRoutine(r Reg, internalNamePrefix string)
	JumpIfBitSet(r Reg, bit uint, name string)
	JumpIfBitNotSet(r Reg, bit uint, name string)
	JumpIfImm(c Cond, r Reg, value int, name string)
	JumpIfReg(c Cond, dest, src Reg, name string)
	Syscall(Syscall)
	Unreachable()
}

type buffer struct {
	bytes.Buffer
}

func (b *buffer) label(name string) {
	b.printf("%s:", symbol(name))
}

func (b *buffer) insn(mnemonic string, operands ...string) {
	b.WriteString("\t" + mnemonic)

	for i, field := range operands {
		switch i {
		case 0:
			b.WriteString("\t")
		default:
			b.WriteString(", ")
		}

		b.WriteString(field)
	}

	fmt.Fprintln(&b.Buffer)
}

func (b *buffer) insnf(format string, args ...interface{}) {
	for i, field := range strings.Fields(fmt.Sprintf(format, args...)) {
		switch i {
		case 0, 1:
			b.WriteString("\t")
		default:
			b.WriteString(" ")
		}

		b.WriteString(field)
	}

	fmt.Fprintln(&b.Buffer)
}

func (b *buffer) printf(format string, args ...interface{}) {
	for i, field := range strings.Fields(fmt.Sprintf(format, args...)) {
		switch i {
		case 0:
		case 1:
			b.WriteString("\t")
		default:
			b.WriteString(" ")
		}

		b.WriteString(field)
	}

	fmt.Fprintln(&b.Buffer)
}