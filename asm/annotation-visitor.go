package asm

// AnnotationVisitor a visitor to visit a Java annotation. The methods of this class must be called in the following
// order: ( <tt>visit</tt> | <tt>visitEnum</tt> | <tt>visitAnnotation</tt> | <tt>visitArray</tt> )*
// <tt>visitEnd</tt>.
type AnnotationVisitor interface {
	visit(name string, value interface{})
	visitEnum(name, descriptor, value string)
	visitAnnotation(name, descriptor string) AnnotationVisitor
	visitArray(name string) AnnotationVisitor
	visitEnd()
}
