package asm

// Context information about a class being parsed in a {@link ClassReader}.
type Context struct {
	attributePrototypes                        []Attribute
	parsingOptions                             int
	charBuffer                                 []rune
	bootstrapMethodOffsets                     []int
	currentMethodAccessFlags                   int
	currentMethodName                          string
	currentMethodDescriptor                    string
	currentMethodLabels                        []*Label
	currentTypeAnnotationTarget                int
	currentTypeAnnotationTargetPath            *TypePath
	currentLocalVariableAnnotationRangeStarts  []*Label
	currentLocalVariableAnnotationRangeEnds    []*Label
	currentLocalVariableAnnotationRangeIndices []int
	currentFrameOffset                         int
	currentFrameType                           int
	currentFrameLocalCount                     int
	currentFrameLocalCountDelta                int
	currentFrameLocalTypes                     []interface{}
	currentFrameStackCount                     int
	currentFrameStackTypes                     []interface{}
}
