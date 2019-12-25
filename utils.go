package relaysdk

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func ByteToInt(b []byte) int {
	mask := 0xff
	temp := 0
	n := 0
	for i := 0; i < len(b); i++ {
		n <<= 8
		temp = int(b[i]) & mask
		n |= temp
	}
	return n
}

//整形转换成字节
func IntToBytes(n int64, b byte) ([]byte, error) {
	switch b {
	case 1:
		tmp := int8(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		err := binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		if err != nil {
			return bytesBuffer.Bytes(), err
		}
		return bytesBuffer.Bytes(), nil
	case 2:
		tmp := int16(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		err := binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		if err != nil {
			return bytesBuffer.Bytes(), err
		}
		return bytesBuffer.Bytes(), nil
	case 3, 4:
		tmp := int32(n)
		bytesBuffer := bytes.NewBuffer([]byte{})
		err := binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		if err != nil {
			return bytesBuffer.Bytes(), err
		}
		return bytesBuffer.Bytes(), nil
	case 5, 6, 7, 8:
		tmp := n
		bytesBuffer := bytes.NewBuffer([]byte{})
		err := binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		if err != nil {
			return bytesBuffer.Bytes(), err
		}
		return bytesBuffer.Bytes(), nil
	}
	return nil, fmt.Errorf("IntToBytes b param is invaild")
}
