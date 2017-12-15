package asm

const (
	JUMP      = 0
	EXCEPTION = 0x7FFFFFFF
)

// Edge An edge in the control flow graph of a method. Each node of this graph is a basic block,
// represented with the Label corresponding to its first instruction. Each edge goes from one node
// to another, i.e. from one basic block to another (called the predecessor and successor blocks,
// respectively). An edge corresponds either to a jump or ret instruction or to an exception
// handler.
type Edge struct {
	info      int
	successor *Label
	nextEdge  *Edge
}

func NewEdge(info int, successor *Label, nextEdge *Edge) *Edge {
	return &Edge{info, successor, nextEdge}
}
