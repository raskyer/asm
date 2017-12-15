package helper

import "github.com/leaklessgfy/asm/asm"

type ClassVisitor struct {
	OnVisit       func(version, access int, name, signature, superName string, interfaces []string)
	OnVisitField  func(access int, name, descriptor, signature string, value interface{}) asm.FieldVisitor
	OnVisitMethod func(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor
	OnVisitEnd    func()
}

func (c ClassVisitor) Visit(version, access int, name, signature, superName string, interfaces []string) {
	if c.OnVisit != nil {
		c.OnVisit(version, access, name, signature, superName, interfaces)
	}
}

func (c ClassVisitor) VisitSource(source, debug string) {

}

func (c ClassVisitor) VisitModule(name string, access int, version string) asm.ModuleVisitor {
	return nil
}

func (c ClassVisitor) VisitOuterClass(owner, name, descriptor string) {

}

func (c ClassVisitor) VisitAnnotation(descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (c ClassVisitor) VisitTypeAnnotation(typeRef int, typePath *asm.TypePath, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (c ClassVisitor) VisitAttribute(attribute *asm.Attribute) {

}

func (c ClassVisitor) VisitInnerClass(name, outerName, innerName string, access int) {

}

func (c ClassVisitor) VisitField(access int, name, descriptor, signature string, value interface{}) asm.FieldVisitor {
	if c.OnVisitField != nil {
		return c.OnVisitField(access, name, descriptor, signature, value)
	}
	return nil
}

func (c ClassVisitor) VisitMethod(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor {
	if c.OnVisitMethod != nil {
		return c.OnVisitMethod(access, name, descriptor, signature, exceptions)
	}
	return nil
}

func (c ClassVisitor) VisitEnd() {
	if c.OnVisitEnd != nil {
		c.OnVisitEnd()
	}
}

type MethodVisitor struct {
	OnVisitLineNumber func(line int, start *asm.Label)
	OnVisitTypeInsn   func(opcode int, typed string)
}

func (m MethodVisitor) VisitParameter(name string, access int) {

}

func (m MethodVisitor) VisitAnnotationDefault() asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitAnnotation(descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitTypeAnnotation(typeRef int, typePath *asm.TypePath, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitAnnotableParameterCount(parameterCount int, visible bool) {

}

func (m MethodVisitor) VisitParameterAnnotation(parameter int, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitAttribute(attribute *asm.Attribute) {

}

func (m MethodVisitor) VisitCode() {

}

func (m MethodVisitor) VisitFrame(typed, nLocal int, local interface{}, nStack int, stack interface{}) {

}

func (m MethodVisitor) VisitInsn(opcode int) {

}

func (m MethodVisitor) VisitIntInsn(opcode, operand int) {

}

func (m MethodVisitor) VisitVarInsn(opcode, vard int) {

}

func (m MethodVisitor) VisitTypeInsn(opcode int, typed string) {
	if m.OnVisitTypeInsn != nil {
		m.OnVisitTypeInsn(opcode, typed)
	}
}

func (m MethodVisitor) VisitFieldInsn(opcode int, owner, name, descriptor string) {

}

func (m MethodVisitor) VisitMethodInsn(opcode int, owner, name, descriptor string) {

}

func (m MethodVisitor) VisitMethodInsnB(opcode int, owner, name, descriptor string, isInterface bool) {

}

func (m MethodVisitor) VisitInvokeDynamicInsn(name, descriptor string, bootstrapMethodHande *asm.Handle, bootstrapMethodArguments ...interface{}) {

}

func (m MethodVisitor) VisitJumpInsn(opcode int, label *asm.Label) {

}

func (m MethodVisitor) VisitLabel(label *asm.Label) {

}

func (m MethodVisitor) VisitLdcInsn(value interface{}) {

}

func (m MethodVisitor) VisitIincInsn(vard, increment int) {

}

func (m MethodVisitor) VisitTableSwitchInsn(min, max int, dflt *asm.Label, labels ...*asm.Label) {

}

func (m MethodVisitor) VisitLookupSwitchInsn(dflt *asm.Label, keys []int, labels []*asm.Label) {

}

func (m MethodVisitor) VisitMultiANewArrayInsn(descriptor string, numDimensions int) {

}

func (m MethodVisitor) VisitInsnAnnotation(typeRef int, typePath *asm.TypePath, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitTryCatchBlock(start, end, handler *asm.Label, typed string) {

}

func (m MethodVisitor) VisitTryCatchAnnotation(typeRef int, typePath *asm.TypePath, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitLocalVariable(name, descriptor, signature string, start, end *asm.Label, index int) {

}

func (m MethodVisitor) VisitLocalVariableAnnotation(typeRef int, typePath *asm.TypePath, start, end []*asm.Label, index []int, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (m MethodVisitor) VisitLineNumber(line int, start *asm.Label) {
	if m.OnVisitLineNumber != nil {
		m.OnVisitLineNumber(line, start)
	}
}

func (m MethodVisitor) VisitMaxs(maxStack int, maxLocals int) {

}

func (m MethodVisitor) VisitEnd() {

}
