package metadata

import "fmt"

func GetTagType(data []byte) MetadataTagType {
	if StartsWithId3V2Sync(data) {
		fmt.Println("Tag type id3v2.")
		return MetadataTagTypeId3V2
	}
	return MetadataTagTypeUnknown
}

func GetTagSize(data []byte) int64 {
	tagType := GetTagType(data)

	// Id3V2
	if tagType == MetadataTagTypeId3V2 {
		if !IsValidId3V2Header(data) {
			fmt.Println("Invalid header")
			return -1
		}
		return int64(GetId3V2TagSize(data))
	}
	return 0
}
