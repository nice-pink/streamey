package main

import (
	"flag"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/audio"
	"github.com/nice-pink/streamey/pkg/icecast"
	"github.com/nice-pink/streamey/pkg/metricmanager"
	"github.com/nice-pink/streamey/pkg/network"
	"github.com/nice-pink/streamey/pkg/util"
)

var wg sync.WaitGroup

func main() {
	log.Info("--- Start streamey ---")

	// flags
	url := flag.String("url", "", "Destination url")
	filepath := flag.String("filepath", "", "File to stream.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	bitrate := flag.Int64("bitrate", 0, "Send bitrate.")
	sr := flag.Int("sr", 44100, "Sampe rate.")
	test := flag.Bool("test", false, "Test sending with internal reader.")
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	isIcecast := flag.Bool("isIcecast", false, "Send icecast.")
	validateTest := flag.String("validateTest", "", "Validation test.")
	metrics := flag.Bool("metrics", false, "Add metrics.")
	metricPrefix := flag.String("metricPrefix", "streamey_", "Metric prefix.")
	metricPort := flag.Int("metricPort", 9090, "Metric port.")
	flag.Parse()

	// read file
	data := GetData(*filepath)

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
	go Stream(streamUrl, float64(*bitrate), *sr, data, *reconnect, *isIcecast, *verbose)

	wg.Add(goRoutineCounter)
	wg.Wait()
}

func GetData(filepath string) []byte {
	filepathFinal := filepath
	if strings.HasPrefix(filepath, "http") {
		filepathFinal = "file.mp3"
		util.DownloadFile(filepath, filepathFinal)
	}

	// is local file
	file, err := os.Open(filepathFinal)
	if err != nil {
		log.Err(err, "Cannot open file.")
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err, "Cannot read file.")
	}
	return data
}

func Stream(url string, bitrate float64, sampleRate int, data []byte, reconnect bool, isIcecast bool, verbose bool) {
	var header []byte
	if isIcecast {
		// header
		meta := icecast.IcyMeta{Bitrate: int(bitrate), Channels: 2, SampleRate: sampleRate}
		icyAdd, err := icecast.GetIcyAddress(url)
		if err != nil {
			log.Error("Icy address")
			panic(err)
		}
		header, err = icecast.GetIcecastPutHeader(icyAdd, meta)
		if err != nil {
			os.Exit(2)
		}
		// send buffer
		address := icyAdd.Domain + ":" + icyAdd.Port
		network.StreamBuffer(address, bitrate, header, data, reconnect, verbose)
	} else {
		// send buffer as is
		network.StreamBuffer(url, bitrate, header, data, reconnect, verbose)
	}
	wg.Done()
}

func Receive(validate string, metrics bool, verbose bool) {
	if strings.ToLower(validate) == "audio" {
		log.Newline()
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
		log.Newline()
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
