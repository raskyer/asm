package main

import (
	"github.com/leaklessgfy/asm/asm"
)

type SimpleVisitor struct {
	OnVisit       func(version, access int, name, signature, superName string, interfaces []string)
	OnVisitField  func(access int, name, descriptor, signature string, value interface{}) asm.FieldVisitor
	OnVisitMethod func(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor
	OnVisitEnd    func()
}

func (s SimpleVisitor) Visit(version, access int, name, signature, superName string, interfaces []string) {
	if s.OnVisit != nil {
		s.OnVisit(version, access, name, signature, superName, interfaces)
	}
}

func (s SimpleVisitor) VisitSource(source, debug string) {

}

func (s SimpleVisitor) VisitModule(name string, access int, version string) asm.ModuleVisitor {
	return nil
}

func (s SimpleVisitor) VisitOuterClass(owner, name, descriptor string) {

}

func (s SimpleVisitor) VisitAnnotation(descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (s SimpleVisitor) VisitTypeAnnotation(typeRef int, typePath *asm.TypePath, descriptor string, visible bool) asm.AnnotationVisitor {
	return nil
}

func (s SimpleVisitor) VisitAttribute(attribute *asm.Attribute) {

}

func (s SimpleVisitor) VisitInnerClass(name, outerName, innerName string, access int) {

}

func (s SimpleVisitor) VisitField(access int, name, descriptor, signature string, value interface{}) asm.FieldVisitor {
	if s.OnVisitField != nil {
		return s.OnVisitField(access, name, descriptor, signature, value)
	}
	return nil
}

func (s SimpleVisitor) VisitMethod(access int, name, descriptor, signature string, exceptions []string) asm.MethodVisitor {
	if s.OnVisitMethod != nil {
		return s.OnVisitMethod(access, name, descriptor, signature, exceptions)
	}
	return nil
}

func (s SimpleVisitor) VisitEnd() {
	if s.OnVisitEnd != nil {
		s.OnVisitEnd()
	}
}
