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
		StackPtr:  Reg{RSP, XSP, "stack"},
		SyscallNr: Reg{RAX, X8, "syscall"},
		SysParams: []Reg{
			{RDI, X0, "sysparam0"},
			{RSI, X1, "sysparam1"},
			{RDX, X2, "sysparam2"},
			{R10, X3, "sysparam3"},
			{R8, X4, "sysparam4"},
			{R9, X5, "sysparam5"},
		},
		SysResult: Reg{RAX, X0, "sysresult"},
		LibParams: []Reg{
			{RDI, X0, "libparam0"},
			{RSI, X1, "libparam1"},
			{RDX, X2, "libparam2"},
			{RCX, X3, "libparam3"},
			{R8, X4, "libparam4"},
			{R9, X5, "libparam5"},
		},
		LibResult: Reg{RAX, X0, "libresult"},
	}
}
