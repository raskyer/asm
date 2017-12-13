package asm

import "errors"

const FLAG_DEBUG_ONLY = 1
const FLAG_JUMP_TARGET = 2
const FLAG_RESOLVED = 4
const FLAG_REACHABLE = 8
const FLAG_SUBROUTINE_CALLER = 16
const FLAG_SUBROUTINE_START = 32
const FLAG_SUBROUTINE_BODY = 64
const FLAG_SUBROUTINE_END = 128
const LINE_NUMBERS_CAPACITY_INCREMENT = 4
const VALUES_CAPACITY_INCREMENT = 6
const FORWARD_REFERENCE_TYPE_MASK = 0xF0000000
const FORWARD_REFERENCE_TYPE_SHORT = 0x10000000
const FORWARD_REFERENCE_TYPE_WIDE = 0x20000000
const FORWARD_REFERENCE_HANDLE_MASK = 0x0FFFFFFF

var EMPTY_LIST = &Label{}

type Label struct {
	info             interface{}
	flags            int16
	lineNumber       int16
	otherLineNumbers []int
	bytecodeOffset   int
	valueCount       int16
	values           []int
	inputStackSize   int16
	outputStackSize  int16
	outputStackMax   int16
	frame            interface{} //Frame
	nextBasicBlock   *Label
	outgoingEdges    interface{} //Edge
	nextListElement  *Label
}

func (l Label) getOffset() (int, error) {
	if (l.flags & FLAG_RESOLVED) == 0 {
		return 0, errors.New("Illegal State - Label offset position has not been resolved yet")
	}
	return l.bytecodeOffset, nil
}

func (l Label) getCanonicalInstance() *Label {
	if l.frame == nil {
		return &l
	}
	return nil //l.frame.owner
}

func (l *Label) addLineNumber(lineNumber int) {
	if l.lineNumber == 0 {
		l.lineNumber = int16(lineNumber)
	} else {
		if l.otherLineNumbers == nil {
			l.otherLineNumbers = make([]int, LINE_NUMBERS_CAPACITY_INCREMENT)
		}
		otherLineNumberCount := l.otherLineNumbers[0]
		l.otherLineNumbers[0]++
		if otherLineNumberCount >= len(l.otherLineNumbers) {
			newLineNumbers := make([]int, len(l.otherLineNumbers)+VALUES_CAPACITY_INCREMENT)
			//System.arraycopy(l.otherLineNumbers, 0, newLineNumbers, 0, len(l.otherLineNumbers))
			l.otherLineNumbers = newLineNumbers
		}
		l.otherLineNumbers[otherLineNumberCount] = lineNumber
	}
}

func (l Label) accept(methodVisitor MethodVisitor, visitLineNumbers bool) {
	methodVisitor.visitLabel(&l)
	if visitLineNumbers && l.lineNumber != 0 {
		methodVisitor.visitLineNumber(int(l.lineNumber)&0xFFFF, &l)
		if l.otherLineNumbers != nil {
			for i := 1; i <= l.otherLineNumbers[0]; i++ {
				methodVisitor.visitLineNumber(l.otherLineNumbers[i], &l)
			}
		}
	}
}
