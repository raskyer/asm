package asm

import (
	"errors"

	"github.com/leaklessgfy/asm/asm/constants"
	"github.com/leaklessgfy/asm/asm/frame"
	"github.com/leaklessgfy/asm/asm/opcodes"
	"github.com/leaklessgfy/asm/asm/symbol"
	"github.com/leaklessgfy/asm/asm/typereference"
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
	c.AcceptB(classVisitor, make([]*Attribute, 0), parsingOptions)
}

// AcceptB Makes the given visitor visit the JVMS ClassFile structure passed to the constructor of this {@link ClassReader}.
func (c ClassReader) AcceptB(classVisitor ClassVisitor, attributePrototypes []*Attribute, parsingOptions int) {
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
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitAnnotation(annotationDescriptor, true), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeInvisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitAnnotation(annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeVisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitAnnotation(annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeInvisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleTypeAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleTypeAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			currentAnnotationOffset = c.readTypeAnnotationTarget(context, currentAnnotationOffset)
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(classVisitor.VisitTypeAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
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
			numberOfClasses--
			classVisitor.VisitInnerClass(c.readClass(currentClassesOffset, charBuffer), c.readClass(currentAttributeOffset+2, charBuffer), c.readClass(currentClassesOffset+4, charBuffer), c.readUnsignedShort(currentClassesOffset+6))
			currentClassesOffset += 8
		}
	}

	fieldsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for fieldsCount > 0 {
		fieldsCount--
		currentOffset = c.readField(classVisitor, context, currentOffset)
	}
	methodsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for methodsCount > 0 {
		methodsCount--
		currentOffset = c.readMethod(classVisitor, context, currentOffset)
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
			packageCount--
			moduleVisitor.VisitPackage(c.readPackage(currentPackageOffset, buffer))
			currentPackageOffset += 2
		}
	}

	requiresCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for requiresCount > 0 {
		requiresCount--
		requires := c.readModuleB(currentOffset, buffer)
		requiresFlags := c.readUnsignedShort(currentOffset + 2)
		requiresVersion := c.readUTF8(currentOffset+4, buffer)
		currentOffset += 6
		moduleVisitor.VisitRequire(requires, requiresFlags, requiresVersion)
	}

	exportsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for exportsCount > 0 {
		exportsCount--
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
	}

	opensCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for opensCount > 0 {
		opensCount--
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
		usesCount--
		moduleVisitor.VisitUse(c.readClass(currentOffset, buffer))
		currentOffset += 2
	}

	providesCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for providesCount > 0 {
		providesCount--
		provides := c.readClass(currentOffset, buffer)
		providesWithCount := c.readUnsignedShort(currentOffset + 2)
		currentOffset += 4
		providesWith := make([]string, providesWithCount)
		for i := 0; i < providesWithCount; i++ {
			providesWith[i] = c.readClass(currentOffset, buffer)
			currentOffset += 2
		}
		moduleVisitor.VisitProvide(provides, providesWith...)
	}

	moduleVisitor.VisitEnd()
}

func (c ClassReader) readField(classVisitor ClassVisitor, context *Context, fieldInfoOffset int) int {
	charBuffer := context.charBuffer
	currentOffset := fieldInfoOffset
	accessFlags := c.readUnsignedShort(currentOffset)
	name := c.readUTF8(currentOffset+2, charBuffer)
	descriptor := c.readUTF8(currentOffset+4, charBuffer)
	currentOffset += 6

	var constantValue interface{}
	var signature string

	runtimeVisibleAnnotationsOffset := 0
	runtimeInvisibleAnnotationsOffset := 0
	runtimeVisibleTypeAnnotationsOffset := 0
	runtimeInvisibleTypeAnnotationsOffset := 0
	var attributes *Attribute

	attributesCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2

	for attributesCount > 0 {
		attributesCount--
		attributeName := c.readUTF8(currentOffset, charBuffer)
		attributeLength := c.readInt(currentOffset + 2)
		currentOffset += 6

		switch attributeName {
		case "ConstantValue":
			constantvalueIndex := c.readUnsignedShort(currentOffset)
			if constantvalueIndex != 0 {
				constantValue, _ = c.readConst(constantvalueIndex, charBuffer)
			}
			break
		case "Signature":
			signature = c.readUTF8(currentOffset, charBuffer)
			break
		case "Deprecated":
			accessFlags |= opcodes.ACC_DEPRECATED
			break
		case "Synthetic":
			accessFlags |= opcodes.ACC_SYNTHETIC
			break
		case "RuntimeVisibleAnnotations":
			runtimeVisibleAnnotationsOffset = currentOffset
			break
		case "RuntimeVisibleTypeAnnotations":
			runtimeVisibleTypeAnnotationsOffset = currentOffset
			break
		case "RuntimeInvisibleAnnotations":
			runtimeInvisibleAnnotationsOffset = currentOffset
			break
		case "RuntimeInvisibleTypeAnnotations":
			runtimeInvisibleTypeAnnotationsOffset = currentOffset
			break
		default:
			attribute := c.readAttribute(context.attributePrototypes, attributeName, currentOffset, attributeLength, charBuffer, -1, nil)
			attribute.nextAttribute = attributes
			attributes = attribute
			break
		}
		currentOffset += attributeLength
	}

	fieldVisitor := classVisitor.VisitField(accessFlags, name, descriptor, signature, constantValue)
	if fieldVisitor == nil {
		return currentOffset
	}

	if runtimeVisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeVisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeVisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(fieldVisitor.VisitAnnotation(annotationDescriptor, true), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeInvisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(fieldVisitor.VisitAnnotation(annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeVisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeVisibleTypeAnnotationsOffset)
		currentAnnotationOffset := runtimeVisibleTypeAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			currentAnnotationOffset = c.readTypeAnnotationTarget(context, currentAnnotationOffset)
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			annotationVisitor := fieldVisitor.VisitTypeAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, true)
			currentAnnotationOffset = c.readElementValues(annotationVisitor, currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeInvisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleTypeAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleTypeAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			currentAnnotationOffset = c.readTypeAnnotationTarget(context, currentAnnotationOffset)
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			annotationVisitor := fieldVisitor.VisitTypeAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, false)
			currentAnnotationOffset = c.readElementValues(annotationVisitor, currentAnnotationOffset, true, charBuffer)
		}
	}

	for attributes != nil {
		nextAttribute := attributes.nextAttribute
		attributes.nextAttribute = nil
		fieldVisitor.VisitAttribute(attributes)
		attributes = nextAttribute
	}

	fieldVisitor.VisitEnd()
	return currentOffset
}

