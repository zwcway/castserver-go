package protocol

type Package struct {
	t   Type
	pos int
	d   []byte
}

func (p *Package) Read(s int) (ret []byte, err error) {
	if len(p.d) < p.pos+s {
		err = NewOverError()
		return
	}
	ret = p.d[p.pos : p.pos+s]
	p.pos += s
	return
}

func (p *Package) ReadUint8() (v uint8, err error) {
	if len(p.d) < p.pos+0 {
		err = NewOverError()
		return
	}

	v = p.d[p.pos]
	p.pos += 1

	return
}

func (p *Package) ReadUint16() (v uint16, err error) {
	if len(p.d) <= p.pos+1 {
		err = NewOverError()
		return
	}

	v = uint16(p.d[p.pos+1])<<8 | uint16(p.d[p.pos])
	p.pos += 2

	return
}

func (p *Package) ReadUint24() (v uint32, err error) {
	if len(p.d) < p.pos+2 {
		err = NewOverError()
		return
	}

	v = uint32(p.d[p.pos+2])<<16 | uint32(p.d[p.pos+1])<<8 | uint32(p.d[p.pos])
	p.pos += 3

	return
}

func (p *Package) ReadUint32() (v uint32, err error) {
	if len(p.d) < p.pos+3 {
		err = NewOverError()
		return
	}
	v = uint32(p.d[p.pos+3])<<24 | uint32(p.d[p.pos+2])<<16 | uint32(p.d[p.pos+1])<<8 | uint32(p.d[p.pos])
	p.pos += 4

	return
}

func (p *Package) Write(v []byte) error {
	if len(p.d) < p.pos+len(v) {
		return NewOverError()
	}
	for i, d := range v {
		p.d[p.pos+i] = d
	}
	return nil
}

func (p *Package) WriteUint8(v uint8) error {
	if len(p.d) < p.pos+0 {
		return NewOverError()
	}

	p.d[p.pos] = byte(v)
	p.pos += 1

	return nil
}

func (p *Package) WriteUint16(v uint16) error {
	if len(p.d) < p.pos+1 {
		return NewOverError()
	}

	p.d[p.pos] = byte(v)
	p.d[p.pos+1] = byte(v >> 8)
	p.pos += 2

	return nil
}

func (p *Package) WriteUint24(v uint32) error {
	if len(p.d) < p.pos+2 {
		return NewOverError()
	}

	p.d[p.pos] = byte(v)
	p.d[p.pos+1] = byte(v >> 8)
	p.d[p.pos+2] = byte(v >> 16)
	p.pos += 3

	return nil
}

func (p *Package) WriteUint32(v uint32) error {
	if len(p.d) < p.pos+3 {
		return NewOverError()
	}

	p.d[p.pos] = byte(v)
	p.d[p.pos+1] = byte(v >> 8)
	p.d[p.pos+2] = byte(v >> 16)
	p.d[p.pos+3] = byte(v >> 24)
	p.pos += 4

	return nil
}

func (p *Package) LastBytes(s int) []byte {
	return p.d[p.pos-s : p.pos]
}

func (p *Package) Bytes() []byte {
	return p.d[:p.pos]
}

func (p *Package) Size() int {
	return len(p.d)
}

func (p *Package) DataSize() int {
	return p.pos
}

func (p *Package) Type() Type {
	if len(p.d) > 0 {
		return Type(p.d[0])
	}
	return PT_UNKNOWN
}

func FromBinary(p []byte) *Package {
	return &Package{0, 0, p}
}

func NewPackage(size int) *Package {
	return &Package{0, 0, make([]byte, size)}
}
