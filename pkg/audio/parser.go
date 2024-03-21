package audio

import (
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nice-pink/streamey/pkg/metadata"
)

type Parser struct {
	// data
	fullDataSize uint64
	currentData  []byte
	skipped      uint64
	// tag
	skippedTag bool
	// tagSize         int64 = 0
	currentTagIndex int64
	tagEnd          int64
	// audio
	encoding               Encoding
	foundEncoding          bool
	audioInfo              AudioInfos
	foundFirstFrame        bool
	skippedUntilFirstFrame uint64
	unitsTotal             uint64
	bytesTotal             uint64
	audioType              AudioType
}

func NewParser() *Parser {
	return &Parser{audioType: AudioTypeUnknown}
}

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
	if strings.ToUpper(ext) == "AAC" || strings.ToUpper(ext) == "M4A" {
		return AudioTypeAac
	}
	return AudioTypeUnknown
}

func GetFirstFrameIndex(data []byte, offset uint64, audioTypeGuessed AudioType) int64 {
	if audioTypeGuessed == AudioTypeMp3 {
		header := GetNextMpegHeader(data, offset)
		if header != nil {
			return header.Index
		}
		return -1
	}
	if audioTypeGuessed == AudioTypeAac {
		header := GetNextAdtsHeader(data, offset)
		if header != nil {
			encoded := hex.EncodeToString(data[header.Index : header.Index+7])
			fmt.Println(encoded)
			return header.Index
		}
		return -1
	}
	return -1 // int64(offset)
}

