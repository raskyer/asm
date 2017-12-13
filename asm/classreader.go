package asm

import (
	"errors"
	"fmt"

	"github.com/leaklessgfy/asm/asm/opcodes"
	"github.com/leaklessgfy/asm/asm/symbol"
)

// ClassReader A parser to make a {@link ClassVisitor} visit a ClassFile structure, as defined in the Java
// Virtual Machine Specification (JVMS). This class parses the ClassFile content and calls the
// appropriate visit methods of a given {@link ClassVisitor} for each field, method and bytecode
// instruction encountered.
type ClassReader struct {
	b                  []byte
	cpInfoOffsets      []int
	constantUtf8Values []string
	maxStringLength    int
	header             int
}

// SKIP_CODE a flag to skip the Code attributes. If this flag is set the Code attributes are neither parsed nor visited.
const SKIP_CODE = 1

// SKIP_DEBUG a flag to skip the SourceFile, SourceDebugExtension, LocalVariableTable, LocalVariableTypeTable
// and LineNumberTable attributes. If this flag is set these attributes are neither parsed nor
// visited (i.e. {@link ClassVisitor#visitSource}, {@link MethodVisitor#visitLocalVariable} and
// {@link MethodVisitor#visitLineNumber} are not called).
const SKIP_DEBUG = 2

// SKIP_FRAMES a flag to skip the StackMap and StackMapTable attributes. If this flag is set these attributes
// are neither parsed nor visited (i.e. {@link MethodVisitor#visitFrame} is not called). This flag
// is useful when the {@link ClassWriter#COMPUTE_FRAMES} option is used: it avoids visiting frames
// that will be ignored and recomputed from scratch.
const SKIP_FRAMES = 4

// EXPAND_FRAMS a flag to expand the stack map frames. By default stack map frames are visited in their
// original format (i.e. "expanded" for classes whose version is less than V1_6, and "compressed"
// for the other classes). If this flag is set, stack map frames are always visited in expanded
// format (this option adds a decompression/compression step in ClassReader and ClassWriter which
// degrades performance quite a lot).
const EXPAND_FRAMS = 8

// EXPAND_ASM_INSNS A flag to expand the ASM specific instructions into an equivalent sequence of standard bytecode
// instructions. When resolving a forward jump it may happen that the signed 2 bytes offset
// reserved for it is not sufficient to store the bytecode offset. In this case the jump
// instruction is replaced with a temporary ASM specific instruction using an unsigned 2 bytes
// offset (see {@link Label#resolve}). This internal flag is used to re-read classes containing
// such instructions, in order to replace them with standard instructions. In addition, when this
// flag is used, goto_w and jsr_w are <i>not</i> converted into goto and jsr, to make sure that
// infinite loops where a goto_w is replaced with a goto in ClassReader and converted back to a
// goto_w in ClassWriter cannot occur.
const EXPAND_ASM_INSNS = 256

// NewClassReader constructs a new {@link ClassReader} object.
func NewClassReader(classFile []byte) (*ClassReader, error) {
	return classReader(classFile, 0, len(classFile))
}

