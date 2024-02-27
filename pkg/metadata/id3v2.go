package metadata

import (
	"encoding/binary"

	"github.com/nice-pink/streamey/pkg/util"
)

type IdV3Tag struct {
	TagSize int32
}

const (
	HeaderSize int    = 10
	Id3V3Sync  string = "494433"
	FooterMask string = "10"
)

func GetTag(data []byte) IdV3Tag {
	var tag IdV3Tag
	tag.TagSize = GetIdV3HeaderSize(data)
	return tag
}

func IsValidIdV3Header(data []byte) bool {
	if len(data) < HeaderSize {
		return false
	}

	if !StartsWithIdV3Sync(data) {
		return false
	}

	return true
}

func StartsWithIdV3Sync(data []byte) bool {
	return util.BytesEqualHex(Id3V3Sync, data)
}

func HasIdV3Footer(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	bytes := []byte{data[5]}
	return util.BytesEqualHexWithMask(FooterMask, FooterMask, bytes)
}

func GetIdV3HeaderSize(data []byte) int32 {
	bits := util.BitsFromBytes(data, 48, 32)
	headerValue := binary.BigEndian.Uint32(bits)

	footerSize := 0
	if HasIdV3Footer(data) {
		footerSize = HeaderSize
	}

	return int32(util.Unsynchsafe(headerValue)) + int32(footerSize)
}