func GetAudioType(data []byte, offset uint64, typeGuessed AudioType) AudioType {
	// get audio type
	if typeGuessed == AudioTypeMp3 && StartsWithMpegSync(data[offset:]) {
		fmt.Println("Is mp3")
		return AudioTypeMp3
	}
	if typeGuessed == AudioTypeAac && StartsWithAdtsSync(data, offset) {
		fmt.Println("Is aac")
		return AudioTypeAac

	}
	fmt.Println("Is unknown")
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

func (p *Parser) Parse(data []byte, filepath string, includeUnitEncoding bool, verbose bool, printLogs bool) *AudioInfos {
	// skip tag
	tagSize := metadata.GetTagSize(data)
	if tagSize < 0 {
		fmt.Println("Error: Tag size could not be evaluated.")
		tagSize = 0
	} else if tagSize > 0 {
		fmt.Println("Tag size:", tagSize)
	}

	// parse audio
	var audioStart int64 = 0
	if tagSize > 0 {
		audioStart = tagSize - 1
	}

	// get audio type
	p.audioType = GuessAudioType(filepath)
	if p.audioType == AudioTypeUnknown {
		fmt.Println("Unknown audio type.")
		return nil
	}

	// get first frame
	firstFrameIndex := GetFirstFrameIndex(data, uint64(audioStart), p.audioType)
	fmt.Println("First frame index", firstFrameIndex)
	if firstFrameIndex < 0 {
		return nil
	}
	p.skippedUntilFirstFrame = uint64(firstFrameIndex - tagSize)
	p.foundFirstFrame = firstFrameIndex > audioStart

	// confirm audio type
	if GetAudioType(data, uint64(firstFrameIndex), p.audioType) != p.audioType {
		fmt.Println("Unknown audio type.")
		return nil
	}

	var err error
	if p.audioType == AudioTypeMp3 {
		p.encoding, err = GetEncodingFromFirstMpegHeader(data, uint64(firstFrameIndex))
		if err != nil {
			return nil
		}
		p.audioInfo = ParseMp3(data[firstFrameIndex:], p.encoding, includeUnitEncoding, verbose, verbose)

	} else if p.audioType == AudioTypeAac {
		p.encoding, err = GetEncodingFromFirstAdtsHeader(data, uint64(firstFrameIndex))
		if err != nil {
			return nil
		}
		p.audioInfo = ParseAac(data[firstFrameIndex:], p.encoding, includeUnitEncoding, verbose, verbose)
	}

	fmt.Println()
	p.audioInfo.FirstFrameIndex = firstFrameIndex
	p.audioInfo.TagSize = tagSize
	if printLogs {
		p.PrintAudioInfo()
	}

	p.unitsTotal = uint64(len(p.audioInfo.Units))
	p.bytesTotal = p.audioInfo.Units[p.unitsTotal-1].Index + uint64(p.audioInfo.Units[p.unitsTotal-1].Size)

	// log
	p.fullDataSize = uint64(len(data))
	if printLogs {
		p.LogParserResult(filepath)
	}

	return &p.audioInfo
}

func (p *Parser) ParseBlockwise(data []byte, audioTypeGuessed AudioType, includeUnitEncoding bool, verbose bool, printLogs bool) (*AudioInfos, error) {
	p.fullDataSize += uint64(len(data))
	p.currentData = append(p.currentData, data...)
	dataSize := len(p.currentData)
	// fmt.Println("data size", dataSize)
	var offset int64 = 0

	// skip tag
	if !p.skippedTag && p.currentTagIndex <= p.audioInfo.TagSize {
		if p.audioInfo.TagSize == 0 {
			p.audioInfo.TagSize = metadata.GetTagSize(p.currentData)
			if p.audioInfo.TagSize < 0 {
				fmt.Println("Error: Tag size could not be evaluated.")
				p.audioInfo.TagSize = 0
			} else if p.audioInfo.TagSize > 0 {
				fmt.Println("Tag size:", p.audioInfo.TagSize)
			}
		}

		// skip tag
		if p.audioInfo.TagSize-p.currentTagIndex < int64(dataSize) {
			p.tagEnd = p.audioInfo.TagSize - p.currentTagIndex
			p.currentTagIndex = p.audioInfo.TagSize - 1
			p.skippedTag = true
			fmt.Println("Skipped tag at index:", p.currentTagIndex)
		} else {
			p.currentTagIndex += int64(dataSize)
			p.currentData = p.currentData[:0]
			return nil, nil
		}
	}

	// get audio offset
	offset = GetFirstFrameIndex(p.currentData, uint64(p.tagEnd), audioTypeGuessed)
	// fmt.Println("first offset:", offset)
	if offset < 0 {
		return nil, nil
	}

	if !p.foundFirstFrame {
		// get audio
		p.audioType = GetAudioType(p.currentData, uint64(offset), audioTypeGuessed)
		if p.audioType == AudioTypeUnknown {
			fmt.Println("Unknown audio type.")
			return nil, nil
		}
		if !p.foundEncoding {
			if audioTypeGuessed == AudioTypeMp3 {
				header := GetNextMpegHeader(p.currentData, uint64(offset))
				if header != nil {
					p.encoding = GetMpegEncoding(*header)
					p.foundEncoding = true
				}
			} else if audioTypeGuessed == AudioTypeAac {
				header := GetNextAdtsHeader(p.currentData, uint64(offset))
				if header != nil {
					p.encoding = GetAdtsEncoding(*header)
					p.foundEncoding = true
				}
			}

			// 	var err error
			// 	for {
			// 		p.encoding, err = GetEncodingFromFirstMpegHeader(data, uint64(offset))
			// 		if err == nil {
			// 			p.foundEncoding = true
			// 			fmt.Println("skipped", p.skipped)
			// 			break
			// 		}
			// 		offset = GetFirstFrameIndex(data, uint64(offset)+1, p.audioType)
			// 		// fmt.Println("s first offset:", offset)
			// 		if offset < 0 {
			// 			p.skipped += uint64(dataSize) - uint64(offset)
			// 			p.currentData = p.currentData[:0]
			// 			return nil, nil
			// 		}
			// 	}
		}

		p.foundFirstFrame = offset >= p.tagEnd
		p.skippedUntilFirstFrame = uint64(offset - p.tagEnd)
		p.tagEnd = 0
	}

	// parse audio
	var audioInfoBlock AudioInfos
	if p.audioType == AudioTypeMp3 {
		audioInfoBlock = ParseMp3(p.currentData[offset:], p.encoding, includeUnitEncoding, verbose, verbose)
	} else if p.audioType == AudioTypeAac {
		fmt.Println("Not jet implemented!")
		audioInfoBlock = ParseAac(p.currentData[offset:], p.encoding, includeUnitEncoding, verbose, verbose)
	}

	// remove handled data from
	units := len(audioInfoBlock.Units)
	if units > 0 {
		i := audioInfoBlock.Units[units-1].Index + uint64(audioInfoBlock.Units[units-1].Size) + uint64(offset)
		p.bytesTotal += i
		p.unitsTotal += uint64(units)
		p.currentData = p.currentData[i:]
	}

	// log infos
	if printLogs {
		fmt.Println()
		p.PrintAudioInfo()
		p.LogParserResult("")
	}

	return &audioInfoBlock, nil
}

func (p *Parser) PrintAudioInfo() {
	p.audioInfo.Print()
}

func (p *Parser) LogParserResult(filepath string) {
	// log
	fmt.Println()
	fmt.Println("---")
	if filepath != "" {
		fmt.Println("File path:", filepath)
	}
	fmt.Println("File size:", p.fullDataSize)
	fmt.Println("Tag size:", p.audioInfo.TagSize)
	fmt.Println("Skipped bytes to first frame:", p.skippedUntilFirstFrame)
	fmt.Println("Audio size:", p.fullDataSize-p.skippedUntilFirstFrame-uint64(p.audioInfo.TagSize))
	fmt.Println("Audio frames:", p.unitsTotal)
	fmt.Println("Bytes frames:", p.bytesTotal)
}

func MakeFirstFramePrivate(data []byte, offset uint64, audioType AudioType) {
	if audioType == AudioTypeMp3 {
		SetMpegPrivate(data, offset)
	}
}

// mpeg

func GetEncodingFromFirstMpegHeader(data []byte, offset uint64) (Encoding, error) {
	if len(data)-int(offset) < MpegHeaderSize {
		fmt.Println("Too small")
		return Encoding{}, errors.New("buffer too small")
	}
	// get frame infos
	fmt.Println()
	fmt.Println("*************")
	fmt.Println("Initial header")
	fmt.Println("use encoding!")
	header := GetMpegHeader(data[offset:], int64(offset))
	if !header.IsValid(true) {
		fmt.Println("Error: Header is not valid")
		header.Print(false)
		return Encoding{}, errors.New("invalid header")
	}
	header.Print(true)
	fmt.Println("*************")
	fmt.Println()
	return GetMpegEncoding(header), nil
}

func ParseMp3(data []byte, encoding Encoding, includeUnitEncoding bool, printHeaders bool, verbose bool) AudioInfos {
	audioInfo := GetAudioInfosMpeg(data, 0, encoding, includeUnitEncoding, printHeaders, verbose)
	return audioInfo
}

// aac

func GetEncodingFromFirstAdtsHeader(data []byte, offset uint64) (Encoding, error) {
	if len(data)-int(offset) < AdtsHeaderSizeMax {
		fmt.Println("Too small")
		return Encoding{}, errors.New("buffer too small")
	}
	// get frame infos
	fmt.Println()
	fmt.Println("*************")
	fmt.Println("Initial header", offset)
	fmt.Println("use encoding!")
	header := GetAdtsHeader(data[offset:], int64(offset))
	if !header.IsValid(true) {
		fmt.Println("Error: Header is not valid")
		header.Print(false)
		return Encoding{}, errors.New("invalid header")
	}
	encoded := hex.EncodeToString(data[offset : offset+7])
	fmt.Println(encoded)
	header.Print(true)
	fmt.Println("*************")
	fmt.Println()
	return GetAdtsEncoding(header), nil
}

func ParseAac(data []byte, encoding Encoding, includeUnitEncoding bool, printHeaders bool, verbose bool) AudioInfos {
	fmt.Println("get aac")
	audioInfo := GetAudioInfosAac(data, 0, encoding, includeUnitEncoding, printHeaders, verbose)
	return audioInfo
}
