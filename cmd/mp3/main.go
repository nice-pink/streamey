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
	block := flag.Bool("block", false, "Parse file in blocks.")
	verbose := flag.Bool("verbose", false, "Make output verbose.")
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

	if *block {
		// parse continuously
		Blockwise(data, *verbose)
	} else {
		// parse audio
		audio.Parse(data, *filepath, *verbose)
	}
}

func Blockwise(data []byte, verbose bool) {
	dataSize := len(data)
	index := 0
	for {
		if index >= dataSize {
			break
		}
		iMax := min(index+1024, dataSize)
		audio.ParseBlockwise(data[index:iMax], audio.AudioTypeMp3, verbose)

		index = iMax
	}
}