func classReader(byteBuffer []byte, offset int, length int) (*ClassReader, error) {
	reader := &ClassReader{
		b: byteBuffer,
	}

	if reader.readShort(offset+6) > opcodes.V10 {
		return nil, errors.New("Illegal Argument")
	}

	constantPoolCount := reader.readUnsignedShort(offset + 8)
	reader.cpInfoOffsets = make([]int, constantPoolCount)
	reader.constantUtf8Values = make([]string, constantPoolCount)
	currentCpInfoOffset := offset + 10
	maxStringLength := 0

	for i := 1; i < constantPoolCount; i++ {
		reader.cpInfoOffsets[i] = currentCpInfoOffset + 1
		var cpInfoSize int

		switch byteBuffer[currentCpInfoOffset] {
		case byte(symbol.CONSTANT_FIELDREF_TAG), byte(symbol.CONSTANT_METHODREF_TAG), byte(symbol.CONSTANT_INTERFACE_METHODREF_TAG),
			byte(symbol.CONSTANT_INTEGER_TAG), byte(symbol.CONSTANT_FLOAT_TAG), byte(symbol.CONSTANT_NAME_AND_TYPE_TAG),
			byte(symbol.CONSTANT_INVOKE_DYNAMIC_TAG):
			cpInfoSize = 5
			break
		case byte(symbol.CONSTANT_LONG_TAG), byte(symbol.CONSTANT_DOUBLE_TAG):
			cpInfoSize = 9
			i++
			break
		case byte(symbol.CONSTANT_UTF8_TAG):
			cpInfoSize = 3 + reader.readUnsignedShort(currentCpInfoOffset+1)
			if cpInfoSize > maxStringLength {
				maxStringLength = cpInfoSize
			}
			break
		case byte(symbol.CONSTANT_METHOD_HANDLE_TAG):
			cpInfoSize = 4
			break
		case byte(symbol.CONSTANT_CLASS_TAG), byte(symbol.CONSTANT_STRING_TAG), byte(symbol.CONSTANT_METHOD_TYPE_TAG),
			byte(symbol.CONSTANT_PACKAGE_TAG), byte(symbol.CONSTANT_MODULE_TAG):
			cpInfoSize = 3
			break
		default:
			return nil, errors.New("Assertion Error")
		}
		currentCpInfoOffset += cpInfoSize
	}

	reader.maxStringLength = maxStringLength
	reader.header = currentCpInfoOffset

	return reader, nil
}

// -----------------------------------------------------------------------------------------------
// Accessors
// -----------------------------------------------------------------------------------------------

// GetAccess returns the class's access flags (see {@link Opcodes}). This value may not reflect Deprecated
// and Synthetic flags when bytecode is before 1.5 and those flags are represented by attributes.
func (c *ClassReader) GetAccess() int {
	return c.readUnsignedShort(c.header)
}

// GetClassName returns the internal name of the class (see {@link Type#getInternalName()}).
func (c *ClassReader) GetClassName() string {
	charBuffer := make([]rune, c.maxStringLength)
	return c.readClass(c.header+2, charBuffer)
}

// GetSuperName returns the internal of name of the super class (see {@link Type#getInternalName()}). For
// interfaces, the super class is {@link Object}.
func (c *ClassReader) GetSuperName() string {
	charBuffer := make([]rune, c.maxStringLength)
	return c.readClass(c.header+4, charBuffer)
}

// GetInterfaces returns the internal names of the implemented interfaces (see {@link Type#getInternalName()}).
func (c ClassReader) GetInterfaces() []string {
	currentOffset := c.header + 6
	interfacesCount := c.readUnsignedShort(currentOffset)
	interfaces := make([]string, interfacesCount)
	if interfacesCount > 0 {
		charBuffer := make([]rune, c.maxStringLength)
		for i := 0; i < interfacesCount; i++ {
			currentOffset += 2
			interfaces[i] = c.readClass(currentOffset, charBuffer)
		}
	}
	return interfaces
}

// -----------------------------------------------------------------------------------------------
// Public methods
// -----------------------------------------------------------------------------------------------

// Accept Makes the given visitor visit the JVMS ClassFile structure passed to the constructor of this {@link ClassReader}.
func (c ClassReader) Accept(classVisitor ClassVisitor, parsingOptions int) {
	c.AcceptB(classVisitor, make([]Attribute, 0), parsingOptions)
}

