package audio

import (
	"fmt"
	"strconv"

	"github.com/nice-pink/streamey/pkg/util"
)

// version
type MpegVersion int

const (
	MpegVersion2_5 MpegVersion = iota
	MpegVersionReserved
	MpegVersion2
	MpegVersion1
)

// layer
type MpegLayer int

const (
	MpegLayerReserved MpegLayer = iota
	MpegLayer3
	MpegLayer2
	MpegLayer1
)

// channel mode
type ChannelMode int

const (
	Stereo ChannelMode = iota
	JoinedStereo
	DualChannel
	Mono
)

const (
	MpegHeaderSize int    = 4
	MpegSyncTag    string = "FFE0"
)

// header
type MpegHeader struct {
	Sync          string
	MpegVersion   MpegVersion
	Layer         MpegLayer
	Protected     bool
	Bitrate       int
	SampleRate    int
	Padding       bool
	Private       bool
	ChannelMode   ChannelMode
	ModeExtension int8
	Copyright     bool
	Original      bool
	Emphasis      int8
	Size          int
}

func GetMpegHeader(data []byte) MpegHeader {
	header := MpegHeader{}

	// version
	version := util.BitsFromBytes(data, 11, 2)
	header.MpegVersion = GetMpegVersion(int8(version[0]))

	// // layer
	layer := util.BitsFromBytes(data, 13, 2)
	header.Layer = GetMpegLayer(int8(layer[0]))

	// //protect
	header.Protected = util.BoolFromBit(data, 15)

	// bitrate
	bitrate := util.BitsFromBytes(data, 16, 4)
	header.Bitrate = GetMpegBitrate(bitrate, MpegLayer3, header.MpegVersion)

	// // sample rate
	sampleRate := util.BitsFromBytes(data, 20, 2)
	header.SampleRate = GetMpegSampleRate(int8(sampleRate[0]), header.MpegVersion)

	// // padding
	header.Padding = util.BoolFromBit(data, 22)

	// // private
	header.Private = util.BoolFromBit(data, 23)

	// // channel mode
	channelMode := util.BitsFromBytes(data, 24, 2)
	header.ChannelMode = GetMpegChannelMode(int8(channelMode[0]))

	// // mode extension
	modeExtension := util.BitsFromBytes(data, 26, 2)
	header.ModeExtension = int8(modeExtension[0])

	// copyright
	header.Copyright = util.BoolFromBit(data, 28)

	// original
	header.Original = util.BoolFromBit(data, 29)

	// emphasis
	emphasis := util.BitsFromBytes(data, 30, 2)
	header.Emphasis = int8(emphasis[0])

	// size
	header.Size = GetMpegFrameSize(data, header, 0, GetMpegFrameSizeSamples())

	return header
}

func StartsWithMpegSync(data []byte) bool {
	return util.BytesEqualHexWithMask(MpegSyncTag, MpegSyncTag, data)
}

func (h MpegHeader) Print(extended bool) {
	fmt.Println("---------------")
	label := "Mp3 " + GetMpegVersionString(h.MpegVersion) + "-" + GetMpegLayerString(h.Layer)
	fmt.Println(label)
	fmt.Println("Bitrate: " + GetBitrateString(h.Bitrate))
	fmt.Println("Sample Rate: " + strconv.Itoa(h.SampleRate))
	if extended {
		fmt.Println("Private: " + util.YesNo(h.Private))
		fmt.Println("Copyright: " + util.YesNo(h.Copyright))
		fmt.Println("Original: " + util.YesNo(h.Original))
		fmt.Println("Protected: " + util.YesNo(h.Protected))
	}
	fmt.Println("Size: " + strconv.Itoa(h.Size))
}

//

func GetMpegEncoding(header MpegHeader) Encoding {
	return Encoding{
		ContainerName: "mp3",
		CodecName:     "mp3",
		Bitrate:       header.Bitrate,
		SampleRate:    header.SampleRate,
		FrameSize:     GetMpegFrameSizeSamples(),
		IsStereo:      header.ChannelMode < 3,
	}
}

