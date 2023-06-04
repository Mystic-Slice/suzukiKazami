package main

import (
	"os"
)

// get the current working directory
var DIR string

// Creates/Cleans up the logging folder
func InitLogger() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	DIR = cwd + "\\tmp\\"
	err = os.RemoveAll(DIR)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(DIR, 0700)
	if err != nil {
		panic(err)
	}
}

/*
 * Log message to file with name `fileName`
 * Also adds the log message to a common log file `commonLog.log`
 */
func log(message string, fileName string) {
	// Write message to file with name filename
	// open output file
	fo, err := os.OpenFile(DIR+fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := fo.WriteString(message); err != nil {
		panic(err)
	}

	if fileName != "commonLog.log" {
		log(message, "commonLog.log")
	}
}
