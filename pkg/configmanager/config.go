package configmanager

import (
	"encoding/json"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/audio"
	"github.com/nice-pink/streamey/pkg/miniomanager"
)

type Config struct {
	Expectations audio.Expectations
	Minio        miniomanager.MinioConfig
}

func Get(configFile string) Config {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Error("Config file does not exist.")
		panic(err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Error("No valid expectations config.")
		panic(err)
	}
	return config
}
