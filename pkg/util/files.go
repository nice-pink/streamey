package util

import (
	"strconv"
	"time"

	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
)

func CleanUp(folder string, sec int64, ignoreHiddenFiles bool, delete bool) {
	// get files
	files := filesystem.ListFiles(folder, sec, ignoreHiddenFiles)

	// logs
	ago := -time.Duration(sec) * time.Second
	dateThreshold := time.Now().Add(ago)
	log.Info()
	log.Info("Files older than:", dateThreshold)
	for _, file := range files {
		log.Info(file)
	}
	log.Info(strconv.Itoa(len(files)), "files")
	log.Info()

	// delete
	if delete {
		log.Info("Delete files...")
		filesystem.DeleteFiles(files)
		log.Info("Done!")
	}
}
