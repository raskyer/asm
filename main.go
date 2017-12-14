package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/leaklessgfy/asm/asm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Bad usage")
		os.Exit(1)
	}

	bytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reader, err := asm.NewClassReader(bytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reader.Accept(&SimpleVisitor{
		OnVisit: func(version, access int, name, signature, superName string, interfaces []string) {
			fmt.Println(name, signature, superName, interfaces)
		},
		OnVisitMethod: func(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor {
			fmt.Println(name)
			return nil
		},
		OnVisitEnd: func() {
			fmt.Println("End")
		},
	}, asm.EXPAND_FRAMS)
}
