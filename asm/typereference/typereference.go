package typereference

const (
	CLASS_TYPE_PARAMETER                 = 0x00
	METHOD_TYPE_PARAMETER                = 0x01
	CLASS_EXTENDS                        = 0x10
	CLASS_TYPE_PARAMETER_BOUND           = 0x11
	METHOD_TYPE_PARAMETER_BOUND          = 0x12
	FIELD                                = 0x13
	METHOD_RETURN                        = 0x14
	METHOD_RECEIVER                      = 0x15
	METHOD_FORMAL_PARAMETER              = 0x16
	THROWS                               = 0x17
	LOCAL_VARIABLE                       = 0x40
	RESOURCE_VARIABLE                    = 0x41
	EXCEPTION_PARAMETER                  = 0x42
	INSTANCEOF                           = 0x43
	NEW                                  = 0x44
	RUCTOR_REFERENCE                     = 0x45
	METHOD_REFERENCE                     = 0x46
	CAST                                 = 0x47
	CONSTRUCTOR_INVOCATION_TYPE_ARGUMENT = 0x48
	METHOD_INVOCATION_TYPE_ARGUMENT      = 0x49
	CONSTRUCTOR_REFERENCE_TYPE_ARGUMENT  = 0x4A
	METHOD_REFERENCE_TYPE_ARGUMENT       = 0x4B
)
