package game

import (
	"bytes"
	"encoding/binary"
)

type Reader struct {
	data []byte
	off  int
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

func (r *Reader) Remaining() int {
	if r.off >= len(r.data) {
		return 0
	}
	return len(r.data) - r.off
}

func (r *Reader) ReadUint32BE() uint32 {
	if r.Remaining() < 4 {
		return 0
	}
	v := binary.BigEndian.Uint32(r.data[r.off:])
	r.off += 4
	return v
}

func (r *Reader) ReadUint16BE() uint16 {
	if r.Remaining() < 2 {
		return 0
	}
	v := binary.BigEndian.Uint16(r.data[r.off:])
	r.off += 2
	return v
}

func (r *Reader) ReadBytes(n int) []byte {
	if n <= 0 || r.Remaining() <= 0 {
		return nil
	}
	if r.Remaining() < n {
		n = r.Remaining()
	}
	out := r.data[r.off : r.off+n]
	r.off += n
	return out
}

func (r *Reader) ReadFixedString(n int) string {
	b := r.ReadBytes(n)
	if len(b) == 0 {
		return ""
	}
	if idx := bytes.IndexByte(b, 0); idx >= 0 {
		b = b[:idx]
	}
	return string(b)
}
