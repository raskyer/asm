package asm

type Attribute struct {
	typed         string
	content       []byte
	nextAttribute *Attribute
}

func NewAttribute(typed string) *Attribute {
	return &Attribute{
		typed: typed,
	}
}

func (a Attribute) isUnknow() bool {
	return true
}

func (a Attribute) isCodeAttribute() bool {
	return false
}

func (a Attribute) getLabels() []Label {
	return nil
}

func (a Attribute) read(classReader *ClassReader, offset int, length int, charBuffer []rune, codeAttributeOffset int, labels []*Label) *Attribute {
	attribute := NewAttribute(a.typed)
	attribute.content = make([]byte, length)
	//System.arraycopy(classReader.b, offset, attribute.content, 0, length)
	return attribute
}

//ClassWriter
func (a Attribute) write(classWriter interface{}, code []byte, codeLength int, maxStack int, maxLocals int) {
	//return new ByteVector(content)
}

func (a Attribute) getAttributeCount() int {
	count := 0
	attribute := &a
	for attribute != nil {
		count++
		attribute = attribute.nextAttribute
	}
	return count
}

func (a Attribute) computeAttributesSize(symbolTable interface{}) int {
	codeLength := 0
	maxStack := -1
	maxLocals := -1
	return a._computeAttributesSize(symbolTable, nil, codeLength, maxStack, maxLocals)
}

func (a Attribute) _computeAttributesSize(symbolTable interface{}, code []byte, codeLength int, maxStack int, maxLocals int) int {
	//ClassWriter classWrite = symbolTable.classWriter
	size := 0
	attribute := &a
	for attribute != nil {
		//symbolTable.addConstantUtf8(attribute.typed)
		//size += 6 + attribute.write(classWriter, code, codeLength, maxStack, maxLocals).length
		attribute = attribute.nextAttribute
	}
	return size
}

//SymbolTable, ByteVector
func (a Attribute) putAttribute(symbolTable interface{}, output interface{}) {
	codeLength := 0
	maxStack := -1
	maxLocals := -1
	a._putAttribute(symbolTable, nil, codeLength, maxStack, maxLocals, output)
}

func (a Attribute) _putAttribute(symbolTable interface{}, code []byte, codeLength int, maxStack int, maxLocals int, output interface{}) {
	//ClassWriter classWrite = symbolTable.classWriter
	attribute := &a
	for attribute != nil {
		//ByteVector attributeContent = attribute.write(classWriter, code, codeLength, maxStack, maxLocals)
		//output.putShort(symbolTable.addConstantUtf8(attribute.typed)).putInt(attributeContent.length)
		//output.putByteArray(attributeContent.data, 0, attributeContent.length)
		attribute = attribute.nextAttribute
	}
}
