package audio

import (
	"fmt"
	"strconv"

	"github.com/nice-pink/streamey/pkg/util"
)

type AudioType int

const (
	AudioTypeUnknown AudioType = iota
	AudioTypeMp3
	AudioTypeAAC
)

// encoding

type Encoding struct {
	ContainerName string // e.g. wav, mpeg, aac, ...
	CodecName     string
	SampleRate    int
	Bitrate       int
	Stereo        bool
	Profile       string // aac profile; could also be pcm for wav
	FrameSize     int    // samples per frame
}

func IsEncoded(e Encoding) bool {
	return e.CodecName != ""
}

func (e Encoding) Print() {
	fmt.Println("Container: ", e.ContainerName)
	fmt.Println("Codec: ", e.CodecName)
	fmt.Println("Profile: ", e.Profile)
	fmt.Println("Sample rate: ", strconv.Itoa(e.SampleRate))
	fmt.Println("Bitrate: ", strconv.Itoa(e.Bitrate))
	fmt.Println("Frame size: ", strconv.Itoa(e.FrameSize))
	fmt.Println("Is Stereo: ", util.YesNo(e.Stereo))
}

// unit infos

type UnitInfo struct {
	Index uint64
	Size  int
}

type AudioInfos struct {
	Units                []UnitInfo
	IsCBR                bool
	IsSampleRateConstant bool
	Encoding             Encoding
	TagSize              int64
	FirstFrameIndex      int64
}

func (a AudioInfos) Print() {
	fmt.Println("Encoding:")
	a.Encoding.Print()
	fmt.Println("Is CBR: ", util.YesNo(a.IsCBR))
	fmt.Println("Is sample rate constant: ", util.YesNo(a.IsSampleRateConstant))
	fmt.Println("Unit count: ", strconv.Itoa(len(a.Units)))
	fmt.Println("Tag size: ", a.TagSize)
}

// expectations

type Expectations struct {
	Encoding Encoding
	CBR      bool
}
