package asm

type Handle struct {
	tag         int
	owner       string
	name        string
	descriptor  string
	isInterface bool
}