func (c ClassReader) readMethod(classVisitor ClassVisitor, context *Context, methodInfoOffset int) int {
	charBuffer := context.charBuffer
	currentOffset := methodInfoOffset
	context.currentMethodAccessFlags = c.readUnsignedShort(currentOffset)
	context.currentMethodName = c.readUTF8(currentOffset+2, charBuffer)
	context.currentMethodDescriptor = c.readUTF8(currentOffset+4, charBuffer)
	currentOffset += 6

	codeOffset := 0
	exceptionsOffset := 0
	var exceptions []string
	signature := 0
	runtimeVisibleAnnotationsOffset := 0
	runtimeInvisibleAnnotationsOffset := 0
	runtimeVisibleParameterAnnotationsOffset := 0
	runtimeInvisibleParameterAnnotationsOffset := 0
	runtimeVisibleTypeAnnotationsOffset := 0
	runtimeInvisibleTypeAnnotationsOffset := 0
	annotationDefaultOffset := 0
	methodParametersOffset := 0
	var attributes *Attribute

	attributesCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for attributesCount > 0 {
		attributesCount--
		attributeName := c.readUTF8(currentOffset, charBuffer)
		attributeLength := c.readInt(currentOffset + 2)
		currentOffset += 6

		switch attributeName {
		case "Code":
			if (context.parsingOptions & SKIP_CODE) == 0 {
				codeOffset = currentOffset
			}
			break
		case "Exceptions":
			exceptionsOffset = currentOffset
			exceptions = make([]string, c.readUnsignedShort(exceptionsOffset))
			currentExceptionOffset := exceptionsOffset + 2
			for i := 0; i < len(exceptions); i++ {
				exceptions[i] = c.readClass(currentExceptionOffset, charBuffer)
				currentExceptionOffset += 2
			}
			break
		case "Signature":
			signature = c.readUnsignedShort(currentOffset)
			break
		case "Deprecated":
			context.currentMethodAccessFlags |= opcodes.ACC_DEPRECATED
			break
		case "RuntimeVisibleAnnotations":
			runtimeVisibleAnnotationsOffset = currentOffset
			break
		case "RuntimeVisibleTypeAnnotations":
			runtimeVisibleTypeAnnotationsOffset = currentOffset
			break
		case "AnnotationDefault":
			annotationDefaultOffset = currentOffset
			break
		case "Synthetic":
			context.currentMethodAccessFlags |= opcodes.ACC_SYNTHETIC
			break
		case "RuntimeInvisibleAnnotations":
			runtimeInvisibleAnnotationsOffset = currentOffset
			break
		case "RuntimeInvisibleTypeAnnotations":
			runtimeInvisibleTypeAnnotationsOffset = currentOffset
			break
		case "RuntimeVisibleParameterAnnotations":
			runtimeVisibleParameterAnnotationsOffset = currentOffset
			break
		case "RuntimeInvisibleParameterAnnotations":
			runtimeInvisibleParameterAnnotationsOffset = currentOffset
			break
		case "MethodParameters":
			methodParametersOffset = currentOffset
			break
		default:
			attribute := c.readAttribute(context.attributePrototypes, attributeName, currentOffset, attributeLength, charBuffer, -1, nil)
			attribute.nextAttribute = attributes
			attributes = attribute
			break
		}
		currentOffset += attributeLength
	}

	var sig string
	if signature != 0 {
		sig = c.readUTF(signature, charBuffer)
	}
	methodVisitor := classVisitor.VisitMethod(context.currentMethodAccessFlags, context.currentMethodName, context.currentMethodDescriptor, sig, exceptions)
	if methodVisitor == nil {
		return currentOffset
	}

	/* MethodWriter instanceof ? */

	if methodParametersOffset != 0 {
		parametersCount := c.readByte(methodParametersOffset)
		currentParameterOffset := methodParametersOffset + 1
		for parametersCount > 0 {
			parametersCount--
			methodVisitor.VisitParameter(c.readUTF8(currentParameterOffset, charBuffer), c.readUnsignedShort(currentParameterOffset+2))
			currentParameterOffset += 4
		}
	}

	if annotationDefaultOffset != 0 {
		annotationVisitor := methodVisitor.VisitAnnotationDefault()
		c.readElementValue(annotationVisitor, annotationDefaultOffset, "", charBuffer)
		if annotationVisitor != nil {
			annotationVisitor.VisitEnd()
		}
	}

	if runtimeVisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeVisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeVisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(methodVisitor.VisitAnnotation(annotationDescriptor, true), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeInvisibleAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			currentAnnotationOffset = c.readElementValues(methodVisitor.VisitAnnotation(annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeVisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeVisibleTypeAnnotationsOffset)
		currentAnnotationOffset := runtimeVisibleTypeAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			currentAnnotationOffset = c.readTypeAnnotationTarget(context, currentAnnotationOffset)
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			annotationVisitor := methodVisitor.VisitTypeAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, true)
			currentAnnotationOffset = c.readElementValues(annotationVisitor, currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeInvisibleTypeAnnotationsOffset != 0 {
		numAnnotations := c.readUnsignedShort(runtimeInvisibleTypeAnnotationsOffset)
		currentAnnotationOffset := runtimeInvisibleTypeAnnotationsOffset + 2
		for numAnnotations > 0 {
			numAnnotations--
			currentAnnotationOffset = c.readTypeAnnotationTarget(context, currentAnnotationOffset)
			annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
			currentAnnotationOffset += 2
			annotationVisitor := methodVisitor.VisitTypeAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, false)
			currentAnnotationOffset = c.readElementValues(annotationVisitor, currentAnnotationOffset, true, charBuffer)
		}
	}

	if runtimeVisibleParameterAnnotationsOffset != 0 {
		c.readParameterAnnotations(methodVisitor, context, runtimeVisibleParameterAnnotationsOffset, true)
	}

	if runtimeInvisibleParameterAnnotationsOffset != 0 {
		c.readParameterAnnotations(methodVisitor, context, runtimeInvisibleParameterAnnotationsOffset, false)
	}

	for attributes != nil {
		nextAttribute := attributes.nextAttribute
		attributes.nextAttribute = nil
		methodVisitor.VisitAttribute(attributes)
		attributes = nextAttribute
	}

	if codeOffset != 0 {
		methodVisitor.VisitCode()
		c.readCode(methodVisitor, context, codeOffset)
	}

	methodVisitor.VisitEnd()
	return currentOffset
}

// ----------------------------------------------------------------------------------------------
// Methods to parse a Code attribute
// ----------------------------------------------------------------------------------------------

func (c ClassReader) readCode(methodVisitor MethodVisitor, context *Context, codeOffset int) {
	currentOffset := codeOffset
	b := c.b
	charBuffer := context.charBuffer
	maxStack := c.readUnsignedShort(currentOffset)
	maxLocals := c.readUnsignedShort(currentOffset + 2)
	codeLength := c.readInt(currentOffset + 4)
	currentOffset += 8

	bytecodeStartOffset := currentOffset
	bytecodeEndOffset := currentOffset + codeLength
	context.currentMethodLabels = make([]*Label, codeLength+1)
	labels := context.currentMethodLabels

	for currentOffset < bytecodeEndOffset {
		bytecodeOffset := currentOffset - bytecodeStartOffset
		opcode := b[currentOffset] & 0xFF
		switch opcode {
		case constants.NOP, constants.ACONST_NULL, constants.ICONST_M1, constants.ICONST_0, constants.ICONST_1, constants.ICONST_2,
			constants.ICONST_3, constants.ICONST_4, constants.ICONST_5, constants.LCONST_0, constants.LCONST_1, constants.FCONST_0, constants.FCONST_1,
			constants.FCONST_2, constants.DCONST_0, constants.DCONST_1, constants.IALOAD, constants.LALOAD, constants.FALOAD, constants.DALOAD,
			constants.AALOAD, constants.BALOAD, constants.CALOAD, constants.SALOAD, constants.IASTORE, constants.LASTORE, constants.FASTORE, constants.DASTORE,
			constants.AASTORE, constants.BASTORE, constants.CASTORE, constants.SASTORE, constants.POP, constants.POP2, constants.DUP, constants.DUP_X1, constants.DUP_X2,
			constants.DUP2, constants.DUP2_X1, constants.DUP2_X2, constants.SWAP, constants.IADD, constants.LADD, constants.FADD, constants.DADD, constants.ISUB,
			constants.LSUB, constants.FSUB, constants.DSUB, constants.IMUL, constants.LMUL, constants.FMUL, constants.DMUL, constants.IDIV, constants.LDIV, constants.FDIV,
			constants.DDIV, constants.IREM, constants.LREM, constants.FREM, constants.DREM, constants.INEG, constants.LNEG, constants.FNEG, constants.DNEG, constants.ISHL,
			constants.LSHL, constants.ISHR, constants.LSHR, constants.IUSHR, constants.LUSHR, constants.IAND, constants.LAND, constants.IOR, constants.LOR, constants.IXOR,
			constants.LXOR, constants.I2L, constants.I2F, constants.I2D, constants.L2I, constants.L2F, constants.L2D, constants.F2I, constants.F2L, constants.F2D,
			constants.D2I, constants.D2L, constants.D2F, constants.I2B, constants.I2C, constants.I2S, constants.LCMP, constants.FCMPL, constants.FCMPG, constants.DCMPL,
			constants.DCMPG, constants.IRETURN, constants.LRETURN, constants.FRETURN, constants.DRETURN, constants.ARETURN, constants.RETURN, constants.ARRAYLENGTH,
			constants.ATHROW, constants.MONITORENTER, constants.MONITOREXIT, constants.ILOAD_0, constants.ILOAD_1, constants.ILOAD_2, constants.ILOAD_3, constants.LLOAD_0,
			constants.LLOAD_1, constants.LLOAD_2, constants.LLOAD_3, constants.FLOAD_0, constants.FLOAD_1, constants.FLOAD_2, constants.FLOAD_3, constants.DLOAD_0,
			constants.DLOAD_1, constants.DLOAD_2, constants.DLOAD_3, constants.ALOAD_0, constants.ALOAD_1, constants.ALOAD_2, constants.ALOAD_3, constants.ISTORE_0,
			constants.ISTORE_1, constants.ISTORE_2, constants.ISTORE_3, constants.LSTORE_0, constants.LSTORE_1, constants.LSTORE_2, constants.LSTORE_3, constants.FSTORE_0,
			constants.FSTORE_1, constants.FSTORE_2, constants.FSTORE_3, constants.DSTORE_0, constants.DSTORE_1, constants.DSTORE_2, constants.DSTORE_3, constants.ASTORE_0,
			constants.ASTORE_1, constants.ASTORE_2, constants.ASTORE_3:
			currentOffset++
			break
		case constants.IFEQ, constants.IFNE, constants.IFLT, constants.IFGE, constants.IFGT, constants.IFLE, constants.IF_ICMPEQ, constants.IF_ICMPNE, constants.IF_ICMPLT,
			constants.IF_ICMPGE, constants.IF_ICMPGT, constants.IF_ICMPLE, constants.IF_ACMPEQ, constants.IF_ACMPNE, constants.GOTO, constants.JSR, constants.IFNULL,
			constants.IFNONNULL:
			c.createLabel(bytecodeOffset+int(c.readShort(currentOffset+1)), labels)
			currentOffset += 3
			break
		case constants.ASM_IFEQ, constants.ASM_IFNE, constants.ASM_IFLT, constants.ASM_IFGE, constants.ASM_IFGT, constants.ASM_IFLE, constants.ASM_IF_ICMPEQ,
			constants.ASM_IF_ICMPNE, constants.ASM_IF_ICMPLT, constants.ASM_IF_ICMPGE, constants.ASM_IF_ICMPGT, constants.ASM_IF_ICMPLE, constants.ASM_IF_ACMPEQ,
			constants.ASM_IF_ACMPNE, constants.ASM_GOTO, constants.ASM_JSR, constants.ASM_IFNULL, constants.ASM_IFNONNULL:
			c.createLabel(bytecodeOffset+c.readUnsignedShort(currentOffset+1), labels)
			currentOffset += 3
			break
		case constants.GOTO_W, constants.JSR_W, constants.ASM_GOTO_W:
			c.createLabel(bytecodeOffset+c.readInt(currentOffset+1), labels)
			currentOffset += 5
			break
		case constants.WIDE:
			if (b[currentOffset+1] & 0xFF) == opcodes.IINC {
				currentOffset += 6
			} else {
				currentOffset += 4
			}
			break
		case constants.TABLESWITCH:
			currentOffset += 4 - (bytecodeOffset & 3)
			c.createLabel(bytecodeOffset+c.readInt(currentOffset), labels)
			numTableEntries := c.readInt(currentOffset+8) - c.readInt(currentOffset+4) + 1
			currentOffset += 12
			for numTableEntries > 0 {
				numTableEntries--
				c.createLabel(bytecodeOffset+c.readInt(currentOffset), labels)
				currentOffset += 4
			}
			break
		case constants.LOOKUPSWITCH:
			currentOffset += 4 - (bytecodeOffset & 3)
			c.createLabel(bytecodeOffset+c.readInt(currentOffset), labels)
			numSwitchCases := c.readInt(currentOffset + 4)
			currentOffset += 8
			for numSwitchCases > 0 {
				numSwitchCases--
				c.createLabel(bytecodeOffset+c.readInt(currentOffset+4), labels)
				currentOffset += 8
			}
			break
		case constants.ILOAD, constants.LLOAD, constants.FLOAD, constants.DLOAD, constants.ALOAD, constants.ISTORE,
			constants.LSTORE, constants.FSTORE, constants.DSTORE, constants.ASTORE, constants.RET, constants.BIPUSH, constants.NEWARRAY, constants.LDC:
			currentOffset += 2
			break
		case constants.SIPUSH, constants.LDC_W, constants.LDC2_W, constants.GETSTATIC, constants.PUTSTATIC, constants.GETFIELD, constants.PUTFIELD,
			constants.INVOKEVIRTUAL, constants.INVOKESPECIAL, constants.INVOKESTATIC, constants.NEW, constants.ANEWARRAY, constants.CHECKCAST, constants.INSTANCEOF,
			constants.IINC:
			currentOffset += 3
			break
		case constants.INVOKEINTERFACE, constants.INVOKEDYNAMIC:
			currentOffset += 5
			break
		case constants.MULTIANEWARRAY:
			currentOffset += 4
			break
		default:
			//throw error
			panic(errors.New("AssertionError"))
			break
		}
	}

	{
		exceptionTableLength := c.readUnsignedShort(currentOffset)
		currentOffset += 2
		for exceptionTableLength > 0 {
			exceptionTableLength--
			start := c.createLabel(c.readUnsignedShort(currentOffset), labels)
			end := c.createLabel(c.readUnsignedShort(currentOffset+2), labels)
			handler := c.createLabel(c.readUnsignedShort(currentOffset+4), labels)
			catchType := c.readUTF8(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+6)], charBuffer)
			currentOffset += 8
			methodVisitor.VisitTryCatchBlock(start, end, handler, catchType)
		}
	}

	stackMapFrameOffset := 0
	stackMapTableEndOffset := 0
	compressedFrames := true
	localVariableTableOffset := 0
	localVariableTypeTableOffset := 0
	var visibleTypeAnnotationOffsets []int
	var invisibleTypeAnnotationOffsets []int
	var attributes *Attribute

	attributesCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for attributesCount > 0 {
		attributesCount--
		attributeName := c.readUTF8(currentOffset, charBuffer)
		attributeLength := c.readInt(currentOffset + 2)
		currentOffset += 6

		switch attributeName {
		case "LocalVariableTable":
			if (context.parsingOptions & SKIP_DEBUG) == 0 {
				localVariableTableOffset = currentOffset
				localVariableTableLength := c.readUnsignedShort(currentOffset)
				currentOffset += 2
				for localVariableTableLength > 0 {
					localVariableTableLength--
					startPc := c.readUnsignedShort(currentOffset)
					c.createDebugLabel(startPc, labels)
					length := c.readUnsignedShort(currentOffset + 2)
					c.createDebugLabel(startPc+length, labels)
					currentOffset += 10
				}
				continue
			}
			break
		case "LocalVariableTypeTable":
			localVariableTypeTableOffset = currentOffset
			break
		case "LineNumberTable":
			if (context.parsingOptions & SKIP_DEBUG) == 0 {
				lineNumberTableLength := c.readUnsignedShort(currentOffset)
				currentOffset += 2
				for lineNumberTableLength > 0 {
					lineNumberTableLength--
					startPc := c.readUnsignedShort(currentOffset)
					lineNumber := c.readUnsignedShort(currentOffset + 2)
					currentOffset += 4
					c.createDebugLabel(startPc, labels)
					labels[startPc].addLineNumber(lineNumber)
				}
				continue
			}
			break
		case "RuntimeVisibleTypeAnnotations":
			visibleTypeAnnotationOffsets = c.readTypeAnnotations(methodVisitor, context, currentOffset, true)
			break
		case "RuntimeInvisibleTypeAnnotations":
			invisibleTypeAnnotationOffsets = c.readTypeAnnotations(methodVisitor, context, currentOffset, false)
			break
		case "StackMapTable":
			if (context.parsingOptions & SKIP_FRAMES) == 0 {
				stackMapFrameOffset = currentOffset + 2
				stackMapTableEndOffset = currentOffset + attributeLength
			}
			break
		case "StackMap":
			if (context.parsingOptions & SKIP_FRAMES) == 0 {
				stackMapFrameOffset = currentOffset + 2
				stackMapTableEndOffset = currentOffset + attributeLength
				compressedFrames = false
			}
			break
		default:
			attribute := c.readAttribute(context.attributePrototypes, attributeName, currentOffset, attributeLength, charBuffer, codeOffset, labels)
			attribute.nextAttribute = attributes
			attributes = attribute
			break
		}
		currentOffset += attributeLength
	}

	expandFrames := (context.parsingOptions & EXPAND_FRAMS) != 0
	if stackMapFrameOffset != 0 {
		context.currentFrameOffset = -1
		context.currentFrameType = 0
		context.currentFrameLocalCount = 0
		context.currentFrameLocalCountDelta = 0
		context.currentFrameLocalTypes = make([]interface{}, maxLocals)
		context.currentFrameStackCount = 0
		context.currentFrameStackTypes = make([]interface{}, maxStack)
		if expandFrames {
			c.computeImplicitFame(context)
		}
		for offset := stackMapFrameOffset; offset < stackMapTableEndOffset-2; offset++ {
			if b[offset] == frame.ITEM_UNINITIALIZED {
				potentialBytecodeOffset := c.readUnsignedShort(offset + 1)
				if potentialBytecodeOffset >= 0 && potentialBytecodeOffset < codeLength {
					if (b[bytecodeStartOffset+potentialBytecodeOffset] & 0xFF) == opcodes.NEW {
						c.createLabel(potentialBytecodeOffset, labels)
					}
				}
			}
		}
	}
	if expandFrames && (context.parsingOptions&EXPAND_ASM_INSNS) != 0 {
		methodVisitor.VisitFrame(opcodes.F_NEW, maxLocals, nil, 0, nil)
	}

	currentVisibleTypeAnnotationIndex := 0
	currentVisibleTypeAnnotationBytecodeOffset := c.getTypeAnnotationBytecodeOffset(visibleTypeAnnotationOffsets, 0)
	currentInvisibleTypeAnnotationIndex := 0
	currentInvisibleTypeAnnotationBytecodeOffset := c.getTypeAnnotationBytecodeOffset(invisibleTypeAnnotationOffsets, 0)
	insertFrame := false

	wideJumpOpcodeDelta := 0
	if (context.parsingOptions & EXPAND_ASM_INSNS) == 0 {
		wideJumpOpcodeDelta = constants.WIDE_JUMP_OPCODE_DELTA
	}
	currentOffset = bytecodeStartOffset

	for currentOffset < bytecodeEndOffset {
		currentBytecodeOffset := currentOffset - bytecodeStartOffset
		currentLabel := labels[currentBytecodeOffset]
		if currentLabel != nil {
			currentLabel.accept(methodVisitor, (context.parsingOptions&SKIP_DEBUG) == 0)
		}

		for stackMapFrameOffset != 0 && (context.currentFrameOffset == currentBytecodeOffset || context.currentFrameOffset == -1) {
			if context.currentFrameOffset != -1 {
				if !compressedFrames || expandFrames {
					methodVisitor.VisitFrame(opcodes.F_NEW, context.currentFrameLocalCount, context.currentFrameLocalTypes, context.currentFrameStackCount, context.currentFrameStackTypes)
				} else {
					methodVisitor.VisitFrame(context.currentFrameType, context.currentFrameLocalCountDelta, context.currentFrameLocalTypes, context.currentFrameStackCount, context.currentFrameStackTypes)
				}
				insertFrame = false
			}
			if stackMapFrameOffset < stackMapTableEndOffset {
				stackMapFrameOffset = c.readStackMapFrame(stackMapFrameOffset, compressedFrames, expandFrames, context)
			} else {
				stackMapFrameOffset = 0
			}
		}

		if insertFrame {
			if context.parsingOptions&EXPAND_FRAMS != 0 {
				methodVisitor.VisitFrame(constants.F_INSERT, 0, nil, 0, nil)
			}
			insertFrame = false
		}

		opcode := b[currentOffset] & 0xFF
		switch opcode {
		case constants.NOP, constants.ACONST_NULL, constants.ICONST_M1,
			constants.ICONST_0, constants.ICONST_1, constants.ICONST_2, constants.ICONST_3, constants.ICONST_4, constants.ICONST_5,
			constants.LCONST_0, constants.LCONST_1,
			constants.FCONST_0, constants.FCONST_1, constants.FCONST_2,
			constants.DCONST_0, constants.DCONST_1,
			constants.IALOAD, constants.LALOAD, constants.FALOAD, constants.DALOAD, constants.AALOAD, constants.BALOAD, constants.CALOAD, constants.SALOAD,
			constants.IASTORE, constants.LASTORE, constants.FASTORE, constants.DASTORE, constants.AASTORE, constants.BASTORE, constants.CASTORE, constants.SASTORE,
			constants.POP, constants.POP2,
			constants.DUP, constants.DUP_X1, constants.DUP_X2, constants.DUP2, constants.DUP2_X1, constants.DUP2_X2,
			constants.SWAP, constants.IADD, constants.LADD, constants.FADD, constants.DADD,
			constants.ISUB, constants.LSUB, constants.FSUB, constants.DSUB,
			constants.IMUL, constants.LMUL, constants.FMUL, constants.DMUL,
			constants.IDIV, constants.LDIV, constants.FDIV, constants.DDIV,
			constants.IREM, constants.LREM, constants.FREM, constants.DREM,
			constants.INEG, constants.LNEG, constants.FNEG, constants.DNEG,
			constants.ISHL, constants.LSHL, constants.ISHR, constants.LSHR, constants.IUSHR, constants.LUSHR,
			constants.IAND, constants.LAND, constants.IOR, constants.LOR, constants.IXOR, constants.LXOR,
			constants.I2L, constants.I2F, constants.I2D, constants.L2I, constants.L2F, constants.L2D,
			constants.F2I, constants.F2L, constants.F2D,
			constants.D2I, constants.D2L, constants.D2F,
			constants.I2B, constants.I2C, constants.I2S,
			constants.LCMP, constants.FCMPL, constants.FCMPG, constants.DCMPL, constants.DCMPG,
			constants.IRETURN, constants.LRETURN, constants.FRETURN, constants.DRETURN, constants.ARETURN, constants.RETURN,
			constants.ARRAYLENGTH, constants.ATHROW,
			constants.MONITORENTER, constants.MONITOREXIT:
			methodVisitor.VisitInsn(int(opcode))
			currentOffset++
			break
		case constants.ILOAD_0, constants.ILOAD_1, constants.ILOAD_2, constants.ILOAD_3,
			constants.LLOAD_0, constants.LLOAD_1, constants.LLOAD_2, constants.LLOAD_3,
			constants.FLOAD_0, constants.FLOAD_1, constants.FLOAD_2, constants.FLOAD_3,
			constants.DLOAD_0, constants.DLOAD_1, constants.DLOAD_2, constants.DLOAD_3,
			constants.ALOAD_0, constants.ALOAD_1, constants.ALOAD_2, constants.ALOAD_3:
			opcode -= constants.ILOAD_0
			methodVisitor.VisitVarInsn(int(opcodes.ILOAD+(opcode>>2)), int(opcode&0x3))
			currentOffset++
			break
		case constants.ISTORE_0, constants.ISTORE_1, constants.ISTORE_2, constants.ISTORE_3,
			constants.LSTORE_0, constants.LSTORE_1, constants.LSTORE_2, constants.LSTORE_3,
			constants.FSTORE_0, constants.FSTORE_1, constants.FSTORE_2, constants.FSTORE_3,
			constants.DSTORE_0, constants.DSTORE_1, constants.DSTORE_2, constants.DSTORE_3,
			constants.ASTORE_0, constants.ASTORE_1, constants.ASTORE_2, constants.ASTORE_3:
			opcode -= constants.ISTORE_0
			methodVisitor.VisitVarInsn(int(opcodes.ISTORE+(opcode>>2)), int(opcode&0x3))
			currentOffset++
			break
		case constants.IFEQ, constants.IFNE, constants.IFLT, constants.IFGE, constants.IFGT, constants.IFLE,
			constants.IF_ICMPEQ, constants.IF_ICMPNE, constants.IF_ICMPLT, constants.IF_ICMPGE, constants.IF_ICMPGT, constants.IF_ICMPLE,
			constants.IF_ACMPEQ, constants.IF_ACMPNE, constants.GOTO, constants.JSR, constants.IFNULL, constants.IFNONNULL:
			methodVisitor.VisitJumpInsn(int(opcode), labels[currentBytecodeOffset+int(c.readShort(currentOffset+1))])
			currentOffset += 3
			break
		case constants.GOTO_W, constants.JSR_W:
			methodVisitor.VisitJumpInsn(int(opcode)-wideJumpOpcodeDelta, labels[currentBytecodeOffset+c.readInt(currentOffset+1)])
			currentOffset += 5
			break
		case constants.ASM_IFEQ, constants.ASM_IFNE, constants.ASM_IFLT, constants.ASM_IFGE, constants.ASM_IFGT, constants.ASM_IFLE,
			constants.ASM_IF_ICMPEQ, constants.ASM_IF_ICMPNE, constants.ASM_IF_ICMPLT, constants.ASM_IF_ICMPGE, constants.ASM_IF_ICMPGT, constants.ASM_IF_ICMPLE,
			constants.ASM_IF_ACMPEQ, constants.ASM_IF_ACMPNE,
			constants.ASM_GOTO, constants.ASM_JSR, constants.ASM_IFNULL, constants.ASM_IFNONNULL:
			{
				if opcode < constants.ASM_IFNULL {
					opcode = opcode - constants.ASM_OPCODE_DELTA
				} else {
					opcode = opcode - constants.ASM_IFNULL_OPCODE_DELTA
				}
				target := labels[currentBytecodeOffset+c.readUnsignedShort(currentOffset+1)]
				if opcode == opcodes.GOTO || opcode == opcodes.JSR {
					methodVisitor.VisitJumpInsn(int(opcode+constants.WIDE_JUMP_OPCODE_DELTA), target)
				} else {
					if opcode < opcodes.GOTO {
						opcode = ((opcode + 1) ^ 1) - 1
					} else {
						opcode = opcode ^ 1
					}
					endif := c.createLabel(currentBytecodeOffset+3, labels)
					methodVisitor.VisitJumpInsn(int(opcode), endif)
					methodVisitor.VisitJumpInsn(constants.GOTO_W, target)
					insertFrame = true
				}
				currentOffset += 3
				break
			}
		case constants.ASM_GOTO_W:
			{
				methodVisitor.VisitJumpInsn(constants.GOTO_W, labels[currentBytecodeOffset+c.readInt(currentOffset+1)])
				insertFrame = true
				currentOffset += 5
				break
			}
		case constants.WIDE:
			opcode = b[currentOffset+1] & 0xFF
			if opcode == opcodes.IINC {
				methodVisitor.VisitIincInsn(c.readUnsignedShort(currentOffset+2), int(c.readShort(currentOffset+4)))
				currentOffset += 6
			} else {
				methodVisitor.VisitVarInsn(int(opcode), c.readUnsignedShort(currentOffset+2))
				currentOffset += 4
			}
			break
		case constants.TABLESWITCH:
			{
				currentOffset += 4 - (currentBytecodeOffset & 3)
				defaultLabel := labels[currentBytecodeOffset+c.readInt(currentOffset)]
				low := c.readInt(currentOffset + 4)
				high := c.readInt(currentOffset + 8)
				currentOffset += 12
				table := make([]*Label, high-low+1)
				for i := 0; i < len(table); i++ {
					table[i] = labels[currentBytecodeOffset+c.readInt(currentOffset)]
					currentOffset += 4
				}
				methodVisitor.VisitTableSwitchInsn(low, high, defaultLabel, table...)
				break
			}
		case constants.LOOKUPSWITCH:
			{
				currentOffset += 4 - (currentBytecodeOffset & 3)
				defaultLabel := labels[currentBytecodeOffset+c.readInt(currentOffset)]
				nPairs := c.readInt(currentOffset + 4)
				currentOffset += 8
				keys := make([]int, nPairs)
				values := make([]*Label, nPairs)
				for i := 0; i < nPairs; i++ {
					keys[i] = c.readInt(currentOffset)
					values[i] = labels[currentBytecodeOffset+c.readInt(currentOffset+4)]
					currentOffset += 8
				}
				methodVisitor.VisitLookupSwitchInsn(defaultLabel, keys, values)
				break
			}
		case constants.ILOAD, constants.LLOAD, constants.FLOAD, constants.DLOAD, constants.ALOAD,
			constants.ISTORE, constants.LSTORE, constants.FSTORE, constants.DSTORE, constants.ASTORE,
			constants.RET:
			methodVisitor.VisitVarInsn(int(opcode), int(b[currentOffset+1]&0xFF))
			currentOffset += 2
			break
		case constants.BIPUSH, constants.NEWARRAY:
			methodVisitor.VisitIntInsn(int(opcode), int(b[currentOffset+1]))
			currentOffset += 2
			break
		case constants.SIPUSH:
			methodVisitor.VisitIntInsn(int(opcode), int(c.readShort(currentOffset+1)))
			currentOffset += 3
			break
		case constants.LDC:
			constd, _ := c.readConst(int(b[currentOffset+1]&0xFF), charBuffer)
			methodVisitor.VisitLdcInsn(constd)
			currentOffset += 2
			break
		case constants.LDC_W, constants.LDC2_W:
			constd, _ := c.readConst(c.readUnsignedShort(currentOffset+1), charBuffer)
			methodVisitor.VisitLdcInsn(constd)
			currentOffset += 3
			break
		case constants.GETSTATIC, constants.PUTSTATIC, constants.GETFIELD, constants.PUTFIELD,
			constants.INVOKEVIRTUAL, constants.INVOKESPECIAL, constants.INVOKESTATIC, constants.INVOKEINTERFACE:
			{
				cpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]
				nameAndTypeCpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(cpInfoOffset+2)]
				owner := c.readClass(cpInfoOffset, charBuffer)
				name := c.readUTF8(nameAndTypeCpInfoOffset, charBuffer)
				desc := c.readUTF8(nameAndTypeCpInfoOffset+2, charBuffer)
				if opcode < opcodes.INVOKEVIRTUAL {
					methodVisitor.VisitFieldInsn(int(opcode), owner, name, desc)
				} else {
					itf := b[cpInfoOffset-1] == symbol.CONSTANT_INTERFACE_METHODREF_TAG
					methodVisitor.VisitMethodInsnB(int(opcode), owner, name, desc, itf)
				}
				if opcode == opcodes.INVOKEINTERFACE {
					currentOffset += 5
				} else {
					currentOffset += 3
				}
				break
			}
		case constants.INVOKEDYNAMIC:
			{
				cpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]
				nameAndTypeCpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(cpInfoOffset+2)]
				name := c.readUTF8(nameAndTypeCpInfoOffset, charBuffer)
				desc := c.readUTF8(nameAndTypeCpInfoOffset+2, charBuffer)
				bootstrapMethodOffset := context.bootstrapMethodOffsets[c.readUnsignedShort(cpInfoOffset)]
				handle, _ := c.readConst(c.readUnsignedShort(bootstrapMethodOffset), charBuffer)
				bootstrapMethodArguments := make([]interface{}, c.readUnsignedShort(bootstrapMethodOffset+2))
				bootstrapMethodOffset += 4
				for i := 0; i < len(bootstrapMethodArguments); i++ {
					bootstrapMethodArguments[i], _ = c.readConst(c.readUnsignedShort(bootstrapMethodOffset), charBuffer)
					bootstrapMethodOffset += 2
				}
				methodVisitor.VisitInvokeDynamicInsn(name, desc, handle.(*Handle), bootstrapMethodArguments)
				currentOffset += 5
				break
			}
		case constants.NEW, constants.ANEWARRAY, constants.CHECKCAST, constants.INSTANCEOF:
			methodVisitor.VisitTypeInsn(int(opcode), c.readClass(currentOffset+1, charBuffer))
			currentOffset += 3
			break
		case constants.IINC:
			methodVisitor.VisitIincInsn(int(b[currentOffset+1]&0xFF), int(b[currentOffset+2]))
			currentOffset += 3
			break
		case constants.MULTIANEWARRAY:
			methodVisitor.VisitMultiANewArrayInsn(c.readClass(currentOffset+1, charBuffer), int(b[currentOffset+3]&0xFF))
			currentOffset += 4
			break
		default:
			panic(errors.New("Assertion Error"))
			break
		}

		for visibleTypeAnnotationOffsets != nil && currentVisibleTypeAnnotationIndex < len(visibleTypeAnnotationOffsets) && currentVisibleTypeAnnotationBytecodeOffset <= currentBytecodeOffset {
			if currentVisibleTypeAnnotationBytecodeOffset == currentBytecodeOffset {
				currentAnnotationOffset := c.readTypeAnnotationTarget(context, visibleTypeAnnotationOffsets[currentVisibleTypeAnnotationIndex])
				annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
				currentAnnotationOffset += 2
				c.readElementValues(methodVisitor.VisitInsnAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, true), currentAnnotationOffset, true, charBuffer)
			}
			currentVisibleTypeAnnotationIndex++
			currentVisibleTypeAnnotationBytecodeOffset = c.getTypeAnnotationBytecodeOffset(visibleTypeAnnotationOffsets, currentVisibleTypeAnnotationIndex)
		}

		for invisibleTypeAnnotationOffsets != nil && currentInvisibleTypeAnnotationIndex < len(invisibleTypeAnnotationOffsets) && currentInvisibleTypeAnnotationBytecodeOffset <= currentBytecodeOffset {
			if currentInvisibleTypeAnnotationBytecodeOffset == currentBytecodeOffset {
				currentAnnotationOffset := c.readTypeAnnotationTarget(context, invisibleTypeAnnotationOffsets[currentInvisibleTypeAnnotationIndex])
				annotationDescriptor := c.readUTF8(currentAnnotationOffset, charBuffer)
				currentAnnotationOffset += 2
				c.readElementValues(methodVisitor.VisitInsnAnnotation(context.currentTypeAnnotationTarget, context.currentTypeAnnotationTargetPath, annotationDescriptor, false), currentAnnotationOffset, true, charBuffer)
			}
			currentInvisibleTypeAnnotationIndex++
			currentInvisibleTypeAnnotationBytecodeOffset = c.getTypeAnnotationBytecodeOffset(invisibleTypeAnnotationOffsets, currentInvisibleTypeAnnotationIndex)
		}
	}

	if labels[codeLength] != nil {
		methodVisitor.VisitLabel(labels[codeLength])
	}

	if localVariableTableOffset != 0 && (context.parsingOptions&SKIP_DEBUG) == 0 {
		var typeTable []int
		if localVariableTypeTableOffset != 0 {
			typeTable = make([]int, c.readUnsignedShort(localVariableTypeTableOffset)*3)
			currentOffset = localVariableTypeTableOffset + 2
			for i := len(typeTable); i > 0; {
				i--
				typeTable[i] = currentOffset + 6
				i--
				typeTable[i] = c.readUnsignedShort(currentOffset + 8)
				i--
				typeTable[i] = c.readUnsignedShort(currentOffset)
				currentOffset += 10
			}
		}
		localVariableTableLength := c.readUnsignedShort(localVariableTableOffset)
		currentOffset = localVariableTableOffset + 2
		for localVariableTableLength > 0 {
			localVariableTableLength--
			startPc := c.readUnsignedShort(currentOffset)
			length := c.readUnsignedShort(currentOffset + 2)
			name := c.readUTF8(currentOffset+4, charBuffer)
			descriptor := c.readUTF8(currentOffset+6, charBuffer)
			index := c.readUnsignedShort(currentOffset + 8)
			currentOffset += 10
			var signature string
			if typeTable != nil {
				for i := 0; i < len(typeTable); i += 3 {
					if typeTable[i] == startPc && typeTable[i+1] == index {
						signature = c.readUTF8(typeTable[i+2], charBuffer)
						break
					}
				}
			}
			methodVisitor.VisitLocalVariable(name, descriptor, signature, labels[startPc], labels[startPc+length], index)
		}
	}

	if visibleTypeAnnotationOffsets != nil {
		for i := 0; i < len(visibleTypeAnnotationOffsets); i++ {
			targetType := c.readByte(visibleTypeAnnotationOffsets[i])
			if targetType == typereference.LOCAL_VARIABLE || targetType == typereference.RESOURCE_VARIABLE {
				currentOffset = c.readTypeAnnotationTarget(context, visibleTypeAnnotationOffsets[i])
				annotationDescriptor := c.readUTF8(currentOffset, charBuffer)
				currentOffset += 2
				annotationVisitor := methodVisitor.VisitLocalVariableAnnotation(
					context.currentTypeAnnotationTarget,
					context.currentTypeAnnotationTargetPath,
					context.currentLocalVariableAnnotationRangeStarts,
					context.currentLocalVariableAnnotationRangeEnds,
					context.currentLocalVariableAnnotationRangeIndices,
					annotationDescriptor,
					true,
				)
				currentOffset = c.readElementValues(annotationVisitor, currentOffset, true, charBuffer)
			}
		}
	}

	if invisibleTypeAnnotationOffsets != nil {
		for i := 0; i < len(invisibleTypeAnnotationOffsets); i++ {
			targetType := c.readByte(visibleTypeAnnotationOffsets[i])
			if targetType == typereference.LOCAL_VARIABLE || targetType == typereference.RESOURCE_VARIABLE {
				currentOffset = c.readTypeAnnotationTarget(context, invisibleTypeAnnotationOffsets[i])
				annotationDescriptor := c.readUTF8(currentOffset, charBuffer)
				currentOffset += 2
				annotationVisitor := methodVisitor.VisitLocalVariableAnnotation(
					context.currentTypeAnnotationTarget,
					context.currentTypeAnnotationTargetPath,
					context.currentLocalVariableAnnotationRangeStarts,
					context.currentLocalVariableAnnotationRangeEnds,
					context.currentLocalVariableAnnotationRangeIndices,
					annotationDescriptor,
					false,
				)
				currentOffset = c.readElementValues(annotationVisitor, currentOffset, true, charBuffer)
			}
		}
	}

	for attributes != nil {
		nextAttribute := attributes.nextAttribute
		attributes.nextAttribute = nil
		methodVisitor.VisitAttribute(attributes)
		attributes = nextAttribute
	}

	methodVisitor.VisitMaxs(maxStack, maxLocals)
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
	charBuffer := context.charBuffer
	currentOffset := runtimeTypeAnnotationsOffset
	typeAnnotationsOffsets := make([]int, c.readUnsignedShort(currentOffset))
	currentOffset += 2

	for i := 0; i < len(typeAnnotationsOffsets); i++ {
		typeAnnotationsOffsets[i] = currentOffset
		targetType := c.readInt(currentOffset)
		switch targetType >> 24 {
		case typereference.CLASS_TYPE_PARAMETER, typereference.METHOD_TYPE_PARAMETER, typereference.METHOD_FORMAL_PARAMETER:
			currentOffset += 2
			break
		case typereference.FIELD, typereference.METHOD_RETURN, typereference.METHOD_RECEIVER:
			currentOffset++
			break
		case typereference.LOCAL_VARIABLE, typereference.RESOURCE_VARIABLE:
			tableLength := c.readUnsignedShort(currentOffset + 1)
			currentOffset += 3
			for tableLength > 0 {
				tableLength--
				startPc := c.readUnsignedShort(currentOffset)
				length := c.readUnsignedShort(currentOffset + 2)
				currentOffset += 6
				c.createLabel(startPc, context.currentMethodLabels)
				c.createLabel(startPc+length, context.currentMethodLabels)
			}
			break
		case typereference.CAST, typereference.CONSTRUCTOR_INVOCATION_TYPE_ARGUMENT, typereference.METHOD_INVOCATION_TYPE_ARGUMENT,
			typereference.CONSTRUCTOR_REFERENCE_TYPE_ARGUMENT, typereference.METHOD_REFERENCE_TYPE_ARGUMENT:
			currentOffset += 4
			break
		case typereference.CLASS_EXTENDS, typereference.CLASS_TYPE_PARAMETER_BOUND, typereference.METHOD_TYPE_PARAMETER_BOUND,
			typereference.THROWS, typereference.EXCEPTION_PARAMETER, typereference.INSTANCEOF, typereference.NEW, typereference.CONSTRUCTOR_REFERENCE,
			typereference.METHOD_REFERENCE:
			currentOffset += 3
			break
		default:
			panic(errors.New("Assertion Error"))
			break
		}

		pathLength := c.readByte(currentOffset)
		if (targetType >> 24) == typereference.EXCEPTION_PARAMETER {
			var path *TypePath
			if pathLength != 0 {
				path = NewTypePath(c.b, currentOffset)
			}
			currentOffset += 1 + 2*int(pathLength)
			annotationDescriptor := c.readUTF8(currentOffset, charBuffer)
			currentOffset += 2
			currentOffset = c.readElementValues(methodVisitor.VisitTryCatchAnnotation(targetType&0xFFFFF00, path, annotationDescriptor, visible), currentOffset, true, charBuffer)
		} else {
			currentOffset += 3 + 2*int(pathLength)
			currentOffset = c.readElementValues(nil, currentOffset, true, charBuffer)
		}
	}

	return typeAnnotationsOffsets
}

