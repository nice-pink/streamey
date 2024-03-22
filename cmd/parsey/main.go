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
	repeatBlock := flag.Int("n", 0, "Repeat blocks.")
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
		guessedAudioType := audio.GuessAudioType(*filepath)
		// parse continuously
		Blockwise(data, guessedAudioType, *repeatBlock, *verbose)
	} else {
		// parse audio
		parser := audio.NewParser()
		parser.Parse(data, *filepath, false, *verbose, true)
	}
}

func Blockwise(data []byte, guessedAudioType audio.AudioType, repeatBuffer int, verbose bool) {
	parser := audio.NewParser()
	dataSize := len(data)
	repeated := 0
	index := 0
	for {
		if index >= dataSize {
			repeated++
			if repeated > repeatBuffer {
				break
			}
			// repeat
			index = 0
		}
		iMax := min(index+1024, dataSize)
		parser.ParseBlockwise(data[index:iMax], guessedAudioType, false, verbose, false)

		index = iMax
	}

	println()
	parser.PrintAudioInfo()
	parser.LogParserResult("")
}