// AcceptB Makes the given visitor visit the JVMS ClassFile structure passed to the constructor of this {@link ClassReader}.
func (c ClassReader) AcceptB(classVisitor ClassVisitor, attributePrototypes []Attribute, parsingOptions int) {
	context := &Context{
		attributePrototypes: attributePrototypes,
		parsingOptions:      parsingOptions,
		charBuffer:          make([]rune, c.maxStringLength),
	}

	charBuffer := context.charBuffer
	currentOffset := c.header
	accessFlags := c.readUnsignedShort(currentOffset)
	thisClass := c.readClass(currentOffset+2, charBuffer)
	superClass := c.readClass(currentOffset+4, charBuffer)
	interfaces := make([]string, c.readUnsignedShort(currentOffset+6))
	currentOffset += 8

	for i := 0; i < len(interfaces); i++ {
		interfaces[i] = c.readClass(currentOffset, charBuffer)
		currentOffset += 2
	}

	innerClassesOffset := 0
	enclosingMethodOffset := 0
	signature := ""
	sourceFile := ""
	sourceDebugExtension := ""
	runtimeVisibleAnnotationsOffset := 0
	runtimeInvisibleAnnotationsOffset := 0
	runtimeVisibleTypeAnnotationsOffset := 0
	runtimeInvisibleTypeAnnotationsOffset := 0
	moduleOffset := 0
	modulePackagesOffset := 0
	moduleMainClass := ""
	var attributes *Attribute

	currentAttributeOffset := c.getFirstAttributeOffset()
	for i := c.readUnsignedShort(currentAttributeOffset - 2); i > 0; i-- {
		attributeName := c.readUTF8(currentAttributeOffset, charBuffer)
		attributeLength := c.readInt(currentAttributeOffset + 2)
		currentAttributeOffset += 6

		switch attributeName {
		case "SourceFile":
			sourceFile = c.readUTF8(currentAttributeOffset, charBuffer)
			break
		case "InnerClasses":
			innerClassesOffset = currentAttributeOffset
			break
		case "EnclosingMethod":
			enclosingMethodOffset = currentAttributeOffset
			break
		case "Signature":
			signature = c.readUTF8(currentAttributeOffset, charBuffer)
			break
		case "RuntimeVisibleAnnotations":
			runtimeVisibleAnnotationsOffset = currentAttributeOffset
			break
		case "RuntimeVisibleTypeAnnotations":
			runtimeVisibleTypeAnnotationsOffset = currentAttributeOffset
			break
		case "Deprecated":
			accessFlags |= opcodes.ACC_DEPRECATED
			break
		case "Synthetic":
			accessFlags |= opcodes.ACC_SYNTHETIC
			break
		case "SourceDebugExtension":
			sourceDebugExtension = c.readUTFB(currentAttributeOffset, attributeLength, make([]rune, attributeLength))
			break
		case "RuntimeInvisibleAnnotations":
			runtimeInvisibleAnnotationsOffset = currentAttributeOffset
			break
		case "RuntimeInvisibleTypeAnnotations":
			runtimeInvisibleTypeAnnotationsOffset = currentAttributeOffset
			break
		case "Module":
			moduleOffset = currentAttributeOffset
			break
		case "ModuleMainClass":
			moduleMainClass = c.readClass(currentAttributeOffset, charBuffer)
			break
		case "ModulePackages":
			modulePackagesOffset = currentAttributeOffset
			break
		case "BootstrapMethods":
			bootstrapMethodOffsets := make([]int, c.readUnsignedShort(currentAttributeOffset))
			currentBootstrapMethodOffset := currentAttributeOffset + 2
			for j := 0; j < len(bootstrapMethodOffsets); j++ {
				bootstrapMethodOffsets[j] = currentBootstrapMethodOffset
				currentBootstrapMethodOffset += 4 + c.readUnsignedShort(currentBootstrapMethodOffset+2)*2
			}
			context.bootstrapMethodOffsets = bootstrapMethodOffsets
			break
		default:
			attribute := c.readAttribute(attributePrototypes, attributeName, currentAttributeOffset, attributeLength, charBuffer, -1, nil)
			attribute.nextAttribute = attributes
			attributes = attribute
		}
		currentAttributeOffset += attributeLength
	}

	classVisitor.Visit(c.readInt(c.cpInfoOffsets[1]-7), accessFlags, thisClass, signature, superClass, interfaces)

	if (parsingOptions&SKIP_DEBUG) == 0 && (sourceFile != "" || sourceDebugExtension != "") {
		classVisitor.VisitSource(sourceFile, sourceDebugExtension)
	}

	if moduleOffset != 0 {
		c.readModule(classVisitor, context, moduleOffset, modulePackagesOffset, moduleMainClass)
	}

	if enclosingMethodOffset != 0 {
		className := c.readClass(enclosingMethodOffset, charBuffer)
		methodIndex := c.readUnsignedShort(enclosingMethodOffset + 2)
		var name string
		var typed string
		if methodIndex != 0 {
			name = c.readUTF8(c.cpInfoOffsets[methodIndex], charBuffer)
			typed = c.readUTF8(c.cpInfoOffsets[methodIndex]+2, charBuffer)
		}
		classVisitor.VisitOuterClass(className, name, typed)
	}

	if runtimeVisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeVisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeVisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitAnnotation(annotationDescriptor, true), currentAnnotationOffset, true, charBuffer)
			numAnnotations--
		}
	}

	if runtimeInvisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitAnnotation(annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
			numAnnotations--
		}
	}

	if runtimeVisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitAnnotation(annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
			numAnnotations--
		}
	}

	if runtimeInvisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleTypeAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleTypeAnnotationsOffset + 2
		for numAnnotations > 0 {
			currentAnnotationOffset = c.readTypeAnnotationTarget(context, currentAnnotationOffset)
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitTypeAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath.(int), annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
			numAnnotations--
		}
	}

	for attributes != nil {
		nextAttribute := attributes.nextAttribute
		attributes.nextAttribute = nil
		classVisitor.VisitAttribute(attributes)
		attributes = nextAttribute
	}

	if innerClassesOffset != 0 {
		numberOfClasses := c.readUnsignedShort(innerClassesOffset)
		currentClassesOffset := innerClassesOffset + 2
		for numberOfClasses > 0 {
			classVisitor.VisitInnerClass(c.readClass(currentClassesOffset, charBuffer), c.readClass(currentAttributeOffset+2, charBuffer), c.readClass(currentClassesOffset+4, charBuffer), c.readUnsignedShort(currentClassesOffset+6))
			currentClassesOffset += 8
			numberOfClasses--
		}
	}

	fieldsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for fieldsCount > 0 {
		currentOffset = c.readField(classVisitor, context, currentOffset)
		fieldsCount--
	}
	methodsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for methodsCount > 0 {
		currentOffset = c.readMethod(classVisitor, context, currentOffset)
		methodsCount--
	}

	classVisitor.VisitEnd()
}