func (c ClassReader) getTypeAnnotationBytecodeOffset(typeAnnotationOffsets []int, typeAnnotationIndex int) int {
	if typeAnnotationOffsets == nil || typeAnnotationIndex >= len(typeAnnotationOffsets) || c.readByte(typeAnnotationOffsets[typeAnnotationIndex]) < typereference.INSTANCEOF {
		return -1
	}
	return c.readUnsignedShort(typeAnnotationOffsets[typeAnnotationIndex] + 1)
}

func (c ClassReader) readTypeAnnotationTarget(context *Context, typeAnnotationOffset int) int {
	currentOffset := typeAnnotationOffset
	targetType := c.readInt(typeAnnotationOffset)

	switch targetType >> 24 {
	case typereference.CLASS_TYPE_PARAMETER, typereference.METHOD_TYPE_PARAMETER, typereference.METHOD_FORMAL_PARAMETER:
		targetType &= 0xFFFF0000
		currentOffset += 2
		break
	case typereference.FIELD, typereference.METHOD_RETURN, typereference.METHOD_RECEIVER:
		targetType &= 0xFF000000
		currentOffset++
		break
	case typereference.LOCAL_VARIABLE, typereference.RESOURCE_VARIABLE:
		targetType &= 0xFF000000
		tableLength := c.readUnsignedShort(currentOffset + 1)
		currentOffset += 3
		context.currentLocalVariableAnnotationRangeStarts = make([]*Label, tableLength)
		context.currentLocalVariableAnnotationRangeEnds = make([]*Label, tableLength)
		context.currentLocalVariableAnnotationRangeIndices = make([]int, tableLength)
		for i := 0; i < tableLength; i++ {
			startPc := c.readUnsignedShort(currentOffset)
			length := c.readUnsignedShort(currentOffset + 2)
			index := c.readUnsignedShort(currentOffset + 4)
			currentOffset += 6
			context.currentLocalVariableAnnotationRangeStarts[i] = c.createLabel(startPc, context.currentMethodLabels)
			context.currentLocalVariableAnnotationRangeEnds[i] = c.createLabel(startPc+length, context.currentMethodLabels)
			context.currentLocalVariableAnnotationRangeIndices[i] = index
		}
		break
	case typereference.CAST, typereference.CONSTRUCTOR_INVOCATION_TYPE_ARGUMENT, typereference.METHOD_INVOCATION_TYPE_ARGUMENT,
		typereference.CONSTRUCTOR_REFERENCE_TYPE_ARGUMENT, typereference.METHOD_REFERENCE_TYPE_ARGUMENT:
		targetType &= 0xFF0000FF
		currentOffset += 4
		break
	case typereference.CLASS_EXTENDS, typereference.CLASS_TYPE_PARAMETER_BOUND, typereference.METHOD_TYPE_PARAMETER_BOUND,
		typereference.THROWS, typereference.EXCEPTION_PARAMETER:
		targetType &= 0xFFFFFF00
		currentOffset += 3
		break
	case typereference.INSTANCEOF, typereference.NEW, typereference.CONSTRUCTOR_REFERENCE, typereference.METHOD_REFERENCE:
		targetType &= 0xFF000000
		currentOffset += 3
		break
	default:
		panic(errors.New("Assertion Error"))
		break
	}
	context.currentTypeAnnotationTarget = targetType
	pathLength := c.readByte(currentOffset)
	if pathLength != 0 {
		context.currentTypeAnnotationTargetPath = NewTypePath(c.b, currentOffset)
	}

	return currentOffset + 1 + 2*int(pathLength)
}

