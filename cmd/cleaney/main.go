package main

import (
	"flag"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/util"
)

func main() {
	log.Info("--- Start cleaney ---")

	folder := flag.String("folder", "", "Folder containing files.")
	sec := flag.Int64("sec", -1, "All files are deleted older than 'sec' seconds.")
	delete := flag.Bool("delete", false, "Delete files older than defined.")
	flag.Parse()

	if *folder == "" {
		flag.Usage()
		os.Exit(2)
	}

	util.CleanUp(*folder, *sec, true, *delete)
}
