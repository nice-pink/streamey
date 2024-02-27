package audio

import (
	"fmt"
	"os"

	"github.com/nice-pink/streamey/pkg/metadata"
)

func GetAudioType(data []byte) AudioType {
	// get audio type
	if StartsWithMpegSync(data) {
		return AudioTypeMp3
	}
	return AudioTypeUnknown
}

func Parse(data []byte) {
	// skip tag
	tagSize := metadata.GetTagSize(data)
	if tagSize < 0 {
		fmt.Println("Error: Tag size could not be evaluated.")
		tagSize = 0
	}
	hasTag := tagSize > 0

	// parse audio
	audioType := GetAudioType(data[tagSize:])
	if audioType == AudioTypeUnknown {
		return
	}

	if audioType == AudioTypeMp3 {
		ParseMp3(data[tagSize:], hasTag)
		return
	}

	if audioType == AudioTypeAAC {
		fmt.Println("Not jet implemented!")
		return
	}
}

func MakeFirstFramePrivate(data []byte, audioType AudioType) {
	if audioType == AudioTypeMp3 {
		SetMpegPrivate(data)
	}
}

func ParseMp3(data []byte, hasTag bool) {
	// skip metadata if any
	// metaSize := 0
	// if metadata.StartsWithId3V2Sync(data) {
	// 	metaSize = int(metadata.GetId3V2TagSize(data))
	// 	if metaSize > 0 {
	// 		data = data[metaSize:]
	// 	}
	// }

	// get frame infos
	header := GetMpegHeader(data)
	if !header.IsValid() {
		fmt.Println("Error: Header is not valid")
		header.Print(false)
		os.Exit(2)
	}

	encoding := GetMp3Encoding(header)
	audioInfo := GetAudioInfos(data, 0, encoding, true)
	audioInfo.ContainsTag = hasTag

	fmt.Println()
	audioInfo.Print()
}
