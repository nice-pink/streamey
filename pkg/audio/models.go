package audio

// encoding

type Encoding struct {
	ContainerName string // e.g. wav, mpeg, aac, ...
	CodecName     string
	SampleRate    int
	Bitrate       int
	IsStereo      bool
	Profile       string // aac profile; could also be pcm for wav
	FrameSize     int    // samples per frame
}

func IsEncoded(e Encoding) bool {
	return e.CodecName != ""
}

// unit infos

type UnitInfo struct {
	Index uint64
	Size  int
}
