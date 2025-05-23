package configmanager

import (
	"encoding/json"
	"os"

	"github.com/nice-pink/audio-tool/pkg/audio/encodings"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/miniomanager"
)

type ReadConfig struct {
	Expectations encodings.Expectations
	Minio        miniomanager.MinioConfig
}

func GetReadConfig(configFile string) ReadConfig {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Error("Config file does not exist.")
		panic(err)
	}

	var config ReadConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Error("No valid expectations config.")
		panic(err)
	}
	return config
}
