// Copyright (c) 2021 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ga

// Specific value per CPU architecture.
type Specific struct {
	AMD64 int
	ARM64 int
}

// Reg ister per CPU architecture.
type Reg struct {
	AMD64 RegAMD64
	ARM64 RegARM64
}

// Syscall number per CPU architecture.
type Syscall Specific

// Arch itecture of CPU.
type Arch interface {
	Machine() string      // GNU-style CPU architecture name (x86_64, aarch64).
	Specify(Specific) int // Get value for the CPU architecture.
	newAssembly(*System, *buffer) ArchAssembly
}

// Indexed by Go-style CPU architecture name (amd64, arm64).
var Archs = map[string]Arch{
	"amd64": AMD64,
	"arm64": ARM64,
}
