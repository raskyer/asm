package asm

// CLASS_TYPE_PARAMETER the sort of type references that target a type parameter of a generic class. See {@link #getSort}.
const CLASS_TYPE_PARAMETER = 0x00

// METHOD_TYPE_PARAMETER the sort of type references that target a type parameter of a generic method. See {@link #getSort}.
const METHOD_TYPE_PARAMETER = 0x01

// CLASS_EXTENDS The sort of type references that target the super class of a class or one of the interfaces it implements. See {@link #getSort}.
const CLASS_EXTENDS = 0x10

// CLASS_TYPE_PARAMETER_BOUND the sort of type references that target a bound of a type parameter of a generic class. See {@link #getSort}.
const CLASS_TYPE_PARAMETER_BOUND = 0x11

// METHOD_TYPE_PARAMETER_BOUND the sort of type references that target a bound of a type parameter of a generic method. See {@link #getSort}.
const METHOD_TYPE_PARAMETER_BOUND = 0x12

// FIELD the sort of type references that target the type of a field. See {@link #getSort}.
const FIELD = 0x13

// METHOD_RETURN the sort of type references that target the return type of a method. See {@link #getSort}.
const METHOD_RETURN = 0x14

// METHOD_RECEIVER the sort of type references that target the receiver type of a method. See {@link #getSort}.
const METHOD_RECEIVER = 0x15

/**
* The sort of type references that target the type of a formal parameter of a method. See {@link
* #getSort}.
 */
const METHOD_FORMAL_PARAMETER = 0x16

/**
* The sort of type references that target the type of an exception declared in the throws clause
* of a method. See {@link #getSort}.
 */
const THROWS = 0x17

/**
* The sort of type references that target the type of a local variable in a method. See {@link
* #getSort}.
 */
const LOCAL_VARIABLE = 0x40

/**
* The sort of type references that target the type of a resource variable in a method. See {@link
* #getSort}.
 */
const RESOURCE_VARIABLE = 0x41

/**
* The sort of type references that target the type of the exception of a 'catch' clause in a
* method. See {@link #getSort}.
 */
const EXCEPTION_PARAMETER = 0x42

/**
* The sort of type references that target the type declared in an 'instanceof' instruction. See
* {@link #getSort}.
 */
const INSTANCEOF = 0x43

/**
* The sort of type references that target the type of the object created by a 'new' instruction.
* See {@link #getSort}.
 */
const NEW = 0x44

/**
* The sort of type references that target the receiver type of a constructor reference. See
* {@link #getSort}.
 */
const CONSTRUCTOR_REFERENCE = 0x45

/**
* The sort of type references that target the receiver type of a method reference. See {@link
* #getSort}.
 */
const METHOD_REFERENCE = 0x46

/**
* The sort of type references that target the type declared in an explicit or implicit cast
* instruction. See {@link #getSort}.
 */
const CAST = 0x47

/**
* The sort of type references that target a type parameter of a generic constructor in a
* constructor call. See {@link #getSort}.
 */
const CONSTRUCTOR_INVOCATION_TYPE_ARGUMENT = 0x48

/**
* The sort of type references that target a type parameter of a generic method in a method call.
* See {@link #getSort}.
 */
const METHOD_INVOCATION_TYPE_ARGUMENT = 0x49

/**
* The sort of type references that target a type parameter of a generic constructor in a
* constructor reference. See {@link #getSort}.
 */
const CONSTRUCTOR_REFERENCE_TYPE_ARGUMENT = 0x4A

/**
* The sort of type references that target a type parameter of a generic method in a method
* reference. See {@link #getSort}.
 */
const METHOD_REFERENCE_TYPE_ARGUMENT = 0x4B
