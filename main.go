package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/leaklessgfy/asm/asm"
	"github.com/leaklessgfy/asm/asm/helper"
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

	reader.Accept(&helper.ClassVisitor{
		OnVisit: func(version, access int, name, signature, superName string, interfaces []string) {
		},
		OnVisitField: func(access int, name, descriptor, signature string, value interface{}) asm.FieldVisitor {
			return nil
		},
		OnVisitMethod: func(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor {
			return &helper.MethodVisitor{
				OnVisitLineNumber: func(line int, start *asm.Label) {
					fmt.Println(line)
				},
			}
		},
		OnVisitEnd: func() {
		},
	}, 0)
}
