//nolint:mnd,gosec
package ds

import (
	"encoding/binary"
	"strconv"

	"github.com/google/uuid"
)

const DefaultHash string = `0`

func IsDefaultHash(h string) bool {
	return h == DefaultHash
}

func Int32ToByte(i int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))

	return buf
}

func ByteToInt32(b []byte) int32 {
	if len(b) != 4 {
		panic("illegal buf size")
	}

	r := binary.LittleEndian.Uint32(b)

	return int32(r)
}

func Int64ToByte(i int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(i))

	return buf
}

func Uint32ToByte(i uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, i)

	return buf
}

func Uint64ToByte(i uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)

	return buf
}

func BoolToByte(b bool) []byte {
	if b {
		return []byte{0x1}
	}

	return []byte{0x0}
}

func Int32ToStr(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

func StrToInt32(s string) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1
	}

	return int32(i)
}

func UUIDToByte(s string) []byte {
	id, _ := uuid.Parse(s)
	buf, _ := id.MarshalBinary()

	return buf
}
