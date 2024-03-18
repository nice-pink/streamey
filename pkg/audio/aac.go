package audio

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/nice-pink/streamey/pkg/util"
)

// https://www.p23.nl/projects/aac-header/
// https://hydrogenaud.io/index.php/topic,3026.0.html
// https://developer.apple.com/documentation/quicktime-file-format#//apple_ref/doc/uid/TP40000939-CH1-SW2

// version
type AacMpegVersion int

const (
	AacMpegVersion4 AacMpegVersion = iota
	AacMpegVersion2
)

// Profile
type Mpeg4AudioObject int // aac profile = Mpeg4AudioObject-1

const (
	Mpeg4AudioObjectReserved Mpeg4AudioObject = iota
	Mpeg4AudioObjectAacMain
	Mpeg4AudioObjectAacLC
	Mpeg4AudioObjectAacSSR
	Mpeg4AudioObjectAacLTP
	Mpeg4AudioObjectSBR
	Mpeg4AudioObjectAacScalable
)

// channel mode
type Mpeg4ChannelConfig int

const (
	Mpeg4ChannelConfigAOT Mpeg4ChannelConfig = iota
	Mpeg4ChannelConfig1CenterFront
	Mpeg4ChannelConfigStereo
	Mpeg4ChannelConfig3ChannelFront
	Mpeg4ChannelConfig3ChannelFront1CenterBack
	Mpeg4ChannelConfig5ChannelSurround
	Mpeg4ChannelConfig6ChannelSurroundLFE
	Mpeg4ChannelConfig8ChannelSurround
	Mpeg4ChannelConfigReserved // 8-15 reserved
)

// const
const (
	AdtsHeaderSizeMin      int    = 7 // if CRC
	AdtsHeaderSizeMax      int    = 9 // if CRC
	AdtsSyncWord           string = "FFF0"
	AdtsSyncWordComparator string = "FFF6"
)

var (
	adtsSampleRate = [16]int{
		96000,
		88200,
		64000,
		48000,
		44100,
		32000,
		24000,
		22050,
		16000,
		12000,
		11025,
		8000,
		7350,
		0,  // RESERVED
		0,  // RESERVED
		-1, // ESCAPE VALUE
	}
)

type AdtsHeader struct {
	Sync           string
	MpegVersion    AacMpegVersion
	Layer          int8
	Profile        Mpeg4AudioObject
	Protected      bool
	Bitrate        int
	SampleRate     int
	Private        bool
	ChannelConfig  Mpeg4ChannelConfig
	Copyright      bool
	CopyrightStart bool
	Original       bool
	Home           bool
	Size           int32
	BufferFullness int16
	FrameCount     uint8
	CRC            int16
	Index          int64
}

func GetAdtsHeader(data []byte, index int64) AdtsHeader {
	header := AdtsHeader{}
	header.Index = index

	// version
	version := util.BitsFromBytes(data, 12, 1)
	header.MpegVersion = AacMpegVersion(int8(version[0]))

	// layer
	layer := util.BitsFromBytes(data, 13, 2)
	header.Layer = int8(layer[0])

	// protect
	header.Protected = !util.BoolFromBit(data, 15)

	// profile
	profile := util.BitsFromBytes(data, 16, 2)
	header.Profile = GetAacProfile(int8(profile[0]))

	// sample rate
	sampleRate := util.BitsFromBytes(data, 18, 4)
	header.SampleRate = adtsSampleRate[int(sampleRate[0])]

	// private
	header.Private = util.BoolFromBit(data, 22)

	// channel mode
	channelMode := util.BitsFromBytes(data, 23, 3)
	header.ChannelConfig = GetMpeg4ChannelConfig(int8(channelMode[0]))

	// original
	header.Original = util.BoolFromBit(data, 26)

	// home
	header.Home = util.BoolFromBit(data, 27)

	// copyright
	header.Copyright = util.BoolFromBit(data, 28)
	header.CopyrightStart = util.BoolFromBit(data, 29)

	// frame length
	length := util.BitsFromBytes(data, 30, 13)
	header.Size = int32(binary.BigEndian.Uint16(length[:]))

	// buffer fullness
	fullness := util.BitsFromBytes(data, 43, 11)
	header.BufferFullness = int16(binary.BigEndian.Uint16(fullness[0:]))

	// frame count
	frameCount := util.BitsFromBytes(data, 54, 2)
	header.FrameCount = uint8(frameCount[0])

	if header.Protected {
		header.CRC = int16(binary.BigEndian.Uint16(data[7:9]))
	}

	return header
}

func StartsWithAdtsSync(data []byte, offset uint64) bool {
	return util.BytesEqualHexWithMask(AdtsSyncWord, AdtsSyncWordComparator, data[offset:offset+2])
}

