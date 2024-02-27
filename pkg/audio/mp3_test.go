package audio

import (
	"encoding/hex"
	"testing"
)

func TestStartWithSync(t *testing.T) {
	hexHeader := "FFFAE10C"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithMpegSync(data)

	if !isSync {
		t.Error("Should start with sync but does not.")
	}
}

func TestStartWithSyncFailed(t *testing.T) {
	hexHeader := "EFFAE10C"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithMpegSync(data)

	if isSync {
		t.Error("Should not start with sync but does.")
	}
}

func TestGetHeader(t *testing.T) {
	hexHeader := "FFFAE10C"
	bytes, _ := hex.DecodeString(hexHeader)
	header := GetHeader(bytes)

	if header.MpegVersion != MpegVersion1 {
		t.Errorf("Version: got %d != want %d", header.MpegVersion, MpegVersion1)
	}
	if header.Layer != MpegLayer3 {
		t.Errorf("Layer: got %d != want %d", header.Layer, MpegLayer3)
	}
	if header.Protected != false {
		t.Error("Protect: got true != want false")
	}
	if header.Bitrate != 320 {
		t.Errorf("Bitrate: got %d != want %d", header.Bitrate, 320)
	}
	if header.SampleRate != 44100 {
		t.Errorf("Sample rate: got %d != want %d", header.SampleRate, 44100)
	}
	if header.Padding != false {
		t.Error("Padding: got true != want false")
	}
	if header.Private != true {
		t.Error("Private: got true != want false")
	}
}
