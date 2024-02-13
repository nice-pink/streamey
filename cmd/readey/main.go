package main

import (
	"encoding/json"
	"flag"
	"os"
	"sync"
	"time"

	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/miniomanager"
	"github.com/nice-pink/streamey/pkg/readey"
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
	maxBytes := flag.Int64("maxBytes", -1, "[Optional] Max bytes to read from url.")
	outputFilepath := flag.String("outputFilepath", "", "[Optional] Output file path, if data should be dumped to file.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	minioConfig := flag.String("minioConfig", "", "[Optional] Json config file for minio. Use minio if defined.")
	minioCleanUpAfterSec := flag.Int64("minioCleanUpAfterSec", 0, "[Optional] Cleanup minio bucket after seconds.")
	flag.Parse()

	// read stream
	go ReadStream(*url, *maxBytes, *outputFilepath, *reconnect)

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

func ReadStream(url string, maxBytes int64, outputFilepath string, reconnect bool) {
	readey.ReadStream(url, maxBytes, outputFilepath, reconnect)
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