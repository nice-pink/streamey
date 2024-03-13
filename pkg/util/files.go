package util

import (
	"io"
	"net/http"
	"os"
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
	log.Newline()
	log.Info("Files older than:", dateThreshold)
	for _, file := range files {
		log.Info(file)
	}
	log.Info(strconv.Itoa(len(files)), "files")
	log.Newline()

	// delete
	if delete {
		log.Info("Delete files...")
		filesystem.DeleteFiles(files)
		log.Info("Done!")
	}
}

func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
