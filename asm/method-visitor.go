package asm

// MethodVisitor a visitor to visit a Java method. The methods of this class must be called in the following
// order: ( <tt>visitParameter</tt> )* [ <tt>visitAnnotationDefault</tt> ] (
// <tt>visitAnnotation</tt> | <tt>visitAnnotableParameterCount</tt> |
// <tt>visitParameterAnnotation</tt> <tt>visitTypeAnnotation</tt> | <tt>visitAttribute</tt> )* [
// <tt>visitCode</tt> ( <tt>visitFrame</tt> | <tt>visit<i>X</i>Insn</tt> | <tt>visitLabel</tt> |
// <tt>visitInsnAnnotation</tt> | <tt>visitTryCatchBlock</tt> | <tt>visitTryCatchAnnotation</tt> |
// <tt>visitLocalVariable</tt> | <tt>visitLocalVariableAnnotation</tt> | <tt>visitLineNumber</tt> )*
// <tt>visitMaxs</tt> ] <tt>visitEnd</tt>. In addition, the <tt>visit<i>X</i>Insn</tt> and
// <tt>visitLabel</tt> methods must be called in the sequential order of the bytecode instructions
// of the visited code, <tt>visitInsnAnnotation</tt> must be called <i>after</i> the annotated
// instruction, <tt>visitTryCatchBlock</tt> must be called <i>before</i> the labels passed as
// arguments have been visited, <tt>visitTryCatchBlockAnnotation</tt> must be called <i>after</i>
// the corresponding try catch block has been visited, and the <tt>visitLocalVariable</tt>,
// <tt>visitLocalVariableAnnotation</tt> and <tt>visitLineNumber</tt> methods must be called
// <i>after</i> the labels passed as arguments have been visited.
type MethodVisitor interface {
	visitParameter(name string, access int)
	visitAnnotationDefault() AnnotationVisitor
	visitAnnotation(descriptor string, visible bool) AnnotationVisitor
	visitTypeAnnotation(typeRef int, typePath interface{}, descriptor string, visible bool) AnnotationVisitor //TypePath
	visitAnnotableParameterCount(parameterCount int, visible bool)
	visitParameterAnnotation(parameter int, descriptor string, visible bool) AnnotationVisitor
	visitAttribute(attribute *Attribute)
	visitCode()
	visitFrame(typed, nLocal int, local interface{}, nStack int, stack interface{})
	visitInsn(opcode int)
	visitIntInsn(opcode, operand int)
	visitVarInsn(opcode, vard int)
	visitTypeInsn(opcode, typed int)
	visitFieldInsn(opcode int, owner, name, descriptor string)
	visitMethodInsn(opcode int, owner, name, descriptor string)
	_visitMethodInsn(opcode int, owner, name, descriptor string, isInterface bool)
	visitInvokeDynamicInsn(name, descriptor string, bootstrapMethodHande interface{}, bootstrapMethodArguments ...interface{}) //Handle
	visitJumpInsn(opcode int, label *Label)
	visitLabel(label *Label)
	visitLdcInsn(value interface{})
	visitIincInsn(vard, increment int)
	visitTableSwitchInsn(min, max int, dflt *Label, labels ...*Label)
	visitLookupSwitchInsn(dflt *Label, keys []int, labels []Label)
	visitMultiANewArrayInsn(descriptor string, numDimensions int)
	visitInsnAnnotation(typeRef int, typePath interface{}, descriptor string, visible bool) AnnotationVisitor //TypePath
	visitTryCatchBlock(start, end, handler *Label, typed string)
	visitTryCatchAnnotation(typeRef int, typePath interface{}, descriptor string, visible bool) AnnotationVisitor //TypePath
	visitLocalVariable(name, descriptor, signature string, start, end *Label, index int)
	visitLocalVariableAnnotation(typeRef int, typePath interface{}, start, end *Label, index []int, descriptor string, visible bool) AnnotationVisitor //TypePath
	visitLineNumber(line int, start *Label)
	visitMaxs(maxStack int, maxLocals int)
	visitEnd()
}
