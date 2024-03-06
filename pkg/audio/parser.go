package audio

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nice-pink/streamey/pkg/metadata"
)

var (
	// data
	fullDataSize uint64 = 0
	currentData  []byte
	// tag
	skippedTag bool = false
	// tagSize         int64 = 0
	currentTagIndex int64 = 0
	tagEnd          int64 = 0
	// audio
	encoding               Encoding
	foundEncoding          bool
	audioInfo              AudioInfos
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

func GetFirstFrameIndex(data []byte, offset uint64, audioTypeGuessed AudioType) int64 {
	if audioTypeGuessed == AudioTypeMp3 {
		return GetNextFrameIndexMpeg(data, offset)
	}
	return int64(offset)
}

func GetAudioType(data []byte) AudioType {
	// get audio type
	if StartsWithMpegSync(data) {
		return AudioTypeMp3
	}
	return AudioTypeUnknown
}

func GetAudioTypeFromCodecName(name string) AudioType {
	// get audio type
	if strings.ToUpper(name) == "MP3" {
		return AudioTypeMp3
	}
	return AudioTypeUnknown
}

// parse

func Parse(data []byte, filepath string, includeUnitEncoding bool, verbose bool, printLogs bool) *AudioInfos {
	// skip tag
	audioInfo.TagSize = metadata.GetTagSize(data)
	if audioInfo.TagSize < 0 {
		fmt.Println("Error: Tag size could not be evaluated.")
		audioInfo.TagSize = 0
	} else if audioInfo.TagSize > 0 {
		fmt.Println("Tag size:", audioInfo.TagSize)
	}

	// parse audio
	var audioStart int64 = 0
	if audioInfo.TagSize > 0 {
		audioStart = audioInfo.TagSize - 1
	}
	audioTypeGuessed := GuessAudioType(filepath)
	firstFrameIndex := GetFirstFrameIndex(data, uint64(audioStart), audioTypeGuessed)
	if firstFrameIndex < 0 {
		return nil
	}
	skippedUntilFirstFrame = uint64(firstFrameIndex - audioInfo.TagSize)
	foundFirstFrame = firstFrameIndex > audioStart

	// get audio
	audioType = GetAudioType(data[firstFrameIndex:])
	if audioType == AudioTypeUnknown {
		fmt.Println("Unknown audio type.")
		return nil
	}

	if audioType == AudioTypeMp3 {
		encoding = GetEncodingFromFirstMpegHeader(data, uint64(firstFrameIndex))
		audioInfo = ParseMp3(data[firstFrameIndex:], encoding, includeUnitEncoding, verbose)

	} else if audioType == AudioTypeAAC {
		fmt.Println("Not jet implemented!")
	}

	fmt.Println()
	audioInfo.FirstFrameIndex = audioInfo.TagSize + int64(skippedUntilFirstFrame)
	if printLogs {
		PrintAudioInfo()
	}

	unitsTotal = uint64(len(audioInfo.Units))
	bytesTotal = audioInfo.Units[unitsTotal-1].Index + uint64(audioInfo.Units[unitsTotal-1].Size)

	// log
	fullDataSize = uint64(len(data))
	if printLogs {
		LogParserResult(filepath)
	}

	return &audioInfo
}

func ParseBlockwise(data []byte, audioTypeGuessed AudioType, includeUnitEncoding bool, verbose bool, printLogs bool) (*AudioInfos, error) {
	fullDataSize += uint64(len(data))
	currentData = append(currentData, data...)
	dataSize := len(currentData)
	var offset int64 = 0

	// skip tag
	if !skippedTag && currentTagIndex <= audioInfo.TagSize {
		if audioInfo.TagSize == 0 {
			audioInfo.TagSize = metadata.GetTagSize(currentData)
			if audioInfo.TagSize < 0 {
				fmt.Println("Error: Tag size could not be evaluated.")
				audioInfo.TagSize = 0
			} else if audioInfo.TagSize > 0 {
				fmt.Println("Tag size:", audioInfo.TagSize)
			}
		}

		// skip tag
		if audioInfo.TagSize-currentTagIndex < int64(dataSize) {
			tagEnd = audioInfo.TagSize - currentTagIndex
			currentTagIndex = audioInfo.TagSize - 1
			skippedTag = true
			fmt.Println("Skipped tag at index:", currentTagIndex)
		} else {
			currentTagIndex += int64(dataSize)
			currentData = currentData[:0]
			return nil, nil
		}
	}

	// get audio offset
	offset = GetFirstFrameIndex(currentData, uint64(tagEnd), audioTypeGuessed)
	if offset < 0 {
		return nil, nil
	}

	if !foundFirstFrame {
		// get audio
		audioType = GetAudioType(currentData[offset:])
		if audioType == AudioTypeUnknown {
			fmt.Println("Unknown audio type.")
			return nil, nil
		}
		foundFirstFrame = offset >= tagEnd
		skippedUntilFirstFrame = uint64(offset - tagEnd)
		tagEnd = 0
	}

	// parse audio
	var audioInfoBlock AudioInfos
	if audioType == AudioTypeMp3 {
		if !foundEncoding {
			encoding = GetEncodingFromFirstMpegHeader(currentData, uint64(offset))
			foundEncoding = true
		}
		audioInfoBlock = ParseMp3(currentData[offset:], encoding, includeUnitEncoding, verbose)
	} else if audioType == AudioTypeAAC {
		fmt.Println("Not jet implemented!")
	}

	// remove handled data from
	units := len(audioInfoBlock.Units)
	if units > 0 {
		i := audioInfoBlock.Units[units-1].Index + uint64(audioInfoBlock.Units[units-1].Size) + uint64(offset)
		bytesTotal += i
		unitsTotal += uint64(units)
		currentData = currentData[i:]
	}

	// log infos
	if printLogs {
		fmt.Println()
		PrintAudioInfo()
		LogParserResult("")
	}

	return &audioInfoBlock, nil
}

func PrintAudioInfo() {
	audioInfo.Print()
}

func LogParserResult(filepath string) {
	// log
	fmt.Println()
	fmt.Println("---")
	if filepath != "" {
		fmt.Println("File path:", filepath)
	}
	fmt.Println("File size:", fullDataSize)
	fmt.Println("Tag size:", audioInfo.TagSize)
	fmt.Println("Skipped bytes to first frame:", skippedUntilFirstFrame)
	fmt.Println("Audio size:", fullDataSize-skippedUntilFirstFrame-uint64(audioInfo.TagSize))
	fmt.Println("Audio frames:", unitsTotal)
	fmt.Println("Bytes frames:", bytesTotal)
}

func MakeFirstFramePrivate(data []byte, offset uint64, audioType AudioType) {
	if audioType == AudioTypeMp3 {
		SetMpegPrivate(data, offset)
	}
}

// mpeg

func GetEncodingFromFirstMpegHeader(data []byte, offset uint64) Encoding {
	// get frame infos
	fmt.Println()
	fmt.Println("*************")
	fmt.Println("Inital header")
	fmt.Println("use encoding!")
	header := GetMpegHeader(data)
	if !header.IsValid(true) {
		fmt.Println("Error: Header is not valid")
		header.Print(false)

	}
	header.Print(true)
	fmt.Println("*************")
	fmt.Println()
	return GetMpegEncoding(header)
}

func ParseMp3(data []byte, encoding Encoding, includeUnitEncoding bool, verbose bool) AudioInfos {
	audioInfo := GetAudioInfosMpeg(data, 0, encoding, includeUnitEncoding, true, verbose)
	return audioInfo
}
