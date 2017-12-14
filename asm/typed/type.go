package typed

const (
	VOID                  = 0
	BOOLEAN               = 1
	CHAR                  = 2
	BYTE                  = 3
	SHORT                 = 4
	INT                   = 5
	FLOAT                 = 6
	LONG                  = 7
	DOUBLE                = 8
	ARRAY                 = 9
	OBJECT                = 10
	METHOD                = 11
	INTERNAL              = 12
)

var	PRIMITIVE_DESCRIPTORS = []rune{'V', 'Z', 'C', 'B', 'S', 'I', 'F', 'J', 'D'}
 