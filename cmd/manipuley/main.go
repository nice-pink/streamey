package main

import (
	"flag"
	"io"
	"os"

	"github.com/nice-pink/audio-tool/pkg/audio/encodings"
	"github.com/nice-pink/goutil/pkg/log"
)

func main() {
	input := flag.String("input", "", "Input filepath.")
	output := flag.String("output", "", "Output filepath.")
	private := flag.Bool("private", false, "Make first frame private.")
	verbose := flag.Bool("verbose", false, "Make output verbose.")
	flag.Parse()

	// get file data

	if *input == "" {
		flag.Usage()
		os.Exit(2)
	}

	file, err := os.Open(*input)
	if err != nil {
		log.Err(err, "Cannot open file.")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err, "Cannot read file.")
	}

	// parse audio
	parser := encodings.NewParser()
	audioInfo := parser.Parse(data, *input, false, *verbose, true)
	if audioInfo == nil {
		log.Error("No audio infos!")
	}

	if *private {
		encodings.MakeFirstFramePrivate(data, uint64(audioInfo.FirstFrameIndex), encodings.GuessAudioType(*input))

		if *output != "" {
			err := os.WriteFile(*output, data, 0644)
			if err != nil {
				log.Err(err, "Write output error.")
			}
		}
	}
}