func (c ClassReader) readParameterAnnotations(methodVisitor MethodVisitor, context *Context, runtimeParameterAnnotationsOffset int, visible bool) {
	currentOffset := runtimeParameterAnnotationsOffset
	numParameters := c.b[currentOffset] & 0xFF
	currentOffset++
	methodVisitor.VisitAnnotableParameterCount(int(numParameters), visible)
	charBuffer := context.charBuffer
	for i := 0; i < int(numParameters); i++ {
		numAnnotations := c.readUnsignedShort(currentOffset)
		currentOffset += 2
		for numAnnotations > 0 {
			numAnnotations--
			annotationDescriptor := c.readUTF8(currentOffset, charBuffer)
			currentOffset += 2
			currentOffset = c.readElementValues(methodVisitor.VisitParameterAnnotation(i, annotationDescriptor, visible), currentOffset, true, charBuffer)
		}
	}
}

func (c ClassReader) readElementValues(annotationVisitor AnnotationVisitor, annotationOffset int, named bool, charBuffer []rune) int {
	currentOffset := annotationOffset
	numElementValuePairs := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	if named {
		for numElementValuePairs > 0 {
			numElementValuePairs--
			elementName := c.readUTF8(currentOffset, charBuffer)
			currentOffset = c.readElementValue(annotationVisitor, currentOffset+2, elementName, charBuffer)
		}
	} else {
		for numElementValuePairs > 0 {
			numElementValuePairs--
			currentOffset = c.readElementValue(annotationVisitor, currentOffset, "", charBuffer)
		}
	}
	if annotationVisitor != nil {
		annotationVisitor.VisitEnd()
	}
	return currentOffset
}

