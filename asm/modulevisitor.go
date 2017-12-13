package asm

// ModuleVisitor a visitor to visit a Java module. The methods of this class must be called in the following
// order: <tt>visitMainClass</tt> | ( <tt>visitPackage</tt> | <tt>visitRequire</tt> |
// <tt>visitExport</tt> | <tt>visitOpen</tt> | <tt>visitUse</tt> | <tt>visitProvide</tt> )*
// <tt>visitEnd</tt>.
type ModuleVisitor interface {
	VisitMainClass(mainClass string)
	VisitPackage(packaze string)
	VisitRequire(module string, access int, version string)
	VisitExport(packaze string, access int, modules ...string)
	VisitOpen(packaze string, access int, modules ...string)
	VisitUse(service string)
	VisitProvide(service string, providers ...string)
	VisitEnd()
}
