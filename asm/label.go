package asm

import "errors"
import "math"
import "github.com/leaklessgfy/asm/asm/opcodes"
import "github.com/leaklessgfy/asm/asm/constants"

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
	frame            *Frame
	nextBasicBlock   *Label
	outgoingEdges    *Edge
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
	return l.frame.owner
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
			copy(newLineNumbers, l.otherLineNumbers) //System.arraycopy(l.otherLineNumbers, 0, newLineNumbers, 0, len(l.otherLineNumbers))
			l.otherLineNumbers = newLineNumbers
		}
		l.otherLineNumbers[otherLineNumberCount] = lineNumber
	}
}

func (l Label) accept(methodVisitor MethodVisitor, visitLineNumbers bool) {
	methodVisitor.VisitLabel(&l)
	if visitLineNumbers && l.lineNumber != 0 {
		methodVisitor.VisitLineNumber(int(l.lineNumber)&0xFFFF, &l)
		if l.otherLineNumbers != nil {
			for i := 1; i <= l.otherLineNumbers[0]; i++ {
				methodVisitor.VisitLineNumber(l.otherLineNumbers[i], &l)
			}
		}
	}
}

func (l Label) put() {
	//TODO
}

func (l *Label) addForwardReference(sourceInsnBytecodeOffset, referenceType, referenceHandle int) {
	if l.values == nil {
		l.values = make([]int, VALUES_CAPACITY_INCREMENT)
	}
	if int(l.valueCount) >= len(l.values) {
		newValues := make([]int, len(l.values)+VALUES_CAPACITY_INCREMENT)
		copy(newValues, l.values)
		l.values = newValues
	}
	l.values[l.valueCount] = sourceInsnBytecodeOffset
	l.valueCount++
	l.values[l.valueCount] = referenceType | referenceHandle
	l.valueCount++
}

func (l *Label) resolve(code []byte, bytecodeOffset int) bool {
	l.flags |= FLAG_RESOLVED
	l.bytecodeOffset = bytecodeOffset
	hasAsmInstructions := false
	for i := 0; i < int(l.valueCount); i += 2 {
		sourceInsnBytecodeOffset := l.values[i]
		reference := l.values[i+1]
		relativeOffset := bytecodeOffset - sourceInsnBytecodeOffset
		handle := reference & FORWARD_REFERENCE_HANDLE_MASK
		if (reference & FORWARD_REFERENCE_HANDLE_MASK) == FORWARD_REFERENCE_TYPE_SHORT {
			if relativeOffset < math.MinInt16 || relativeOffset > math.MaxInt16 {
				opcode := code[sourceInsnBytecodeOffset] & 0xFF
				if opcode < opcodes.IFNULL {
					code[sourceInsnBytecodeOffset] = opcode + constants.ASM_OPCODE_DELTA
				} else {
					code[sourceInsnBytecodeOffset] = opcode + constants.ASM_IFNULL_OPCODE_DELTA
				}
				hasAsmInstructions = true
			}
			code[handle] = byte(relativeOffset >> 8) // >>> x3 ?
			handle++
			code[handle] = byte(relativeOffset)
		} else {
			code[handle] = byte(relativeOffset >> 14)
			handle++
			code[handle] = byte(relativeOffset >> 16)
			handle++
			code[handle] = byte(relativeOffset >> 8)
			handle++
			code[handle] = byte(relativeOffset)
		}
	}
	return hasAsmInstructions
}

func (l *Label) markSubroutine(subroutineID int, numSubroutine int) {
	listOfBlocksToProcess := l
	listOfBlocksToProcess.nextListElement = EMPTY_LIST
	for listOfBlocksToProcess != EMPTY_LIST {
		basicBlock := listOfBlocksToProcess
		listOfBlocksToProcess = listOfBlocksToProcess.nextListElement
		basicBlock.nextListElement = nil
		if !basicBlock.isInSubroutine(subroutineID) {
			basicBlock.addToSubroutine(subroutineID, numSubroutine)
			listOfBlocksToProcess = basicBlock.pushSuccessors(listOfBlocksToProcess)
		}
	}
}

func (l *Label) addSubroutineRetSuccessors(subroutineCaller *Label, numSubroutine int) {
	listOfProcessedBlocks := EMPTY_LIST
	listOfBlocksToProcess := l
	listOfBlocksToProcess.nextListElement = EMPTY_LIST
	for listOfBlocksToProcess != EMPTY_LIST {
		basicBlock := listOfBlocksToProcess
		listOfBlocksToProcess = basicBlock.nextListElement
		basicBlock.nextListElement = listOfProcessedBlocks
		listOfProcessedBlocks = basicBlock

		if (basicBlock.flags&FLAG_SUBROUTINE_END) != 0 && !basicBlock.isInSameSubroutine(subroutineCaller) {
			basicBlock.outgoingEdges = NewEdge(int(basicBlock.outputStackSize), subroutineCaller.outgoingEdges.successor, basicBlock.outgoingEdges)
		}

		listOfBlocksToProcess = basicBlock.pushSuccessors(listOfBlocksToProcess)
	}

	for listOfProcessedBlocks != EMPTY_LIST {
		nextListElement := listOfProcessedBlocks.nextListElement
		listOfProcessedBlocks.nextListElement = nil
		listOfProcessedBlocks = nextListElement
	}
}

func (l *Label) pushSuccessors(listOfLabelsToProcess *Label) *Label {
	outgoingEdge := l.outgoingEdges
	for outgoingEdge != nil {
		isJsrTarget := (l.flags&FLAG_SUBROUTINE_CALLER) != 0 && outgoingEdge == outgoingEdge.nextEdge
		if !isJsrTarget {
			if outgoingEdge.successor.nextListElement == nil {
				outgoingEdge.successor.nextListElement = listOfLabelsToProcess
				listOfLabelsToProcess = outgoingEdge.successor
			}
		}
		outgoingEdge = outgoingEdge.nextEdge
	}
	return listOfLabelsToProcess
}

func (l Label) isInSubroutine(subroutineID int) bool {
	if (l.flags & FLAG_SUBROUTINE_BODY) != 0 {
		return (l.values[subroutineID/32] & (1 << (uint(subroutineID) % 32))) != 0
	}
	return false
}

func (l Label) isInSameSubroutine(basicBlock *Label) bool {
	if (l.flags&FLAG_SUBROUTINE_BODY) == 0 || (basicBlock.flags&FLAG_SUBROUTINE_BODY) == 0 {
		return false
	}
	for i := 0; i < len(l.values); i++ {
		if (l.values[i] & basicBlock.values[i]) != 0 {
			return true
		}
	}
	return false
}

func (l *Label) addToSubroutine(subroutineID int, numSubroutine int) {
	if (l.flags & FLAG_SUBROUTINE_BODY) == 0 {
		l.flags |= FLAG_SUBROUTINE_BODY
		l.values = make([]int, numSubroutine/32+1)
	}
	l.values[subroutineID/32] |= (1 << (uint(subroutineID) % 32))
}
