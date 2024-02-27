package metadata

type MetadataTagType int

const (
	MetadataTagTypeNone MetadataTagType = iota
	MetadataTagTypeUnknown
	MetadataTagTypeId3V2
)
