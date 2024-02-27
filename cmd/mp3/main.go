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

	// parse audio
	audio.Parse(data)
}
