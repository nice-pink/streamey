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
	fullDataSize uint64 = 0
	currentData  []byte
	// tag
	hasTag          bool  = false
	skippedTag      bool  = false
	tagSize         int64 = 0
	currentTagIndex int64 = 0
	tagEnd          int64 = 0
	// audio
	foundFirstFrame        bool      = false
	skippedUntilFirstFrame uint64    = 0
	unitsTotal             uint64    = 0
	bytesTotal             uint64    = 0
	audioType              AudioType = AudioTypeUnknown
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
	skippedUntilFirstFrame = firstFrameIndex - uint64(tagSize)
	foundFirstFrame = firstFrameIndex > uint64(audioStart)

	// get audio
	audioType = GetAudioType(data[firstFrameIndex:])
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
	unitsTotal = uint64(len(audioInfo.Units))
	units := len(audioInfo.Units)
	bytesTotal = audioInfo.Units[units-1].Index + uint64(audioInfo.Units[units-1].Size)

	// log
	fullDataSize = uint64(len(data))
	LogResult(filepath)
	// dataSize := len(data)
	// fmt.Println()
	// fmt.Println("---")
	// fmt.Println("File path:", filepath)
	// fmt.Println("File size:", dataSize)
	// fmt.Println("Tag size:", tagSize)
	// fmt.Println("Skipped bytes to first frame:", firstFrameIndex-uint64(tagSize))
	// fmt.Println("Audio size:", dataSize-int(firstFrameIndex))
}

func ParseContinuous(data []byte, audioTypeGuessed AudioType) error {
	fullDataSize += uint64(len(data))
	currentData = append(currentData, data...)
	dataSize := len(currentData)
	var offset uint64 = 0

	// skip tag
	if !skippedTag && currentTagIndex <= tagSize {
		if tagSize == 0 {
			tagSize = metadata.GetTagSize(currentData)
			if tagSize < 0 {
				fmt.Println("Error: Tag size could not be evaluated.")
				tagSize = 0
			} else if tagSize > 0 {
				fmt.Println("Tag size:", tagSize)
			}
		}
		hasTag = tagSize > 0

		// skip tag
		if tagSize-currentTagIndex < int64(dataSize) {

			tagEnd = tagSize - currentTagIndex
			currentTagIndex = tagSize - 1
			skippedTag = true
			fmt.Println("Skipped tag at index:", currentTagIndex)
		} else {
			currentTagIndex += int64(dataSize)
			currentData = currentData[:0]
			return nil
		}
	}

	// parse audio
	offset = GetFirstFrameIndex(currentData, uint64(tagEnd), audioTypeGuessed)
	skippedUntilFirstFrame = offset - uint64(tagEnd)

	if !foundFirstFrame {
		// get audio
		audioType = GetAudioType(currentData[offset:])
		if audioType == AudioTypeUnknown {
			fmt.Println("Unknown audio type.")
			return nil
		}

		foundFirstFrame = offset >= uint64(tagEnd)
	}

	var audioInfo AudioInfos
	if audioType == AudioTypeMp3 {
		audioInfo = ParseMp3(currentData[offset:])
	} else if audioType == AudioTypeAAC {
		fmt.Println("Not jet implemented!")
	}

	// remove handled data from
	units := len(audioInfo.Units)
	if units > 0 {
		i := audioInfo.Units[units-1].Index + uint64(audioInfo.Units[units-1].Size)
		bytesTotal += i
		fmt.Println("Remove until", i)
		currentData = currentData[i:]
	} else {
		return nil
	}
	unitsTotal += uint64(units)

	// log infos
	fmt.Println()
	audioInfo.ContainsTag = hasTag
	audioInfo.Print()

	fmt.Println("total:", unitsTotal)

	LogResult("")

	return nil
}

func LogResult(filepath string) {
	// log
	fmt.Println()
	fmt.Println("---")
	if filepath != "" {
		fmt.Println("File path:", filepath)
	}
	fmt.Println("File size:", fullDataSize)
	fmt.Println("Tag size:", tagSize)
	fmt.Println("Skipped bytes to first frame:", skippedUntilFirstFrame)
	fmt.Println("Audio size:", fullDataSize-skippedUntilFirstFrame-uint64(tagSize))
	fmt.Println("Audio frames:", unitsTotal)
	fmt.Println("Bytes frames:", bytesTotal)
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
	audioInfo := GetAudioInfosMpeg(data, 0, encoding, true)
	return audioInfo
}
