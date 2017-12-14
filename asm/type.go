package asm

import "github.com/leaklessgfy/asm/asm/typed"

type Type struct {
	sort        int
	valueBuffer []rune
	valueOffset int
	valueLength int
}

func getObjectType(internalName string) *Type {
	valueBuffer := []rune(internalName)
	typ := typed.INTERNAL
	if valueBuffer[0] == '[' {
		typ = typed.ARRAY
	}
	return &Type{
		sort:        typ,
		valueBuffer: valueBuffer,
		valueOffset: 0,
		valueLength: len(valueBuffer),
	}
}

func getMethodType(methodDescriptor string) *Type {
	valueBuffer := []rune(methodDescriptor)
	return &Type{
		typed.METHOD,
		valueBuffer,
		0,
		len(valueBuffer),
	}
}
