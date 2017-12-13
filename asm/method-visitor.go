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
	VisitParameter(name string, access int)
	VisitAnnotationDefault() AnnotationVisitor
	VisitAnnotation(descriptor string, visible bool) AnnotationVisitor
	VisitTypeAnnotation(typeRef int, typePath interface{}, descriptor string, visible bool) AnnotationVisitor //TypePath
	VisitAnnotableParameterCount(parameterCount int, visible bool)
	VisitParameterAnnotation(parameter int, descriptor string, visible bool) AnnotationVisitor
	VisitAttribute(attribute *Attribute)
	VisitCode()
	VisitFrame(typed, nLocal int, local interface{}, nStack int, stack interface{})
	VisitInsn(opcode int)
	VisitIntInsn(opcode, operand int)
	VisitVarInsn(opcode, vard int)
	VisitTypeInsn(opcode, typed int)
	VisitFieldInsn(opcode int, owner, name, descriptor string)
	VisitMethodInsn(opcode int, owner, name, descriptor string)
	VisitMethodInsnB(opcode int, owner, name, descriptor string, isInterface bool)
	VisitInvokeDynamicInsn(name, descriptor string, bootstrapMethodHande interface{}, bootstrapMethodArguments ...interface{}) //Handle
	VisitJumpInsn(opcode int, label *Label)
	VisitLabel(label *Label)
	VisitLdcInsn(value interface{})
	VisitIincInsn(vard, increment int)
	VisitTableSwitchInsn(min, max int, dflt *Label, labels ...*Label)
	VisitLookupSwitchInsn(dflt *Label, keys []int, labels []Label)
	VisitMultiANewArrayInsn(descriptor string, numDimensions int)
	VisitInsnAnnotation(typeRef int, typePath interface{}, descriptor string, visible bool) AnnotationVisitor //TypePath
	VisitTryCatchBlock(start, end, handler *Label, typed string)
	VisitTryCatchAnnotation(typeRef int, typePath interface{}, descriptor string, visible bool) AnnotationVisitor //TypePath
	VisitLocalVariable(name, descriptor, signature string, start, end *Label, index int)
	VisitLocalVariableAnnotation(typeRef int, typePath interface{}, start, end *Label, index []int, descriptor string, visible bool) AnnotationVisitor //TypePath
	VisitLineNumber(line int, start *Label)
	VisitMaxs(maxStack int, maxLocals int)
	VisitEnd()
}