func GetAudioInfosMpeg(data []byte, offset uint64, encoding Encoding, includeUnitEncoding bool, printHeaders bool, verbose bool) AudioInfos {
	audioInfos := AudioInfos{Encoding: encoding, Units: []UnitInfo{}, IsCBR: true, IsSampleRateConstant: true}
	dataSize := len(data)

	var index uint64 = offset
	for {
		// exit?
		if index+uint64(MpegHeaderSize) > uint64(dataSize) {
			break
		}

		// find sync header
		if !StartsWithMpegSync(data[index:]) {
			index++
			continue
		}

		// get frame size
		header := GetMpegHeader(data[index:])
		frameSize := GetMpegFrameSize(data[index:], header, encoding.SampleRate, encoding.FrameSize)
		if frameSize <= 0 {
			index++
			continue
		}

		// validate
		if header.SampleRate != encoding.SampleRate {
			audioInfos.IsSampleRateConstant = false
		}
		if header.Bitrate != encoding.Bitrate {
			audioInfos.IsCBR = false
		}

		// print frame headers
		if printHeaders {
			header.Print(verbose)
		}

		// exit if frame is not complete
		if index+uint64(frameSize) > uint64(dataSize) {
			break
		}

		// append unit
		unitInfo := UnitInfo{Index: index, Size: frameSize, IsPrivate: header.Private}
		if includeUnitEncoding {
			unitInfo.Encoding = GetMpegEncoding(header)
		}
		// fmt.Println(header)
		audioInfos.Units = append(audioInfos.Units, unitInfo)
		index += uint64(frameSize)
	}

	return audioInfos
}

func GetMpegFrameSize(data []byte, header MpegHeader, requiredSampleRate int, frameSizeSamples int) int {
	if header.Layer == MpegLayerReserved || header.MpegVersion == MpegVersionReserved || header.Bitrate < 0 || header.SampleRate < 0 {
		return 0
	}

	if requiredSampleRate > 0 && header.SampleRate != requiredSampleRate {
		return -1
	}

	return GetMpegPacketSize(header.Bitrate, header.SampleRate, frameSizeSamples, 8)
}

func GetMpegPacketSize(bitrate int, sampleRate int, frameSizeSamples int, padding int) int {
	var bytesPerFrame int
	if frameSizeSamples == 0 {
		bytesPerFrame = 144
	} else {
		bytesPerFrame = frameSizeSamples / 8
	}

	packet := float64(bytesPerFrame) * float64(bitrate*1000) / float64(sampleRate+padding)

	return int(packet)
}

func GetNextFrameIndexMpeg(data []byte, offset uint64) int64 {
	dataSize := len(data)
	var index int64 = int64(offset)
	for {
		// exit?
		if index+int64(MpegHeaderSize) > int64(dataSize) {
			break
		}

		// find sync header
		if !StartsWithMpegSync(data[index:]) {
			index++
			continue
		}

		// get frame size
		header := GetMpegHeader(data[index:])
		frameSize := GetMpegFrameSize(data[index:], header, -1, 0)
		if frameSize <= 0 {
			index++
			continue
		}

		return index
	}
	return -1
}

func GetNextMpegHeader(data []byte, offset uint64) *MpegHeader {
	dataSize := len(data)
	var index int64 = int64(offset)
	for {
		// exit?
		if index+int64(MpegHeaderSize) > int64(dataSize) {
			break
		}

		// find sync header
		if !StartsWithMpegSync(data[index:]) {
			index++
			continue
		}

		// get frame size
		header := GetMpegHeader(data[index:])
		frameSize := GetMpegFrameSize(data[index:], header, -1, 0)
		if frameSize <= 0 {
			index++
			continue
		}

		return &header
	}
	return nil
}

//

func (header MpegHeader) IsValid(verbose bool) bool {
	if header.MpegVersion == MpegVersionReserved {
		if verbose {
			fmt.Println("Invalid version")
		}
		return false
	}
	if header.Layer == MpegLayerReserved {
		if verbose {
			fmt.Println("Invalid layer")
		}
		return false
	}
	if header.SampleRate == -1 {
		if verbose {
			fmt.Println("Invalid sample rate")
		}
		return false
	}
	if header.Bitrate == -1 {
		if verbose {
			fmt.Println("Invalid bitrate")
		}
		return false
	}

	return true
}

func GetMpegVersion(value int8) MpegVersion {
	if value == 0 {
		return MpegVersion2_5
	}
	if value == 2 {
		return MpegVersion2
	}
	if value == 3 {
		return MpegVersion1
	}
	return MpegVersionReserved
}

func GetMpegLayer(value int8) MpegLayer {
	if value == 3 {
		return MpegLayer1
	}
	if value == 2 {
		return MpegLayer2
	}
	if value == 1 {
		return MpegLayer3
	}
	return MpegLayerReserved
}

