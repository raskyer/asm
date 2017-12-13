package asm

// AnnotationVisitor a visitor to visit a Java annotation. The methods of this class must be called in the following
// order: ( <tt>visit</tt> | <tt>visitEnum</tt> | <tt>visitAnnotation</tt> | <tt>visitArray</tt> )*
// <tt>visitEnd</tt>.
type AnnotationVisitor interface {
	Visit(name string, value interface{})
	VisitEnum(name, descriptor, value string)
	VisitAnnotation(name, descriptor string) AnnotationVisitor
	VisitArray(name string) AnnotationVisitor
	VisitEnd()
}
