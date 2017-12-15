package asm

import "github.com/leaklessgfy/asm/asm/typed"

type Type struct {
	sort        int
	valueBuffer []rune
	valueOffset int
	valueLength int
}

func getType(typeDescriptor string) *Type {
	valueBuffer := []rune(typeDescriptor)
	return getTypeB(valueBuffer, 0, len(valueBuffer))
}

func getTypeB(descriptorBuffer []rune, descriptorOffset int, descriptorLength int) *Type {
	switch descriptorBuffer[descriptorOffset] {
	case 'V':
		return &Type{typed.VOID, typed.PRIMITIVE_DESCRIPTORS, typed.VOID, 1}
	case 'Z':
		return &Type{typed.BOOLEAN, typed.PRIMITIVE_DESCRIPTORS, typed.BOOLEAN, 1}
	case 'C':
		return &Type{typed.CHAR, typed.PRIMITIVE_DESCRIPTORS, typed.CHAR, 1}
	case 'B':
		return &Type{typed.BYTE, typed.PRIMITIVE_DESCRIPTORS, typed.BYTE, 1}
	case 'S':
		return &Type{typed.SHORT, typed.PRIMITIVE_DESCRIPTORS, typed.SHORT, 1}
	case 'I':
		return &Type{typed.INT, typed.PRIMITIVE_DESCRIPTORS, typed.INT, 1}
	case 'F':
		return &Type{typed.FLOAT, typed.PRIMITIVE_DESCRIPTORS, typed.FLOAT, 1}
	case 'J':
		return &Type{typed.LONG, typed.PRIMITIVE_DESCRIPTORS, typed.LONG, 1}
	case 'D':
		return &Type{typed.DOUBLE, typed.PRIMITIVE_DESCRIPTORS, typed.DOUBLE, 1}
	case '[':
		return &Type{typed.ARRAY, descriptorBuffer, descriptorOffset, descriptorLength}
	case 'L':
		return &Type{typed.OBJECT, descriptorBuffer, descriptorOffset + 1, descriptorLength - 2}
	case '(':
		return &Type{typed.METHOD, descriptorBuffer, descriptorOffset, descriptorLength}
	default:
		//throw new AssertionError
		break
	}
	return nil
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