func (h AdtsHeader) Print(extended bool) {
	fmt.Println("---------------")
	label := "Aac " + GetAacProfileString(h.Profile)
	fmt.Println(label)
	fmt.Println("Bitrate: " + GetBitrateString(h.Bitrate))
	fmt.Println("Sample Rate: " + strconv.Itoa(h.SampleRate))
	if extended {
		fmt.Println("Private: " + util.YesNo(h.Private))
		fmt.Println("Copyright: " + util.YesNo(h.Copyright))
		fmt.Println("Original: " + util.YesNo(h.Original))
		fmt.Println("Protected: " + util.YesNo(h.Protected))
		fmt.Println("Home: " + util.YesNo(h.Home))
	}
	fmt.Println("Size: " + strconv.Itoa(int(h.Size)))
	fmt.Println("Index: " + strconv.Itoa(int(h.Index)))
}

func GetAdtsEncoding(header AdtsHeader) Encoding {
	return Encoding{
		ContainerName: "m4a",
		CodecName:     "aac",
		Bitrate:       header.Bitrate,
		SampleRate:    header.SampleRate,
		FrameSize:     int(header.Size),
		IsStereo:      header.ChannelConfig == Mpeg4ChannelConfigStereo,
	}
}

func GetAacProfile(value int8) Mpeg4AudioObject {
	if value > 7 {
		return Mpeg4AudioObjectReserved
	}
	return Mpeg4AudioObject(value - 1)
}

func GetMpeg4ChannelConfig(value int8) Mpeg4ChannelConfig {
	if value > 7 {
		return Mpeg4ChannelConfigReserved
	}
	return Mpeg4ChannelConfig(value)
}

// parse

func GetAudioInfosAac(data []byte, offset uint64, encoding Encoding, includeUnitEncoding bool, printHeaders bool, verbose bool) AudioInfos {
	audioInfos := AudioInfos{Encoding: encoding, Units: []UnitInfo{}, IsCBR: true, IsSampleRateConstant: true}
	dataSize := len(data)

	var index uint64 = offset
	for {
		// exit?
		if index+uint64(AdtsHeaderSizeMax) > uint64(dataSize) {
			break
		}

		// find sync header
		if !StartsWithAdtsSync(data, index) {
			index++
			continue
		}

		// get frame size
		header := GetAdtsHeader(data[index:], int64(index))
		if header.Size <= 0 {
			index++
			continue
		}

		// validate
		if header.SampleRate != encoding.SampleRate {
			audioInfos.IsSampleRateConstant = false
		}
		// if header.Bitrate != encoding.Bitrate {
		// 	audioInfos.IsCBR = false
		// }

		// print frame headers
		if printHeaders {
			header.Print(verbose)
		}

		// exit if frame is not complete
		if index+uint64(header.Size) > uint64(dataSize) {
			break
		}

		// append unit
		unitInfo := UnitInfo{Index: index, Size: int(header.Size), IsPrivate: header.Private}
		if includeUnitEncoding {
			unitInfo.Encoding = GetAdtsEncoding(header)
		}
		// fmt.Println(header)
		audioInfos.Units = append(audioInfos.Units, unitInfo)
		index += uint64(header.Size)
	}

	return audioInfos
}

func GetNextAdtsHeader(data []byte, offset uint64) *AdtsHeader {
	dataSize := len(data)
	var index int64 = int64(offset)
	for {
		// exit?
		if index+int64(AdtsHeaderSizeMax) > int64(dataSize) {
			break
		}

		// find sync header
		if !StartsWithAdtsSync(data, uint64(index)) {
			index++
			continue
		}

		// get frame size
		header := GetAdtsHeader(data[index:], index)
		if !header.IsValid(false) {
			index++
			continue
		}

		// get frame size
		if header.Size <= 0 {
			index++
			continue
		}

		header.Print(true)
		return &header
	}
	return nil
}

func (header AdtsHeader) IsValid(verbose bool) bool {
	if header.SampleRate <= 0 {
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
	if header.Layer != 0 {
		if verbose {
			fmt.Println("Invalid layer")
		}
		return false
	}
	if header.Private {
		if verbose {
			fmt.Println("Invalid layer")
		}
		return false
	}

	return true
}

func SetAdtsPrivate(header []byte, offset uint64) {
	var mask uint8 = 2
	header[offset+2] = header[offset+2] | mask
}

func SetAdtsUnPrivate(header []byte, offset uint64) {
	var mask uint8 = 253
	header[offset+2] = header[offset+2] & mask
}

// labels

func GetAacProfileString(profile Mpeg4AudioObject) string {
	return ""
}
