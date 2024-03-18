package metadata

import (
	"encoding/binary"

	"github.com/nice-pink/streamey/pkg/util"
)

type Id3V2Tag struct {
	TagSize int64
}

const (
	Id3V2HeaderSize int    = 10
	Id3V2Sync       string = "494433"
	Id3V2FooterMask string = "10"
)

func GetId3V2Tag(data []byte) Id3V2Tag {
	var tag Id3V2Tag
	tag.TagSize = GetId3V2TagSize(data)
	return tag
}

func IsValidId3V2Header(data []byte) bool {
	if len(data) < Id3V2HeaderSize {
		return false
	}

	if !StartsWithId3V2Sync(data) {
		return false
	}

	return true
}

func StartsWithId3V2Sync(data []byte) bool {
	return util.BytesEqualHex(Id3V2Sync, data)
}

func HasId3V2Footer(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	bytes := []byte{data[5]}
	return util.BytesEqualHexWithMask(Id3V2FooterMask, Id3V2FooterMask, bytes)
}

func GetId3V2TagSize(data []byte) int64 {
	bytes := data[6:10]
	headerValue := binary.BigEndian.Uint32(bytes)

	footerSize := 0
	if HasId3V2Footer(data) {
		footerSize = Id3V2HeaderSize
	}
	return int64(util.Unsynchsafe(headerValue)) + int64(footerSize)
}
