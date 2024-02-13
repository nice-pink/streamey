package main

import (
	"flag"
	"io"
	"os"
	"sync"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/streamey"
)

var wg sync.WaitGroup

const (
	delay       int64  = 3600
	bucket      string = "data"
	minioFolder string = "audio"
)

func main() {
	log.Info("--- Start streamey ---")
	log.Time()

	// flags
	url := flag.String("url", "", "Destination url")
	filepath := flag.String("filepath", "", "File to stream.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	bitrate := flag.Int64("bitrate", 0, "Send bitrate.")
	test := flag.Bool("test", false, "Test sending with internal reader.")
	flag.Parse()

	// read file
	file, err := os.Open(*filepath)
	if err != nil {
		log.Err(err, "Cannot open file.")
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err, "Cannot read file.")
	}

	goRoutineCounter := 0

	if *test {
		goRoutineCounter++
		go Receive()
	}

	goRoutineCounter++
	go Stream(*url, float64(*bitrate), data, *reconnect)

	wg.Add(goRoutineCounter)
	wg.Wait()
}

func Stream(url string, bitrate float64, data []byte, reconnect bool) {
	streamey.Stream(url, bitrate, data, reconnect)
	wg.Done()
}

func Receive() {
	streamey.ReadTest(9999, true)
	wg.Done()
}
