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

// const
const (
	MpegHeaderSize  int    = 4
	MpegSyncTag     string = "FFE0"
	MpegLayersMax   int    = 4
	MpegVersionsMax int    = 4
)

var (
	mpegSampleRates = [MpegVersionsMax][3]int{
		{11025, 12000, 8000},  //MpegVersion2_5
		{0, 0, 0},             //MpegVersionReserved
		{22050, 24000, 16000}, //MpegVersion2
		{44100, 48000, 32000}, //MpegVersion1
	}
	mpegBitrates = [MpegVersionsMax][MpegLayersMax][15]int{
		{ // MpegVersion2_5
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},                       // MpegLayerReserved
			{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // MpegLayer3
			{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // MpegLayer2
			{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256}, // MpegLayer1
		},
		{ // MpegVersionReserved
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // MpegLayerReserved
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // MpegLayer3
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // MpegLayer2
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // MpegLayer1
		},
		{ // MpegVersion2
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},                       // MpegLayerReserved
			{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // MpegLayer3
			{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},      // MpegLayer2
			{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256}, // MpegLayer1
		},
		{ // MpegVersion1
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},                          // MpegLayerReserved
			{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320},     // MpegLayer3
			{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384},    // MpegLayer2
			{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448}, // MpegLayer1
		},
	}
	mpegSamples = [MpegVersionsMax][MpegLayersMax]int{
		{ // MpegVersion2_5
			0,
			576,
			1152,
			384,
		},
		{ // MpegVersionReserved
			0,
			0,
			0,
			0,
		},
		{ // MpegVersion2
			0,
			576,
			1152,
			384,
		},
		{ // MpegVersion1
			0,
			1152,
			1152,
			384,
		},
	}
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
	Index         int64
}

func GetMpegHeader(data []byte, index int64) MpegHeader {
	header := MpegHeader{}
	header.Index = index

	// version
	version := util.BitsFromBytes(data, 11, 2)
	header.MpegVersion = MpegVersion(int8(version[0]))

	// // layer
	layer := util.BitsFromBytes(data, 13, 2)
	header.Layer = MpegLayer(int8(layer[0]))

	// //protect
	header.Protected = util.BoolFromBit(data, 15)

	// bitrate
	bitrate := util.BitsFromBytes(data, 16, 4)
	header.Bitrate = GetMpegBitrate(int8(bitrate[0]), MpegLayer3, header.MpegVersion)

	// // sample rate
	sampleRate := util.BitsFromBytes(data, 20, 2)
	header.SampleRate = GetMpegSampleRate(int8(sampleRate[0]), header.MpegVersion)

	// // padding
	header.Padding = util.BoolFromBit(data, 22)

	// // private
	header.Private = util.BoolFromBit(data, 23)

	// // channel mode
	channelMode := util.BitsFromBytes(data, 24, 2)
	header.ChannelMode = ChannelMode(int8(channelMode[0]))

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
	header.Size = GetMpegFrameSize(data, header, 0, GetMpegFrameSizeSamples(header.MpegVersion, header.Layer))

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
		FrameSize:     GetMpegFrameSizeSamples(header.MpegVersion, header.Layer),
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
		header := GetMpegHeader(data[index:], int64(index))
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

	// is valid
	var bytesPerFrame int
	if frameSizeSamples == 0 {
		bytesPerFrame = 144
	} else {
		bytesPerFrame = frameSizeSamples / 8
	}

	padding := 8
	packet := float64(bytesPerFrame) * float64(header.Bitrate*1000) / float64(header.SampleRate+padding)
	return int(packet)
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
		header := GetMpegHeader(data[index:], index)
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

func SetMpegPrivate(header []byte, offset uint64) {
	var mask uint8 = 1
	header[offset+2] = header[offset+2] | mask
}

func SetMpegUnPrivate(header []byte, offset uint64) {
	var mask uint8 = 14
	header[offset+2] = header[offset+2] & mask
}

//

func GetMpegSampleRate(value int8, version MpegVersion) int {
	if value > 2 {
		return -1
	}
	return mpegSampleRates[version][value]
}

func GetMpegBitrate(value int8, layer MpegLayer, version MpegVersion) int {
	if value == 0 || layer == MpegLayerReserved || version == MpegVersionReserved {
		return 0
	}
	return mpegBitrates[version][layer][value]
}

func GetMpegFrameSizeSamples(version MpegVersion, layer MpegLayer) int {
	return mpegSamples[version][layer]
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
