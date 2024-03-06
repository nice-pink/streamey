package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/audio"
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
	validate := flag.String("validate", "", "Validation type. [audio]")
	outputFilepath := flag.String("outputFilepath", "", "[Optional] Output file path, if data should be dumped to file.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	minioConfig := flag.String("minioConfig", "", "[Optional] Json config file for minio. Use minio if defined.")
	minioCleanUpAfterSec := flag.Int64("minioCleanUpAfterSec", 0, "[Optional] Cleanup minio bucket after seconds.")
	verbose := flag.Bool("verbose", false, "Verbose Logging.")
	flag.Parse()

	// read stream
	go ReadStream(*url, *maxBytes, *outputFilepath, *reconnect, *timeout, *validate, *verbose)

	// start minio sync
	goRoutineCounter := 1
	if *minioConfig != "" {
		goRoutineCounter++

		go ManageMinio(*minioConfig, delay, *outputFilepath, *minioCleanUpAfterSec, *reconnect)
	}

	// wait for go routines to be done
	wg.Add(goRoutineCounter)
	wg.Wait()
}

// stream

func ReadStream(url string, maxBytes uint64, outputFilepath string, reconnect bool, timeout int, validate string, verbose bool) {
	var validator audio.Validator
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
		validator = *audio.NewValidator(expectations, verbose)
	}
	network.ReadStream(url, maxBytes, outputFilepath, reconnect, time.Duration(timeout)*time.Second, validator)
	wg.Done()
}

// minio

func ManageMinio(configFile string, delay int64, localFolder string, minioCleanUpAfterSec int64, loop bool) {
	// read config
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Error(err, "Minio config file does not exist.")
		os.Exit(2)
	}

	var config miniomanager.MinioConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Error(err, "No valid minio config.")
		os.Exit(2)
	}

	useSsl := true
	m := miniomanager.NewMinioManager(config, useSsl)

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
