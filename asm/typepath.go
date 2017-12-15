package asm

type TypePath struct {
	typePathContainer []byte
	typePathOffset    int
}

func NewTypePath(b []byte, offset int) *TypePath {
	return &TypePath{
		b,
		offset,
	}
}

func NewTypePathFromString(typePath string) *TypePath {
	if typePath == "" || len(typePath) == 0 {
		return nil
	}

	//typePathLength := len(typePath)
	/*
			ByteVector output = new ByteVector(typePathLength);
		    output.putByte(0);
		    for (int i = 0; i < typePathLength; ) {
		      char c = typePath.charAt(i++);
		      if (c == '[') {
		        output.put11(ARRAY_ELEMENT, 0);
		      } else if (c == '.') {
		        output.put11(INNER_TYPE, 0);
		      } else if (c == '*') {
		        output.put11(WILDCARD_BOUND, 0);
		      } else if (c >= '0' && c <= '9') {
		        int typeArg = c - '0';
		        while (i < typePathLength && (c = typePath.charAt(i)) >= '0' && c <= '9') {
		          typeArg = typeArg * 10 + c - '0';
		          i += 1;
		        }
		        if (i < typePathLength && typePath.charAt(i) == ';') {
		          i += 1;
		        }
		        output.put11(TYPE_ARGUMENT, typeArg);
		      }
		    }
		    output.data[0] = (byte) (output.length / 2);
	*/
	return &TypePath{}
}

func (t TypePath) getLength() int {
	return int(t.typePathContainer[t.typePathOffset])
}

func (t TypePath) getStep(index int) int {
	return int(t.typePathContainer[t.typePathOffset+2*index+1])
}

func (t TypePath) getStepArgument(index int) int {
	return int(t.typePathContainer[t.typePathOffset+2*index+2])
}
