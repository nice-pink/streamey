package metadata

import (
	"encoding/binary"
	"fmt"

	"github.com/nice-pink/streamey/pkg/util"
)

type QuicktimeTag struct {
	TagSize int64
}

const (
	QuicktimeHeaderSizeMin int    = 12
	QuicktimeSync          string = "667479704D344120"
)

func GetQuicktimeTag(data []byte) QuicktimeTag {
	var tag QuicktimeTag
	tag.TagSize = GetQuicktimeTagSize(data)
	return tag
}

func IsValidQuicktimeHeader(data []byte) bool {
	if len(data) < QuicktimeHeaderSizeMin {
		fmt.Println("too small")
		return false
	}

	if !StartsWithQuicktimeSync(data) {
		return false
	}

	return true
}

func StartsWithQuicktimeSync(data []byte) bool {
	return util.BytesEqualHex(QuicktimeSync, data[4:13])
}

func GetQuicktimeTagSize(data []byte) int64 {
	dataSize := len(data)

	// count up all container sizes until sync word for aac is found!
	var size int64 = GetBlockSize(data, 0)
	var blockSize int64 = 0
	for {
		blockSize = GetBlockSize(data, uint64(size))
		if blockSize == 0 {
			fmt.Println("No block size")
			break
		}

		if size+blockSize == int64(dataSize) {
			break
		}
		size += blockSize
		// fmt.Println("New size", size)
	}

	return size
}

func IntStartsWithAdtsSync(data []byte, offset uint64) bool {
	return util.BytesEqualHexWithMask("FFF0", "FFF6", data[offset:offset+2])
}

func GetBlockSize(data []byte, offset uint64) int64 {
	return int64(binary.BigEndian.Uint32(data[offset : offset+4]))
}
