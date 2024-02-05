package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
)

func main() {
	log.Info("--- Start cleaney ---")
	log.Time()

	folder := flag.String("folder", "", "Folder containing files.")
	sec := flag.Int64("sec", -1, "All files are deleted older than 'sec' seconds.")
	delete := flag.Bool("delete", false, "Delete files older than defined.")
	flag.Parse()

	if *folder == "" {
		flag.Usage()
		os.Exit(2)
	}

	// get files
	files := filesystem.ListFiles(*folder, *sec, true)

	// logs
	ago := -time.Duration(*sec) * time.Second
	dateThreshold := time.Now().Add(ago)
	log.Info()
	log.Info("Files older than:", dateThreshold)
	for _, file := range files {
		log.Info(file)
	}
	log.Info(strconv.Itoa(len(files)), "files")
	log.Info()

	// delete
	if *delete {
		log.Info("Delete files...")
		filesystem.DeleteFiles(files)
		log.Info("Done!")
	}
}
