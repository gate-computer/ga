// Copyright (c) 2021 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const header = `// Generated by internal/generate-syscalls.go, DO NOT EDIT!

package linux

import (
	"gate.computer/ga"
)

var (
`

const footer = `)
`

func main() {
	archSyms := make(map[string]map[string]int)

	for _, arch := range []string{"AMD64", "ARM64"} {
		lowarch := strings.ToLower(arch)
		filename := path.Join(runtime.GOROOT(), "src/cmd/vendor/golang.org/x/sys/unix", fmt.Sprintf("zsysnum_linux_%s.go", lowarch))
		syms, err := parse(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filename, err)
			os.Exit(1)
		}
		archSyms[arch] = syms
	}

	var (
		symArchs = make(map[string]map[string]int)
		syms     []string
		seenSyms = make(map[string]struct{})
		symWidth int
	)

	for arch, syscalls := range archSyms {
		for sym, nr := range syscalls {
			archs, found := symArchs[sym]
			if !found {
				archs = make(map[string]int)
				symArchs[sym] = archs
			}
			archs[arch] = nr

			if _, seen := seenSyms[sym]; !seen {
				syms = append(syms, sym)
				seenSyms[sym] = struct{}{}

				if len(sym) > symWidth {
					symWidth = len(sym)
				}
			}
		}
	}

	sort.Strings(syms)

	symFormat := fmt.Sprintf("\t%%-%ds = ga.Syscall{", symWidth)

	b := bytes.NewBuffer(nil)
	b.WriteString(header)

	for _, sym := range syms {
		archs := symArchs[sym]
		if len(archs) != 2 {
			continue
		}

		var names []string
		for name := range archs {
			names = append(names, name)
		}

		sort.Strings(names)

		fmt.Fprintf(b, symFormat, sym)
		sep := ""
		for _, name := range names {
			fmt.Fprint(b, sep)
			fmt.Fprintf(b, "%s: %d", name, archs[name])
			sep = ", "
		}
		fmt.Fprintln(b, "}")
	}

	b.WriteString(footer)

	if err := ioutil.WriteFile("linux/syscall.go", b.Bytes(), 0666); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var re = regexp.MustCompile(`^\s*(SYS_[A-Z0-9_]+)\s*=\s*([0-9]+)\b`)

func parse(filename string) (map[string]int, error) {
	syscalls := make(map[string]int)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(data)

	for {
		line, err := r.ReadString('\n')
		match := re.FindStringSubmatch(line)
		if len(match) != 0 {
			sym := match[1]
			num, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, err
			}
			syscalls[sym] = num
		}

		if err == nil {
			continue
		}
		if err == io.EOF {
			break
		}
		return nil, err
	}

	return syscalls, nil
}
