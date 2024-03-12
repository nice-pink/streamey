package main

import (
	"flag"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/audio"
	"github.com/nice-pink/streamey/pkg/metricmanager"
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
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	validateTest := flag.String("validateTest", "", "Validation test.")
	metrics := flag.Bool("metrics", false, "Add metrics.")
	metricPrefix := flag.String("metricPrefix", "streamey_", "Metric prefix.")
	metricPort := flag.Int("metricPort", 9090, "Metric port.")
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

	// start metrics server
	if *metrics {
		metricmanager.MetricPrefix = *metricPrefix
		go metricmanager.Listen(*metricPort)
	}

	goRoutineCounter := 0

	streamUrl := *url
	if *test {
		goRoutineCounter++
		go Receive(*validateTest, *metrics, *verbose)
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

func Receive(validate string, metrics bool, verbose bool) {
	if strings.ToLower(validate) == "audio" {
		log.Info()
		log.Info("### Audio validation")
		expectations := audio.Expectations{
			IsCBR: true,
			Encoding: audio.Encoding{
				Bitrate:  256,
				IsStereo: true,
			},
		}
		expectations.Print()
		log.Info("###")
		log.Info()
		validator := audio.NewEncodingValidator(true, expectations, metrics, verbose)
		network.ReadTest(9999, true, validator)
	} else if strings.ToLower(validate) == "privatebit" {
		log.Info("PrivateBit validation.")
		validator := audio.NewPrivateBitValidator(true, audio.AudioTypeMp3, metrics, verbose)
		network.ReadTest(9999, true, validator)
	} else {
		validator := network.DummyValidator{}
		network.ReadTest(9999, true, validator)
	}
	wg.Done()
}
