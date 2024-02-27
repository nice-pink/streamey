package util

import (
	"encoding/hex"
	"testing"
)

// bits from bytes

func TestBitsFromBytes(t *testing.T) {
	hexString_1 := "000000394c414d45"

	// no shift
	bytes, _ := hex.DecodeString(hexString_1)
	result := BitsFromBytes(bytes, 32, 8)
	got_1 := hex.EncodeToString(result)
	want_1 := "4c"
	if got_1 != want_1 {
		t.Errorf("1: got %q != want %q", got_1, want_1)
	}

	// more bytes
	bytes, _ = hex.DecodeString(hexString_1)
	result = BitsFromBytes(bytes, 40, 16)
	got_2 := hex.EncodeToString(result)
	want_2 := "4d00"
	if got_2 != want_2 {
		t.Errorf("2: got %q != want %q", got_2, want_2)
	}

	// in byte shift
	hexString_3 := "0300"
	bytes_3, _ := hex.DecodeString(hexString_3)
	result_3 := BitsFromBytes(bytes_3, 6, 2)
	got_3 := hex.EncodeToString(result_3)
	want_3 := "03"
	if got_3 != want_3 {
		t.Errorf("3: got %q != want %q", got_3, want_3)
	}

	// in byte shift
	hexString_4 := "FF5A"
	bytes_4, _ := hex.DecodeString(hexString_4)
	result_4 := BitsFromBytes(bytes_4, 5, 2)
	got_4 := hex.EncodeToString(result_4)
	want_4 := "03"
	if got_4 != want_4 {
		t.Errorf("4: got %q != want %q", got_4, want_4)
	}
}

// shift

func DecodeShift(t *testing.T, hexString string, shift int, msg string) []byte {
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		t.Error("Could not decode.", msg)
	}
	ShiftLeft(bytes, shift)
	return bytes
}

func TestShiftLeft(t *testing.T) {
	// left shift 1
	shift_1 := 1
	hexString_1 := "01"
	bytes_l1 := DecodeShift(t, hexString_1, shift_1, "Shift left by 1")
	got_shift_l1 := hex.EncodeToString(bytes_l1)
	want_shift_l1 := "02"
	if got_shift_l1 != want_shift_l1 {
		t.Errorf("Left shift 1: got %q != want %q", got_shift_l1, want_shift_l1)
	}

	// right shift 1
	bytes_r1 := DecodeShift(t, hexString_1, -shift_1, "Shift left by -1")
	got_shift_r1 := hex.EncodeToString(bytes_r1)
	want_shift_r1 := "00"
	if got_shift_r1 != want_shift_r1 {
		t.Errorf("Right shift 1: got %q != want %q", got_shift_r1, want_shift_r1)
	}

	// left shift 5
	shift_2 := 5
	hexString_2 := "0100"
	bytes_l5 := DecodeShift(t, hexString_2, shift_2, "Shift left by 5")
	got_shift_l5 := hex.EncodeToString(bytes_l5)
	want_shift_l5 := "2000"
	if got_shift_l5 != want_shift_l5 {
		t.Errorf("Left shift 5: got %q != want %q", got_shift_l5, want_shift_l5)
	}

	// right shift 1
	bytes_r5 := DecodeShift(t, hexString_2, -shift_2, "Shift left by -1")
	got_shift_r5 := hex.EncodeToString(bytes_r5)
	want_shift_r5 := "0008"
	if got_shift_r5 != want_shift_r5 {
		t.Errorf("Right shift 5: got %q != want %q", got_shift_r5, want_shift_r5)
	}
}