func GetMpegChannelMode(value int8) ChannelMode {
	if value == 0 {
		return Stereo
	}
	if value == 1 {
		return JoinedStereo
	}
	if value == 2 {
		return DualChannel
	}
	return Mono
}

func GetMpegSampleRate(value int8, mpegVersion MpegVersion) int {
	if value == 0 {
		if mpegVersion == MpegVersion1 {
			return 44100
		}
		if mpegVersion == MpegVersion2 {
			return 22500
		}
		if mpegVersion == MpegVersion2_5 {
			return 11025
		}
	}
	if value == 1 {
		if mpegVersion == MpegVersion1 {
			return 48000
		}
		if mpegVersion == MpegVersion2 {
			return 24000
		}
		if mpegVersion == MpegVersion2_5 {
			return 12000
		}
	}
	if value == 2 {
		if mpegVersion == MpegVersion1 {
			return 32000
		}
		if mpegVersion == MpegVersion2 {
			return 16000
		}
		if mpegVersion == MpegVersion2_5 {
			return 8000
		}
	}

	// reserved
	return -1
}

func GetMpegBitrate(bytes []byte, layer MpegLayer, version MpegVersion) int {
	value := int(bytes[0])
	if value == 0 {
		return 0
	}
	if value == 1 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 8
		}
		return 32
	}
	if value == 2 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 16
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 64
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 40
		}
		return 48
	}
	if value == 3 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 24
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 96
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 48
		}
		return 56
	}
	if value == 4 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 32
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 128
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 56
		}
		return 64
	}
	if value == 5 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 40
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 160
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 64
		}
		return 80
	}
	if value == 6 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 48
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 192
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 80
		}
		return 96
	}
	if value == 7 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 56
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 224
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 96
		}
		return 112
	}
	if value == 8 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 64
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 256
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 112
		}
		return 128
	}
	if value == 9 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 80
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 288
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 128
		}
		if version == MpegVersion1 && layer == MpegLayer2 {
			return 160
		}
		return 144
	}
	if value == 10 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 96
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 320
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 160
		}
		if version == MpegVersion1 && layer == MpegLayer2 {
			return 192
		}
		return 160
	}
	if value == 11 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 112
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 352
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 192
		}
		if version == MpegVersion1 && layer == MpegLayer2 {
			return 224
		}
		return 176
	}
	if value == 12 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 128
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 384
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 224
		}
		if version == MpegVersion1 && layer == MpegLayer2 {
			return 256
		}
		return 192
	}
	if value == 13 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 144
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 416
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 256
		}
		if version == MpegVersion1 && layer == MpegLayer2 {
			return 320
		}
		return 224
	}
	if value == 14 {
		if version == MpegVersion2 && (layer == MpegLayer2 || layer == MpegLayer3) {
			return 160
		}
		if version == MpegVersion1 && layer == MpegLayer1 {
			return 448
		}
		if version == MpegVersion1 && layer == MpegLayer3 {
			return 320
		}
		if version == MpegVersion1 && layer == MpegLayer2 {
			return 384
		}
		return 256
	}

	return -1
}

// labels

func GetMpegVersionString(version MpegVersion) string {
	if version == MpegVersion1 {
		return "V1"
	}
	if version == MpegVersion2 {
		return "V2"
	}
	if version == MpegVersion2_5 {
		return "V2.5"
	}
	return "reserved"
}

func GetMpegLayerString(layer MpegLayer) string {
	if layer == MpegLayer1 {
		return "L1"
	}
	if layer == MpegLayer2 {
		return "L2"
	}
	if layer == MpegLayer3 {
		return "L3"
	}
	return "reserved"
}

func GetBitrateString(bitrate int) string {
	return strconv.Itoa(bitrate) + "kBit"
}

func GetMpegChannelModeString(cm ChannelMode) string {
	if cm == Stereo {
		return "Stereo"
	}
	if cm == JoinedStereo {
		return "JoinedStereo"
	}
	if cm == DualChannel {
		return "DualChannel"
	}
	return "Mono"
}

func GetMpegFrameSizeSamples() int {
	return 1152
}

func SetMpegPrivate(header []byte, offset uint64) {
	var mask uint8 = 1
	header[offset+2] = header[offset+2] | mask
}

func SetMpegUnPrivate(header []byte, offset uint64) {
	var mask uint8 = 14
	header[offset+2] = header[offset+2] & mask
}