// ----------------------------------------------------------------------------------------------
// Methods to parse modules, fields and methods
// ----------------------------------------------------------------------------------------------

func (c ClassReader) readModule(classVisitor ClassVisitor, context *Context, moduleOffset int, modulePackagesOffset int, moduleMainClass string) {
	buffer := context.charBuffer
	currentOffset := moduleOffset
	moduleName := c.readModuleB(currentOffset, buffer)
	moduleFlags := c.readUnsignedShort(currentOffset + 2)
	moduleVersion := c.readUTF8(currentOffset+4, buffer)
	currentOffset += 6
	moduleVisitor := classVisitor.VisitModule(moduleName, moduleFlags, moduleVersion)
	if moduleVisitor == nil {
		return
	}

	if modulePackagesOffset != 0 {
		packageCount := c.readUnsignedShort(modulePackagesOffset)
		currentPackageOffset := modulePackagesOffset + 2
		for packageCount > 0 {
			moduleVisitor.VisitPackage(c.readPackage(currentPackageOffset, buffer))
			currentPackageOffset += 2
			packageCount--
		}
	}

	requiresCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for requiresCount > 0 {
		requires := c.readModuleB(currentOffset, buffer)
		requiresFlags := c.readUnsignedShort(currentOffset + 2)
		requiresVersion := c.readUTF8(currentOffset+4, buffer)
		currentOffset += 6
		moduleVisitor.VisitRequire(requires, requiresFlags, requiresVersion)
		requiresCount--
	}

	exportsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for exportsCount > 0 {
		exports := c.readPackage(currentOffset, buffer)
		exportsFlags := c.readUnsignedShort(currentOffset + 2)
		exportsToCount := c.readUnsignedShort(currentOffset + 4)
		currentOffset += 6
		var exportsTo []string
		if exportsToCount != 0 {
			exportsTo = make([]string, exportsToCount)
			for i := 0; i < exportsToCount; i++ {
				exportsTo[i] = c.readModuleB(currentOffset, buffer)
				currentOffset += 2
			}
		}
		moduleVisitor.VisitExport(exports, exportsFlags, exportsTo...)
		exportsCount--
	}

	opensCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for opensCount > 0 {
		opens := c.readPackage(currentOffset, buffer)
		opensFlags := c.readUnsignedShort(currentOffset + 2)
		opensToCount := c.readUnsignedShort(currentOffset + 4)
		currentOffset += 6
		var opensTo []string
		if opensToCount != 0 {
			opensTo = make([]string, opensToCount)
			for i := 0; i < opensToCount; i++ {
				opensTo[i] = c.readModuleB(currentOffset, buffer)
				currentOffset += 2
			}
		}
		moduleVisitor.VisitOpen(opens, opensFlags, opensTo...)
	}

	usesCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for usesCount > 0 {
		moduleVisitor.VisitUse(c.readClass(currentOffset, buffer))
		currentOffset += 2
		usesCount--
	}

	providesCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for providesCount > 0 {
		provides := c.readClass(currentOffset, buffer)
		providesWithCount := c.readUnsignedShort(currentOffset + 2)
		currentOffset += 4
		providesWith := make([]string, providesWithCount)
		for i := 0; i < providesWithCount; i++ {
			providesWith[i] = c.readClass(currentOffset, buffer)
			currentOffset += 2
		}
		moduleVisitor.VisitProvide(provides, providesWith...)
		providesCount--
	}

	moduleVisitor.VisitEnd()
}

