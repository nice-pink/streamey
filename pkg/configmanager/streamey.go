package configmanager

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/nice-pink/goutil/pkg/log"
)

type StreamFormat int

const (
	StreamFormatPlain StreamFormat = iota
	StreamFormatShoutcast
	StreamFormatIcecast
)

func GetStreamFormat(name string) StreamFormat {
	switch strings.ToLower(name) {
	case "shoutcast":
		return StreamFormatShoutcast
	case "icecast":
		return StreamFormatIcecast
	default:
		return StreamFormatPlain
	}
}

// config

type StreamsConfig struct {
	Items []StreamConfig
}

type StreamConfig struct {
	ChannelName string
	Audio       AudioConfig
	Metadata    MetadataConfig
	Playlist    Playlist
}

type AudioConfig struct {
	TargetUrl  string
	Bitrate    int
	SampleRate int
	Format     string
}

type MetadataConfig struct {
	TargetUrl string
	Template  string
	Headers   map[string]string
}

type Playlist struct {
	ContentType string
	Items       []PlaylistItem
}

type PlaylistItem struct {
	Type     string
	Artist   string
	Title    string
	Album    string
	Filepath string
	Duration float64
}

func GetStreamConfig(filepath string) StreamsConfig {
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Error("Config file does not exist.")
		panic(err)
	}

	var config StreamsConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Error("No valid streams config.")
		panic(err)
	}
	return config
}
