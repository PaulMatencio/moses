package hexa

import (
	"encoding/binary"
	"encoding/hex"
)

func HexaByteToInt32(hexa string) (uint32, error) {
	var (
		data  uint32
		err   error
		bytea []byte
	)
	if bytea, err = hex.DecodeString(hexa); err == nil {
		data = binary.BigEndian.Uint32(bytea)
	}
	return data, err
}

func HexaByteToInt16(hexa string) (uint16, error) {
	var (
		data  uint16
		err   error
		bytea []byte
	)
	if bytea, err = hex.DecodeString(hexa); err == nil {
		data = binary.BigEndian.Uint16(bytea)
	}
	return data, err
}

func HexaByteToInt8(hexa string) (uint8, error) {
	var (
		data  uint8
		err   error
		bytea []byte
	)
	if bytea, err = hex.DecodeString(hexa); err == nil {
		data = bytea[0]
	}
	return data, err
}
