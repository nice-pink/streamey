package audio

import (
	"encoding/hex"
	"testing"
)

func TestAacStartWithSync(t *testing.T) {
	hexHeader := "FFF14EA"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithAdtsSync(data, 0)

	if !isSync {
		t.Error("Should start with sync but does not.")
	}
}

func TestAacStartWithSyncFailed(t *testing.T) {
	hexHeader := "EFFAE10C"
	data, _ := hex.DecodeString(hexHeader)

	isSync := StartsWithAdtsSync(data, 0)

	if isSync {
		t.Error("Should not start with sync but does.")
	}
}

func TestAacGetHeader(t *testing.T) {
	hexHeader := "FFF1CEA8200001"
	bytes, _ := hex.DecodeString(hexHeader)
	header := GetAdtsHeader(bytes, 0)

	if header.MpegVersion != AacMpegVersion4 {
		t.Errorf("Version: got %d != want %d", header.MpegVersion, AacMpegVersion4)
	}
	if header.Layer != 0 {
		t.Errorf("Layer: got %d != want %d", header.Layer, 0)
	}
	if header.Protected != false {
		t.Error("Protect: got true != want false")
	}
	if header.Profile != Mpeg4AudioObjectAacLC {
		t.Errorf("Profile: got %d != want %d", header.Profile, Mpeg4AudioObjectAacLC)
	}
	if header.SampleRate != 48000 {
		t.Errorf("Sample rate: got %d != want %d", header.SampleRate, 48000)
	}
	if header.Private != true {
		t.Error("Private: got false != want true")
	}
	if header.ChannelConfig != Mpeg4ChannelConfigStereo {
		t.Errorf("ChannelConfig: got %d != want %d", header.ChannelConfig, Mpeg4ChannelConfigStereo)
	}
	if header.Original != true {
		t.Error("Original: got false != want true")
	}
	if header.Home != false {
		t.Error("Home: got false != want true")
	}
	if header.Copyright != true {
		t.Error("Copyright: got true != want false")
	}
	if header.CopyrightStart != false {
		t.Error("CopyrightStart: got false != want true")
	}
	if header.Size != 256 {
		t.Errorf("Size: got %d != want %d", header.Size, 256)
	}
}

func TestAacSetPrivate(t *testing.T) {
	// get
	hexHeader := "FFF1CEA8200001"
	bytes, _ := hex.DecodeString(hexHeader)
	header := GetAdtsHeader(bytes, 0)
	if header.Private != true {
		t.Error("Private: got false != want true")
	}
	SetAdtsUnPrivate(bytes, 0)
	unprivateHeader := GetAdtsHeader(bytes, 0)
	if unprivateHeader.Private != false {
		t.Error("Private: got true != want false")
	}
	SetAdtsPrivate(bytes, 0)
	privateHeader := GetAdtsHeader(bytes, 0)
	if privateHeader.Private != true {
		t.Error("Private: got false != want true")
	}
}
