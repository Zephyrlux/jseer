package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	HeaderLen   = 17
	DefaultVers = 0x31
)

// BuildResponse builds a response packet for AS3 client.
// result must be 0 for success; non-zero may trigger client SocketError.
func BuildResponse(cmdID int32, userID uint32, result int32, body []byte) []byte {
	bodyLen := len(body)
	pkgLen := HeaderLen + bodyLen

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(pkgLen))
	buf.WriteByte(DefaultVers)
	binary.Write(buf, binary.BigEndian, uint32(cmdID))
	binary.Write(buf, binary.BigEndian, userID)
	binary.Write(buf, binary.BigEndian, uint32(result))
	buf.Write(body)
	return buf.Bytes()
}

// ParsePacket parses raw packet data into fields.
func ParsePacket(data []byte) (length int, version byte, cmdID int32, userID uint32, seqID int32, body []byte, err error) {
	if len(data) < HeaderLen {
		return 0, 0, 0, 0, 0, nil, fmt.Errorf("packet too short")
	}
	length = int(binary.BigEndian.Uint32(data[0:4]))
	version = data[4]
	cmdID = int32(binary.BigEndian.Uint32(data[5:9]))
	userID = binary.BigEndian.Uint32(data[9:13])
	seqID = int32(binary.BigEndian.Uint32(data[13:17]))
	if len(data) > HeaderLen {
		body = data[HeaderLen:]
	}
	return length, version, cmdID, userID, seqID, body, nil
}

// ReadUint32BE reads uint32 from data at offset.
func ReadUint32BE(data []byte, offset int) uint32 {
	if len(data) < offset+4 {
		return 0
	}
	return binary.BigEndian.Uint32(data[offset:])
}

// WriteUint32BE writes uint32 to buffer.
func WriteUint32BE(buf *bytes.Buffer, value uint32) {
	binary.Write(buf, binary.BigEndian, value)
}

// WriteUint16BE writes uint16 to buffer.
func WriteUint16BE(buf *bytes.Buffer, value uint16) {
	binary.Write(buf, binary.BigEndian, value)
}

// WriteByte writes a byte to buffer.
func WriteByte(buf *bytes.Buffer, value byte) {
	buf.WriteByte(value)
}

// WriteFixedString writes a fixed-length string with zero padding.
func WriteFixedString(buf *bytes.Buffer, s string, length int) {
	for i := 0; i < length; i++ {
		if i < len(s) {
			buf.WriteByte(s[i])
		} else {
			buf.WriteByte(0)
		}
	}
}
