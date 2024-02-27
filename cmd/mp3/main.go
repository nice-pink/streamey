package main

import (
	"flag"
	"io"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/audio"
)

func main() {
	filepath := flag.String("filepath", "", "Filepath")
	flag.Parse()

	// get file data

	if *filepath == "" {
		flag.Usage()
		os.Exit(2)
	}

	file, err := os.Open(*filepath)
	if err != nil {
		log.Err(err, "Cannot open file.")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err, "Cannot read file.")
	}

	// skip metadata if any

	// metaSize := metadata.GetIdV3HeaderSize(data)
	// if metaSize > 0 {
	// 	data = data[metaSize:]
	// }

	// get frame infos

	header := audio.GetMpegHeader(data)
	if !header.IsValid() {
		log.Error("Header is not valid")
		header.Print(false)
		os.Exit(2)
	}

	encoding := audio.GetMp3Encoding(header)
	audioInfo := audio.GetAudioInfos(data, 0, encoding, true)
	log.Info()
	audioInfo.Print()
}