func (c ClassReader) readField(classVisitor ClassVisitor, context *Context, fieldInfoOffset int) int {
	//TODO
	return 0
}

func (c ClassReader) readMethod(classVisitor ClassVisitor, context *Context, methodInfoOffset int) int {
	//TODO
	return 0
}

// ----------------------------------------------------------------------------------------------
// Methods to parse a Code attribute
// ----------------------------------------------------------------------------------------------

func (c ClassReader) readCode(methodVisitor MethodVisitor, contexnt *Context, codeOffset int) {
	//TODO
}

func (c ClassReader) readLabel(bytecodeOffset int, labels []*Label) *Label {
	if labels[bytecodeOffset] == nil {
		labels[bytecodeOffset] = &Label{}
	}
	return labels[bytecodeOffset]
}

func (c ClassReader) createLabel(bytecodeOffset int, labels []*Label) *Label {
	label := c.readLabel(bytecodeOffset, labels)
	label.flags &= ^FLAG_DEBUG_ONLY
	return label
}

func (c ClassReader) createDebugLabel(bytecodeOffset int, labels []*Label) {
	if labels[bytecodeOffset] == nil {
		c.readLabel(bytecodeOffset, labels).flags |= FLAG_DEBUG_ONLY
	}
}

// ----------------------------------------------------------------------------------------------
// Methods to parse annotations, type annotations and parameter annotations
// ----------------------------------------------------------------------------------------------

func (c ClassReader) readTypeAnnotations(methodVisitor MethodVisitor, context *Context, runtimeTypeAnnotationsOffset int, visible bool) []int {
	//TODO
	return nil
}

func (c ClassReader) getTypeAnnotationBytecodeOffset(typeAnnotationOffsets []int, typeAnnotationIndex int) int {
	if typeAnnotationOffsets == nil || typeAnnotationIndex >= len(typeAnnotationOffsets) || c.readByte(typeAnnotationOffsets[typeAnnotationIndex]) < INSTANCEOF {
		return -1
	}

	return c.readUnsignedShort(typeAnnotationOffsets[typeAnnotationIndex] + 1)
}

func (c ClassReader) readTypeAnnotationTarget(context *Context, typeAnnotationOffset int) int {
	//TODO
	return 0
}

func (c ClassReader) readParameterAnnotations(methodVisitor MethodVisitor, context *Context, runtimeParameterAnnotationsOffset int, visible bool) {
	//TODO
}