func (c ClassReader) readElementValue(annotationVisitor AnnotationVisitor, elementValueOffset int, elementName string, charBuffer []rune) int {
	currentOffset := elementValueOffset
	if annotationVisitor == nil {
		switch c.b[currentOffset] & 0xFF {
		case 'e':
			return currentOffset + 5
		case '@':
			return c.readElementValues(nil, currentOffset+3, true, charBuffer)
		case '[':
			return c.readElementValues(nil, currentOffset+1, false, charBuffer)
		default:
			return currentOffset + 3
		}
	}
	switch c.b[currentOffset] & 0xFF {
	case 'B':
		currentOffset++
		annotationVisitor.Visit(elementName, byte(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset)])))
		currentOffset += 2
		break
	case 'C':
		currentOffset++
		annotationVisitor.Visit(elementName, rune(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset)])))
		currentOffset += 2
		break
	case 'D', 'F', 'I', 'J':
		currentOffset++
		constd, _ := c.readConst(c.readUnsignedShort(currentOffset), charBuffer)
		annotationVisitor.Visit(elementName, constd)
		currentOffset += 2
		break
	case 'S':
		currentOffset++
		annotationVisitor.Visit(elementName, int16(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset)])))
		currentOffset += 2
		break
	case 'Z':
		currentOffset++
		val := true
		if c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset)]) == 0 {
			val = false
		}
		annotationVisitor.Visit(elementName, val)
		currentOffset += 2
		break
	case 's':
		currentOffset++
		annotationVisitor.Visit(elementName, c.readUTF8(currentOffset, charBuffer))
		currentOffset += 2
		break
	case 'e':
		currentOffset++
		annotationVisitor.VisitEnum(elementName, c.readUTF8(currentOffset, charBuffer), c.readUTF8(currentOffset+2, charBuffer))
		currentOffset += 4
		break
	case 'c':
		currentOffset++
		annotationVisitor.Visit(elementName, getType(c.readUTF8(currentOffset, charBuffer)))
		currentOffset += 2
		break
	case '@':
		currentOffset++
		currentOffset = c.readElementValues(annotationVisitor.VisitArray(elementName), currentOffset-2, false, charBuffer)
		break
	case '[':
		currentOffset++
		numValues := c.readUnsignedShort(currentOffset)
		currentOffset += 2
		if numValues == 0 {
			return c.readElementValues(annotationVisitor.VisitArray(elementName), currentOffset-2, false, charBuffer)
		}
		switch c.b[currentOffset] & 0xFF {
		case 'B':
			byteValues := make([]byte, numValues)
			for i := 0; i < numValues; i++ {
				byteValues[i] = byte(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]))
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, byteValues)
			break
		case 'Z':
			boolenValues := make([]bool, numValues)
			for i := 0; i < numValues; i++ {
				boolenValues[i] = c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]) != 0
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, boolenValues)
			break
		case 'S':
			shortValues := make([]int16, numValues)
			for i := 0; i < numValues; i++ {
				shortValues[i] = int16(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]))
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, shortValues)
			break
		case 'C':
			charValues := make([]rune, numValues)
			for i := 0; i < numValues; i++ {
				charValues[i] = rune(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]))
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, charValues)
			break
		case 'I':
			intValues := make([]int, numValues)
			for i := 0; i < numValues; i++ {
				intValues[i] = c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)])
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, intValues)
			break
		case 'J':
			longValues := make([]int64, numValues)
			for i := 0; i < numValues; i++ {
				longValues[i] = c.readLong(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)])
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, longValues)
			break
		case 'F':
			floatValues := make([]float32, numValues)
			for i := 0; i < numValues; i++ {
				floatValues[i] = float32(c.readInt(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]))
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, floatValues)
			break
		case 'D':
			doubleValues := make([]float64, numValues)
			for i := 0; i < numValues; i++ {
				doubleValues[i] = float64(c.readLong(c.cpInfoOffsets[c.readUnsignedShort(currentOffset+1)]))
				currentOffset += 3
			}
			annotationVisitor.Visit(elementName, doubleValues)
			break
		default:
			currentOffset = c.readElementValues(annotationVisitor.VisitArray(elementName), currentOffset-2, false, charBuffer)
			break
		}
		break
	default:
		panic(errors.New("Assertion Error"))
		break
	}
	return currentOffset
}

