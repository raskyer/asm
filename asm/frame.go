package asm

type Frame struct {
	owner               *Label
	inputLocals         []int
	inputStack          []int
	outputLocals        []int
	outputStack         []int
	outputStackStart    int16
	outputStackTop      int16
	initializationCount int
	initializations     []int
}
