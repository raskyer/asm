package constants

// ASM specific access flags.
// WARNING: the 16 least significant bits must NOT be used, to avoid conflicts with standard
// access flags, and also to make sure that these flags are automatically filtered out when
// written in class files (because access flags are stored using 16 bits only).
const ACC_CONSTRUCTOR = 0x40000 // method access flag.

// ASM specific stack map frame types, used in {@link ClassVisitor#visitFrame}.

/**
 * A frame inserted between already existing frames. This internal stack map frame type (in
 * addition to the ones declared in {@link Opcodes}) can only be used if the frame content can be
 * computed from the previous existing frame and from the instructions between this existing frame
 * and the inserted one, without any knowledge of the type hierarchy. This kind of frame is only
 * used when an unconditional jump is inserted in a method while expanding an ASM specific
 * instruction. Keep in sync with Opcodes.java.
 */
const F_INSERT = 256

// The JVM opcode values which are not part of the ASM public API.
// See https://docs.oracle.com/javase/specs/jvms/se9/html/jvms-6.html.
const LDC_W = 19
const LDC2_W = 20
const ILOAD_0 = 26
const ILOAD_1 = 27
const ILOAD_2 = 28
const ILOAD_3 = 29
const LLOAD_0 = 30
const LLOAD_1 = 31
const LLOAD_2 = 32
const LLOAD_3 = 33
const FLOAD_0 = 34
const FLOAD_1 = 35
const FLOAD_2 = 36
const FLOAD_3 = 37
const DLOAD_0 = 38
const DLOAD_1 = 39
const DLOAD_2 = 40
const DLOAD_3 = 41
const ALOAD_0 = 42
const ALOAD_1 = 43
const ALOAD_2 = 44
const ALOAD_3 = 45
const ISTORE_0 = 59
const ISTORE_1 = 60
const ISTORE_2 = 61
const ISTORE_3 = 62
const LSTORE_0 = 63
const LSTORE_1 = 64
const LSTORE_2 = 65
const LSTORE_3 = 66
const FSTORE_0 = 67
const FSTORE_1 = 68
const FSTORE_2 = 69
const FSTORE_3 = 70
const DSTORE_0 = 71
const DSTORE_1 = 72
const DSTORE_2 = 73
const DSTORE_3 = 74
const ASTORE_0 = 75
const ASTORE_1 = 76
const ASTORE_2 = 77
const ASTORE_3 = 78
const WIDE = 196
const GOTO_W = 200
const JSR_W = 201

// Constants to convert between normal and wide jump instructions.

// The delta between the GOTO_W and JSR_W opcodes and GOTO and JUMP.
const WIDE_JUMP_OPCODE_DELTA = GOTO_W - GOTO

// Constants to convert JVM opcodes to the equivalent ASM specific opcodes, and vice versa.

// The delta between the ASM_IFEQ, ..., ASM_IF_ACMPNE, ASM_GOTO and ASM_JSR opcodes
// and IFEQ, ..., IF_ACMPNE, GOTO and JSR.
const ASM_OPCODE_DELTA = 49

// The delta between the ASM_IFNULL and ASM_IFNONNULL opcodes and IFNULL and IFNONNULL.
const ASM_IFNULL_OPCODE_DELTA = 20

// ASM specific opcodes, used for long forward jump instructions.
const ASM_IFEQ = IFEQ + ASM_OPCODE_DELTA
const ASM_IFNE = IFNE + ASM_OPCODE_DELTA
const ASM_IFLT = IFLT + ASM_OPCODE_DELTA
const ASM_IFGE = IFGE + ASM_OPCODE_DELTA
const ASM_IFGT = IFGT + ASM_OPCODE_DELTA
const ASM_IFLE = IFLE + ASM_OPCODE_DELTA
const ASM_IF_ICMPEQ = IF_ICMPEQ + ASM_OPCODE_DELTA
const ASM_IF_ICMPNE = IF_ICMPNE + ASM_OPCODE_DELTA
const ASM_IF_ICMPLT = IF_ICMPLT + ASM_OPCODE_DELTA
const ASM_IF_ICMPGE = IF_ICMPGE + ASM_OPCODE_DELTA
const ASM_IF_ICMPGT = IF_ICMPGT + ASM_OPCODE_DELTA
const ASM_IF_ICMPLE = IF_ICMPLE + ASM_OPCODE_DELTA
const ASM_IF_ACMPEQ = IF_ACMPEQ + ASM_OPCODE_DELTA
const ASM_IF_ACMPNE = IF_ACMPNE + ASM_OPCODE_DELTA
const ASM_GOTO = GOTO + ASM_OPCODE_DELTA
const ASM_JSR = JSR + ASM_OPCODE_DELTA
const ASM_IFNULL = IFNULL + ASM_IFNULL_OPCODE_DELTA
const ASM_IFNONNULL = IFNONNULL + ASM_IFNULL_OPCODE_DELTA
const ASM_GOTO_W = 220
