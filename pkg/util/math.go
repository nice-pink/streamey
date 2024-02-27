package util

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
)

// bit shifting

func BoolFromBit(data []byte, bitIndex int) bool {
	byteIndex := bitIndex / 8
	bytes := make([]byte, 1)
	_ = copy(bytes, data[byteIndex:byteIndex+1])

	// shift
	inByteIndex := bitIndex % 8
	if inByteIndex != 0 {
		ShiftLeft(bytes, inByteIndex)
	}
	ShiftLeft(bytes, -7)
	return int8(bytes[0]) == 1
}

func BitsFromBytes(data []byte, bitIndex int, bitCount int) []byte {
	byteIndex := bitIndex / 8
	byteIndexOffset := int(math.Ceil(float64(bitCount) / 8))
	bytes := make([]byte, byteIndexOffset)
	_ = copy(bytes, data[byteIndex:byteIndex+byteIndexOffset])

	// shift
	inByteIndex := bitIndex % 8
	if inByteIndex != 0 {
		ShiftLeft(bytes, inByteIndex)
	}
	ShiftLeft(bytes, -(8 - bitCount))
	return bytes
}

func BytesEqualHex(h string, compare []byte) bool {
	// decode
	value, err := hex.DecodeString(h)
	if err != nil {
		fmt.Println("Error: Can't decode hex string.", err)
		return false
	}

	// validate
	if len(value) > len(compare) {
		fmt.Println("Hex compare data too short.")
		return false
	}

	// compare
	for i, b := range value {
		if compare[i] != b {
			fmt.Println(int8(compare[i]) != int8(b))
			return false
		}
	}
	return true
}

func BytesEqualHexWithMask(h string, mask string, compare []byte) bool {
	// decode hex strings
	value, err := hex.DecodeString(h)
	if err != nil {
		fmt.Println("Error: Can't decode hex string.", err)
		return false
	}
	valueMask, err := hex.DecodeString(mask)
	if err != nil {
		fmt.Println("Error: Can't decode mask.", err)
		return false
	}

	// validate
	if len(value) != len(valueMask) || len(value) > len(compare) {
		fmt.Println("Hex compare data too short.")
		return false
	}

	// compare
	for i, b := range value {
		if compare[i]&valueMask[i] != b {
			fmt.Println(int8(compare[i]&valueMask[i]) != int8(b))
			return false
		}
	}
	return true
}

// unsynchsafe
func Unsynchsafe(value uint32) uint32 {
	var out uint32 = 0
	tmp, err := strconv.ParseInt("7F000000", 10, 32)
	if err != nil {
		return 0
	}
	var mask uint32 = uint32(tmp)
	for {
		if mask == 0 {
			break
		}
		out >>= 1
		out |= value & mask
		mask >>= 8
	}
	return out
}

// ShiftLeft performs a left bit shift operation on the provided bytes.
// If the bits count is negative, a right bit shift is performed.
func ShiftLeft(data []byte, bits int) {
	n := len(data)
	if bits < 0 {
		bits = -bits
		for i := n - 1; i > 0; i-- {
			data[i] = data[i]>>bits | data[i-1]<<(8-bits)
		}
		data[0] >>= bits
	} else {
		for i := 0; i < n-1; i++ {
			data[i] = data[i]<<bits | data[i+1]>>(8-bits)
		}
		data[n-1] <<= bits
	}
}
