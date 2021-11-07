// Copyright (c) 2021 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ga

type System struct {
	StackPtr  Reg
	SyscallNr Reg
	SysParams []Reg
	SysResult Reg
	LibParams []Reg
	LibResult Reg
}

func Linux() *System {
	return &System{
		StackPtr:  Reg{RSP, XSP},
		SyscallNr: Reg{RAX, X8},
		SysParams: []Reg{{RDI, X0}, {RSI, X1}, {RDX, X2}, {R10, X3}, {R8, X4}, {R9, X5}},
		SysResult: Reg{RAX, X0},
		LibParams: []Reg{{RDI, X0}, {RSI, X1}, {RDX, X2}, {RCX, X3}, {R8, X4}, {R9, X5}},
		LibResult: Reg{RAX, X0},
	}
}
