package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nice-pink/streamey/pkg/metadata"
)

var (
	// data
	currentOffset int64 = 0
	currentData   []byte
	// tag
	hasTag  bool  = false
	tagSize int64 = 0
	// audio
	foundFirstFrame bool = false
)

// get type

func GuessAudioType(path string) AudioType {
	ext := filepath.Ext(path)
	if ext == "" {
		return AudioTypeUnknown
	}

	// get type by extension
	ext = strings.TrimPrefix(ext, ".")
	if strings.ToUpper(ext) == "MP3" {
		return AudioTypeMp3
	}
	if strings.ToUpper(ext) == "AAC" {
		return AudioTypeAAC
	}
	return AudioTypeUnknown
}

func GetFirstFrameIndex(data []byte, offset uint64, audioTypeGuessed AudioType) uint64 {
	if audioTypeGuessed == AudioTypeMp3 {
		return uint64(GetNextFrameIndexMpeg(data, offset))
	}
	return offset
}

func GetAudioType(data []byte) AudioType {
	// get audio type
	if StartsWithMpegSync(data) {
		return AudioTypeMp3
	}
	return AudioTypeUnknown
}

// parse

func Parse(data []byte, filepath string) {
	// skip tag
	tagSize = metadata.GetTagSize(data)
	if tagSize < 0 {
		fmt.Println("Error: Tag size could not be evaluated.")
		tagSize = 0
	} else if tagSize > 0 {
		fmt.Println("Tag size:", tagSize)
	}
	hasTag = tagSize > 0

	// parse audio
	var audioStart int64 = 0
	if tagSize > 0 {
		audioStart = tagSize - 1
	}
	audioTypeGuessed := GuessAudioType(filepath)
	firstFrameIndex := GetFirstFrameIndex(data, uint64(audioStart), audioTypeGuessed)
	foundFirstFrame = firstFrameIndex > uint64(audioStart)

	// get audio
	audioType := GetAudioType(data[firstFrameIndex:])
	if audioType == AudioTypeUnknown {
		fmt.Println("Unknown audio type.")
		return
	}

	var audioInfo AudioInfos
	if audioType == AudioTypeMp3 {
		audioInfo = ParseMp3(data[firstFrameIndex:])

	} else if audioType == AudioTypeAAC {
		fmt.Println("Not jet implemented!")
	}

	fmt.Println()
	audioInfo.ContainsTag = hasTag
	audioInfo.Print()

	// log
	dataSize := len(data)
	fmt.Println()
	fmt.Println("---")
	fmt.Println("File path:", filepath)
	fmt.Println("File size:", dataSize)
	fmt.Println("Tag size:", tagSize)
	fmt.Println("Skipped bytes to first frame:", firstFrameIndex-uint64(tagSize))
	fmt.Println("Audio size:", dataSize-int(firstFrameIndex))
}

func MakeFirstFramePrivate(data []byte, audioType AudioType) {
	if audioType == AudioTypeMp3 {
		SetMpegPrivate(data)
	}
}

func ParseMp3(data []byte) AudioInfos {
	// get frame infos
	header := GetMpegHeader(data)
	if !header.IsValid() {
		fmt.Println("Error: Header is not valid")
		header.Print(false)
		os.Exit(2)
	}

	encoding := GetMp3Encoding(header)
	audioInfo := GetAudioInfos(data, 0, encoding, true)
	return audioInfo
}
