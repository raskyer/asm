# ASM

[IN PROGRESS]

This repository is a Golang port of the famous ASM java library for JVM bytecode reading and manipulation.
http://asm.ow2.org/. You can find the original source code at this link : https://gitlab.ow2.org/asm/asm.

## Goal
There are multiple goals for this port. The first one is comparing performance between the actual Java implementation and with this port.
Second reason could be to provide different tools for JVM manipulation through Golang, as, for example, a simple conversion of JVM Bytecode to Go.

## Progress

| CLASS | PROGRESS % |
| ----- | ---------- |
| ClassReader | 70% |
| ClassVisitor | 100% |
| MethodVisitor | 100% |
| FieldVisitor | 100% |
| ModuleVisitor | 100% |
| AnnotationVisitor | 100% |
| Handle | 100% |
| Attribute | ? |
| Context | ? |
| Label | 90% |
| TypeReference | ? |
| Opcodes | ? |
| Symbol | ? |
| TypePath | 0% |
| EDGE | 80% |