func (c ClassReader) readElementValues(annotationVisitor AnnotationVisitor, annotationOffset int, named bool, charBuffer []rune) int {
	//TODO
	return 0
}

func (c ClassReader) readElementValue(annotationVisitor AnnotationVisitor, elementValueOffset int, elementName string, charBuffer []rune) int {
	return 0
}

// ----------------------------------------------------------------------------------------------
// Methods to parse stack map frames
// ----------------------------------------------------------------------------------------------

func (c ClassReader) computeImplicitFame(context *Context) {
	//TODO
}

func (c ClassReader) readStackMapFrame(stackMapFrameOffset int, compressed bool, expand bool, context *Context) int {
	//TODO
	return 0
}

func (c ClassReader) readVerificationTypeInfo(verificationTypeInfoOffset int, frame []interface{}, index int, charBuffer []rune, labels []*Label) int {
	//TODO
	return 0
}

// ----------------------------------------------------------------------------------------------
// Methods to parse attributes
// ----------------------------------------------------------------------------------------------

func (c ClassReader) getFirstAttributeOffset() int {
	currentOffset := c.header + 8 + c.readUnsignedShort(c.header+6)*2
	fieldsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for fieldsCount > 0 {
		attributesCount := c.readUnsignedShort(currentOffset + 6)
		currentOffset += 8
		for attributesCount > 0 {
			currentOffset += 6 + c.readInt(currentOffset+2)
			attributesCount--
		}
		fieldsCount--
	}

	methodsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for methodsCount > 0 {
		attributesCount := c.readUnsignedShort(currentOffset + 6)
		currentOffset += 8
		for attributesCount > 0 {
			currentOffset += 6 + c.readInt(currentOffset+2)
			attributesCount--
		}
		methodsCount--
	}

	return currentOffset + 2
}

func (c ClassReader) readAttribute(attributePrototypes []Attribute, typed string, offset int, length int, charBuffer []rune, codeAttributeOffset int, labels []*Label) *Attribute {
	for i := 0; i < len(attributePrototypes); i++ {
		if attributePrototypes[i].typed == typed {
			return attributePrototypes[i].read(&c, offset, length, charBuffer, codeAttributeOffset, labels)
		}
	}
	return NewAttribute(typed).read(&c, offset, length, nil, -1, nil)
}

// -----------------------------------------------------------------------------------------------
// Utility methods: low level parsing
// -----------------------------------------------------------------------------------------------

func (c ClassReader) getItemCount() int {
	return len(c.cpInfoOffsets)
}

func (c ClassReader) getItem(constantPoolEntryIndex int) int {
	return c.cpInfoOffsets[constantPoolEntryIndex]
}

func (c ClassReader) getMaxStringLength() int {
	return c.maxStringLength
}

func (c ClassReader) readByte(offset int) byte {
	return c.b[offset] & 0xFF
}

func (c ClassReader) readUnsignedShort(offset int) int {
	b := c.b
	return int(((b[offset] & 0xFF) << 8) | (b[offset+1] & 0xFF))
}

func (c ClassReader) readShort(offset int) int16 {
	b := c.b
	return int16((((b[offset] & 0xFF) << 8) | (b[offset+1] & 0xFF)))
}

func (c ClassReader) readInt(offset int) int {
	b := c.b
	return int(((b[offset] & 0xFF) << 24) | ((b[offset+1] & 0xFF) << 16) | ((b[offset+2] & 0xFF) << 8) | (b[offset+3] & 0xFF))
}

func (c ClassReader) readLong(offset int) int64 {
	var l1 int64
	var l0 int64
	l1 = int64(c.readInt(offset))
	l0 = int64(c.readInt(offset+4) & 0xFFFFFFFF)
	return (l1 << 32) | l0
}

func (c ClassReader) readUTF8(offset int, charBuffer []rune) string {
	constantPoolEntryIndex := c.readUnsignedShort(offset)
	if offset == 0 || constantPoolEntryIndex == 0 {
		return ""
	}
	return c.readUTF(constantPoolEntryIndex, charBuffer)
}

