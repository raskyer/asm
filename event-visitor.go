package main

import (
	"github.com/leaklessgfy/asm/asm"
)

type EventVisitor struct {
	OnVisit    []func(version, access int, name, signature, superName string, interfaces []string)
	OnVisitEnd []func()
}

func (e EventVisitor) Visit(version, access int, name, signature, superName string, interfaces []string) {
	for _, callback := range e.OnVisit {
		callback(version, access, name, signature, superName, interfaces)
	}
}

func (e EventVisitor) VisitSource(source, debug string) {

}

func (e EventVisitor) VisitModule(name string, access, version int) {

}

func (e EventVisitor) VisitOuterClass(owner, name, descriptor string) {

}

func (e EventVisitor) VisitAnnotation(descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (e EventVisitor) VisitTypeAnnotation(typeRef, typePath int, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (e EventVisitor) VisitAttribute(attribute *asm.Attribute) {

}

func (e EventVisitor) VisitInnerClass(name, outerName, innerName string, access int) {

}

func (e EventVisitor) VisitField(access int, name, descriptor, signature string, value interface{}) {

}

func (e EventVisitor) VisitMethod(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor {
	return nil
}

func (e EventVisitor) VisitEnd() {
	for _, callback := range e.OnVisitEnd {
		callback()
	}
}