// ----------------------------------------------------------------------------------------------
// Methods to parse stack map frames
// ----------------------------------------------------------------------------------------------

func (c ClassReader) computeImplicitFame(context *Context) {
	methodDescriptor := context.currentMethodDescriptor
	locals := context.currentFrameLocalTypes
	nLocal := 0
	if (context.currentMethodAccessFlags & opcodes.ACC_STATIC) == 0 {
		if "<init>" == context.currentMethodName {
			locals[nLocal] = opcodes.UNINITIALIZED_THIS
			nLocal++
		} else {
			locals[nLocal] = c.readClass(c.header+2, context.charBuffer)
			nLocal++
		}
	}

	currentMethodDescriptorOffset := 1
	for {
		currentArgumentDescriptorStartOffset := currentMethodDescriptorOffset
		currentMethodDescriptorOffset++
		switch methodDescriptor[currentMethodDescriptorOffset-1] {
		case 'Z', 'C', 'B', 'S', 'I':
			locals[nLocal] = opcodes.INTEGER
			nLocal++
			break
		case 'F':
			locals[nLocal] = opcodes.FLOAT
			nLocal++
			break
		case 'J':
			locals[nLocal] = opcodes.LONG
			nLocal++
			break
		case 'D':
			locals[nLocal] = opcodes.DOUBLE
			nLocal++
			break
		case '[':
			for methodDescriptor[currentMethodDescriptorOffset] == '[' {
				currentMethodDescriptorOffset++
			}
			if methodDescriptor[currentMethodDescriptorOffset] == 'L' {
				currentMethodDescriptorOffset++
				for methodDescriptor[currentMethodDescriptorOffset] != ';' {
					currentMethodDescriptorOffset++
				}
			}
			currentMethodDescriptorOffset++
			locals[nLocal] = methodDescriptor[currentArgumentDescriptorStartOffset:currentMethodDescriptorOffset]
			nLocal++
			break
		case 'L':
			for methodDescriptor[currentMethodDescriptorOffset] != ';' {
				currentMethodDescriptorOffset++
			}
			locals[nLocal] = methodDescriptor[currentArgumentDescriptorStartOffset+1 : currentMethodDescriptorOffset]
			currentMethodDescriptorOffset++
			nLocal++
			break
		default:
			context.currentFrameLocalCount = nLocal
			return
		}
	}
}

