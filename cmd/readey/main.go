package main

import (
	"flag"
	"strings"
	"sync"
	"time"

	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/audio"
	"github.com/nice-pink/streamey/pkg/configmanager"
	"github.com/nice-pink/streamey/pkg/miniomanager"
	"github.com/nice-pink/streamey/pkg/network"
)

var wg sync.WaitGroup

const (
	delay       int64  = 3600
	bucket      string = "data"
	minioFolder string = "audio"
)

func main() {
	log.Info("--- Start readey ---")
	log.Time()

	// flags
	url := flag.String("url", "", "Stream url")
	maxBytes := flag.Uint64("maxBytes", 0, "[Optional] Max bytes to read from url.")
	timeout := flag.Int("timeout", 30, "Timeout. Default: 30sec")
	validate := flag.String("validate", "", "Validation type. [audio, privateBit]")
	outputFilepath := flag.String("outputFilepath", "", "[Optional] Output file path, if data should be dumped to file.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	minioConfig := flag.String("minioConfig", "", "[Optional] Json config file for minio. Use minio if defined.")
	minioCleanUpAfterSec := flag.Int64("minioCleanUpAfterSec", 0, "[Optional] Cleanup minio bucket after seconds.")
	config := flag.String("config", "", "Config file.")
	verbose := flag.Bool("verbose", false, "Verbose Logging.")
	flag.Parse()

	var c configmanager.Config
	if *config != "" {
		c = configmanager.Get(*config)
	}

	// read stream
	go ReadStream(*url, *maxBytes, *outputFilepath, *reconnect, *timeout, c, *validate, *verbose)

	// start minio sync
	goRoutineCounter := 1
	if *minioConfig != "" {
		goRoutineCounter++

		go ManageMinio(c, delay, *outputFilepath, *minioCleanUpAfterSec, *reconnect)
	}

	// wait for go routines to be done
	wg.Add(goRoutineCounter)
	wg.Wait()
}

// stream

func ReadStream(url string, maxBytes uint64, outputFilepath string, reconnect bool, timeout int, config configmanager.Config, validate string, verbose bool) {
	if strings.ToLower(validate) == "audio" {
		log.Info()
		log.Info("### Audio validation")
		// expectations := audio.Expectations{
		// 	IsCBR: true,
		// 	Encoding: audio.Encoding{
		// 		Bitrate:  256,
		// 		IsStereo: true,
		// 	},
		// }
		expectations := config.Expectations
		expectations.Print()
		log.Info("###")
		log.Info()
		validator := audio.NewEncodingValidator(true, expectations, verbose)
		network.ReadStream(url, maxBytes, outputFilepath, reconnect, time.Duration(timeout)*time.Second, validator)
	} else if strings.ToLower(validate) == "privatebit" {
		validator := audio.NewPrivateBitValidator(true, verbose, audio.GuessAudioType(url))
		network.ReadStream(url, maxBytes, outputFilepath, reconnect, time.Duration(timeout)*time.Second, validator)
	} else {
		validator := network.DummyValidator{}
		network.ReadStream(url, maxBytes, outputFilepath, reconnect, time.Duration(timeout)*time.Second, validator)
	}
	wg.Done()
}

// minio

func ManageMinio(config configmanager.Config, delay int64, localFolder string, minioCleanUpAfterSec int64, loop bool) {
	useSsl := true
	m := miniomanager.NewMinioManager(config.Minio, useSsl)

	// run loop
	duration := time.Duration(delay) * time.Second
	for {
		// start with the dalay
		time.Sleep(duration)

		// push
		PushFiles(m, localFolder)

		// delete
		//day := int64(3600 * 24)
		m.DeleteFiles(bucket, minioFolder, minioCleanUpAfterSec)

		if !loop {
			break
		}
	}
	wg.Done()
}

func PushFiles(m *miniomanager.MinioManger, folder string) {
	files := filesystem.ListFiles(folder, 0, true)

	for _, file := range files {
		m.PutFile(bucket, minioFolder, folder, file, true)
	}
}