func (c ClassReader) readUTF(constantPoolEntryIndex int, charBuffer []rune) string {
	value := c.constantUtf8Values[constantPoolEntryIndex]
	if value != "" {
		return value
	}
	cpInfoOffset := c.cpInfoOffsets[constantPoolEntryIndex]
	c.constantUtf8Values[constantPoolEntryIndex] = c.readUTFB(cpInfoOffset+2, c.readUnsignedShort(cpInfoOffset), charBuffer)

	return c.constantUtf8Values[constantPoolEntryIndex]
}

func (c ClassReader) readUTFB(utfOffset int, utfLength int, charBuffer []rune) string {
	currentOffset := utfOffset
	endOffset := currentOffset + utfLength
	strLength := 0
	b := c.b
	for currentOffset < endOffset {
		currentByte := b[currentOffset]
		currentOffset++
		if (currentByte & 0x80) == 0 {
			charBuffer[strLength] = rune(currentByte & 0x7F)
			strLength++
		} else if (currentByte & 0xE0) == 0xC0 {
			charBuffer[strLength] = rune((((currentByte & 0x1F) << 6) + (b[currentOffset] & 0x3F)))
			strLength++
			currentOffset++
		} else {
			d := ((currentByte & 0xF) << 12) + ((b[currentOffset] & 0x3F) << 6)
			currentOffset++
			charBuffer[strLength] = rune((d + (b[currentOffset] & 0x3F)))
			strLength++
		}
	}
	return string(charBuffer)
}

func (c ClassReader) readStringish(offset int, charBuffer []rune) string {
	return c.readUTF8(c.cpInfoOffsets[c.readUnsignedShort(offset)], charBuffer)
}

func (c ClassReader) readClass(offset int, charBuffer []rune) string {
	return c.readStringish(offset, charBuffer)
}

func (c ClassReader) readModuleB(offset int, charBuffer []rune) string {
	return c.readStringish(offset, charBuffer)
}

func (c ClassReader) readPackage(offset int, charBuffer []rune) string {
	return c.readStringish(offset, charBuffer)
}

func (c ClassReader) readConst(constantPoolEntryIndex int, charBuffer []rune) (interface{}, error) {
	cpInfoOffset := c.cpInfoOffsets[constantPoolEntryIndex]
	switch c.b[cpInfoOffset-1] {
	case byte(symbol.CONSTANT_INTEGER_TAG):
		return c.readInt(cpInfoOffset), nil
	case byte(symbol.CONSTANT_FLOAT_TAG):
		return float32(c.readInt(cpInfoOffset)), nil
	case byte(symbol.CONSTANT_LONG_TAG):
		return c.readLong(cpInfoOffset), nil
	case byte(symbol.CONSTANT_DOUBLE_TAG):
		return float64(c.readLong(cpInfoOffset)), nil
	case byte(symbol.CONSTANT_CLASS_TAG):
		return 0, nil //Type.getObjectType(c.readUTF8(cpInfoOffset, charBuffer))
	case byte(symbol.CONSTANT_STRING_TAG):
		return c.readUTF8(cpInfoOffset, charBuffer), nil
	case byte(symbol.CONSTANT_METHOD_TYPE_TAG):
		return 0, nil //Type.getMethodType(c.readUTF8(cpInfoOffset, charBuffer))
	case byte(symbol.CONSTANT_METHOD_HANDLE_TAG):
		referenceKind := c.readByte(cpInfoOffset)
		referenceCpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(cpInfoOffset+1)]
		nameAndTypeCpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(referenceCpInfoOffset+2)]
		owner := c.readClass(referenceCpInfoOffset, charBuffer)
		name := c.readUTF8(nameAndTypeCpInfoOffset, charBuffer)
		desc := c.readUTF8(nameAndTypeCpInfoOffset+2, charBuffer)
		itf := c.b[referenceCpInfoOffset-1] == byte(symbol.CONSTANT_INTERFACE_METHODREF_TAG)
		fmt.Println(referenceKind, owner, name, desc, itf)
		return 0, nil //new Handle(referenceKind, owner, name, desc, itf)
	default:
		return nil, errors.New("Assertion Error")
	}
}
