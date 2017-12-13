package asm

// ClassVisitor A visitor to visit a Java class. The methods of this class must be called in the following order:
// <tt>visit</tt> [ <tt>visitSource</tt> ] [ <tt>visitModule</tt> ][ <tt>visitOuterClass</tt> ] (
// <tt>visitAnnotation</tt> | <tt>visitTypeAnnotation</tt> | <tt>visitAttribute</tt> )* (
// <tt>visitInnerClass</tt> | <tt>visitField</tt> | <tt>visitMethod</tt> )* <tt>visitEnd</tt>.
type ClassVisitor interface {
	Visit(version, access int, name, signature, superName string, interfaces []string)
	VisitSource(source, debug string)
	VisitModule(name string, access, version int) //should return modulevisitor
	VisitOuterClass(owner, name, descriptor string)
	VisitAnnotation(descriptor string, visible bool) AnnotationVisitor
	VisitTypeAnnotation(typeRef, typePath int, descriptor string, visible bool) AnnotationVisitor //typePath : TypePath
	VisitAttribute(attribute *Attribute)
	VisitInnerClass(name, outerName, innerName string, access int)
	VisitField(access int, name, descriptor, signature string, value interface{}) //should return FieldVisitor
	VisitMethod(access int, name, descriptor, signature string, exceptions []string) MethodVisitor
	VisitEnd()
}