func (c ClassReader) readStackMapFrame(stackMapFrameOffset int, compressed bool, expand bool, context *Context) int {
	currentOffset := stackMapFrameOffset
	charBuffer := context.charBuffer
	labels := context.currentMethodLabels
	var frameType int
	if compressed {
		frameType = int(c.b[currentOffset] & 0xFF)
		currentOffset++
	} else {
		frameType = frame.FULL_FRAME
		context.currentFrameOffset = -1
	}
	var offsetDelta int
	context.currentFrameLocalCount = 0
	if frameType < frame.SAME_LOCALS_1_STACK_ITEM_FRAME {
		offsetDelta = frameType
		context.currentFrameType = opcodes.F_SAME
		context.currentFrameStackCount = 0
	} else if frameType < frame.RESERVED {
		offsetDelta = frameType - frame.SAME_LOCALS_1_STACK_ITEM_FRAME
		currentOffset = c.readVerificationTypeInfo(currentOffset, context.currentFrameStackTypes, 0, charBuffer, labels)
		context.currentFrameType = opcodes.F_SAME1
		context.currentFrameStackCount = 1
	} else {
		offsetDelta = c.readUnsignedShort(currentOffset)
		currentOffset += 2
		if frameType == frame.SAME_LOCALS_1_STACK_ITEM_FRAME_EXTENDED {
			currentOffset = c.readVerificationTypeInfo(currentOffset, context.currentFrameStackTypes, 0, charBuffer, labels)
			context.currentFrameType = opcodes.F_SAME1
			context.currentFrameStackCount = 1
		} else if frameType >= frame.CHOP_FRAME && frameType < frame.SAME_FRAME_EXTENDED {
			context.currentFrameType = opcodes.F_CHOP
			context.currentFrameLocalCountDelta = frame.SAME_FRAME_EXTENDED - frameType
			context.currentFrameLocalCount -= context.currentFrameLocalCountDelta
			context.currentFrameStackCount = 0
		} else if frameType == frame.SAME_FRAME_EXTENDED {
			context.currentFrameType = opcodes.F_SAME
			context.currentFrameStackCount = 0
		} else if frameType < frame.FULL_FRAME {
			local := 0
			if expand {
				local = context.currentFrameLocalCount
			}
			for k := frameType - frame.SAME_FRAME_EXTENDED; k > 0; k-- {
				currentOffset = c.readVerificationTypeInfo(currentOffset, context.currentFrameLocalTypes, local, charBuffer, labels)
				local++
			}
			context.currentFrameType = opcodes.F_APPEND
			context.currentFrameLocalCountDelta = frameType - frame.SAME_FRAME_EXTENDED
			context.currentFrameLocalCount += context.currentFrameLocalCountDelta
			context.currentFrameStackCount = 0
		} else {
			numberOfLocals := c.readUnsignedShort(currentOffset)
			currentOffset += 2
			context.currentFrameType = opcodes.F_FULL
			context.currentFrameLocalCountDelta = numberOfLocals
			context.currentFrameLocalCount = numberOfLocals
			for local := 0; local < numberOfLocals; local++ {
				currentOffset = c.readVerificationTypeInfo(currentOffset, context.currentFrameLocalTypes, local, charBuffer, labels)
			}
			numberOfStackItems := c.readUnsignedShort(currentOffset)
			currentOffset += 2
			context.currentFrameStackCount = numberOfStackItems
			for stack := 0; stack < numberOfStackItems; stack++ {
				currentOffset = c.readVerificationTypeInfo(currentOffset, context.currentFrameStackTypes, stack, charBuffer, labels)
			}
		}
	}
	context.currentFrameOffset += offsetDelta + 1
	c.createLabel(context.currentFrameOffset, labels)
	return currentOffset
}

