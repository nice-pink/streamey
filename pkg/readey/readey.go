package readey

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/goutil/pkg/network"
)

func ReadStream(url string, maxBytes int64, outputFilepath string, reconnect bool) {
	config := network.DefaultRequestConfig()
	//config.MaxBytes = 144000000
	config.MaxBytes = maxBytes

	// early exit
	if url == "" {
		log.Info()
		log.Error("Define url!")
		flag.Usage()
		os.Exit(2)
	}

	// log infos
	log.Info("Connect to url", url)
	if maxBytes > 0 {
		log.Info("Should stop after reading bytes:", maxBytes)
	} else {
		log.Info("Will read data until connection breaks.")
	}
	if outputFilepath != "" {
		log.Info("Dump data to file:", outputFilepath)
	}

	log.Info()
	iteration := 0
	for {
		log.Info("Start connection")
		log.Time()
		r := network.NewRequester(config)
		filepath := getFilePath(outputFilepath, iteration, true)
		r.ReadStream(url, filepath)
		log.Time()
		log.Info()
		if !reconnect {
			break
		}
		iteration++
	}
}

func getFilePath(baseFilePath string, iteration int, addTimestamp bool) string {
	if baseFilePath == "" {
		return ""
	}
	if addTimestamp {
		now := time.Now()
		baseFilePath += "_" + strconv.FormatInt(now.Unix(), 10)
	}
	return baseFilePath + "_" + strconv.Itoa(iteration)
}
