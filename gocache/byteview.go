package gocache

type ByteView struct{
	b []byte
}

func (v ByteView) Len() int{
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte{
	return cloneBytes(v.b)
}

func (v ByteView) String() string{
	return string(v.b)
}

//深拷贝防止被外部程序修改
func cloneBytes(b []byte) []byte{
	c := make([]byte,len(b))
	copy(c,b)
	return c
}


