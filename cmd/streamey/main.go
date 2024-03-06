package main

import (
	"flag"
	"io"
	"os"
	"sync"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/network"
)

var wg sync.WaitGroup

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

	streamUrl := *url
	if *test {
		goRoutineCounter++
		go Receive()
		// overwrite url
		streamUrl = "localhost:9999"
	}

	goRoutineCounter++
	go Stream(streamUrl, float64(*bitrate), data, *reconnect)

	wg.Add(goRoutineCounter)
	wg.Wait()
}

func Stream(url string, bitrate float64, data []byte, reconnect bool) {
	network.StreamBuffer(url, bitrate, data, reconnect)
	wg.Done()
}

func Receive() {
	network.ReadTest(9999, true)
	wg.Done()
}