func (c ClassReader) readVerificationTypeInfo(verificationTypeInfoOffset int, framed []interface{}, index int, charBuffer []rune, labels []*Label) int {
	currentOffset := verificationTypeInfoOffset
	tag := c.b[currentOffset] & 0xFF
	currentOffset++
	switch tag {
	case frame.ITEM_TOP:
		framed[index] = opcodes.TOP
		break
	case frame.ITEM_INTEGER:
		framed[index] = opcodes.INTEGER
		break
	case frame.ITEM_FLOAT:
		framed[index] = opcodes.FLOAT
		break
	case frame.ITEM_DOUBLE:
		framed[index] = opcodes.DOUBLE
		break
	case frame.ITEM_LONG:
		framed[index] = opcodes.LONG
		break
	case frame.ITEM_NULL:
		framed[index] = opcodes.NULL
		break
	case frame.ITEM_UNINITIALIZED_THIS:
		framed[index] = opcodes.UNINITIALIZED_THIS
		break
	case frame.ITEM_OBJECT:
		framed[index] = c.readClass(currentOffset, charBuffer)
		currentOffset += 2
		break
	default:
		framed[index] = c.createLabel(c.readUnsignedShort(currentOffset), labels)
		currentOffset += 2
	}
	return currentOffset
}

// ----------------------------------------------------------------------------------------------
// Methods to parse attributes
// ----------------------------------------------------------------------------------------------

func (c ClassReader) getFirstAttributeOffset() int {
	currentOffset := c.header + 8 + c.readUnsignedShort(c.header+6)*2
	fieldsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for fieldsCount > 0 {
		fieldsCount--
		attributesCount := c.readUnsignedShort(currentOffset + 6)
		currentOffset += 8
		for attributesCount > 0 {
			attributesCount--
			currentOffset += 6 + c.readInt(currentOffset+2)
		}
	}

	methodsCount := c.readUnsignedShort(currentOffset)
	currentOffset += 2
	for methodsCount > 0 {
		methodsCount--
		attributesCount := c.readUnsignedShort(currentOffset + 6)
		currentOffset += 8
		for attributesCount > 0 {
			attributesCount--
			currentOffset += 6 + c.readInt(currentOffset+2)
		}
	}

	return currentOffset + 2
}

func (c ClassReader) readAttribute(attributePrototypes []*Attribute, typed string, offset int, length int, charBuffer []rune, codeAttributeOffset int, labels []*Label) *Attribute {
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
			currentOffset++
			strLength++
		}
	}
	str := make([]rune, strLength)
	copy(str, charBuffer[0:strLength])
	return string(str)
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
		return getObjectType(c.readUTF8(cpInfoOffset, charBuffer)), nil
	case byte(symbol.CONSTANT_STRING_TAG):
		return c.readUTF8(cpInfoOffset, charBuffer), nil
	case byte(symbol.CONSTANT_METHOD_TYPE_TAG):
		return getMethodType(c.readUTF8(cpInfoOffset, charBuffer)), nil
	case byte(symbol.CONSTANT_METHOD_HANDLE_TAG):
		referenceKind := c.readByte(cpInfoOffset)
		referenceCpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(cpInfoOffset+1)]
		nameAndTypeCpInfoOffset := c.cpInfoOffsets[c.readUnsignedShort(referenceCpInfoOffset+2)]
		owner := c.readClass(referenceCpInfoOffset, charBuffer)
		name := c.readUTF8(nameAndTypeCpInfoOffset, charBuffer)
		desc := c.readUTF8(nameAndTypeCpInfoOffset+2, charBuffer)
		itf := c.b[referenceCpInfoOffset-1] == byte(symbol.CONSTANT_INTERFACE_METHODREF_TAG)
		return &Handle{
			tag:         int(referenceKind),
			owner:       owner,
			name:        name,
			descriptor:  desc,
			isInterface: itf,
		}, nil
	default:
		return nil, errors.New("Assertion Error")
	}
}
