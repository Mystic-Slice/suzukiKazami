package main

import (
	"fmt"
)

const numberOfSites = 10

func main() {
	fmt.Printf("Initializing logger\n")
	InitLogger()

	// For each pair of sites, create two channels for each way of communication
	// Channels communication goes from process i to process j
	var allChannels [numberOfSites][numberOfSites]chan int
	for i := 0; i < numberOfSites; i++ {
		for j := 0; j < numberOfSites; j++ {
			allChannels[i][j] = make(chan int, 10000)
		}
	}

	fmt.Printf("Creating and initializing token\n")
	// Create a token
	var token Token
	token.initialize()
	// Give token to site 0
	token.currentOwner = 0
	go token.logger()

	fmt.Printf("Number of sites = %v\n", numberOfSites)
	fmt.Printf("Creating and starting sites\n")
	// Create sites and start go routines for each
	var allSites [numberOfSites]Site
	for i := 0; i < numberOfSites; i++ {
		sendChannels := allChannels[i][:]
		listenChannels := make([]chan int, 0)
		for j := 0; j < numberOfSites; j++ {
			listenChannels = append(listenChannels, allChannels[j][i])
		}

		allSites[i].initialize(i, sendChannels, listenChannels, &token)
		go allSites[i].logger()
		go allSites[i].run()
		fmt.Printf("Site %v up and running\n", i)
	}

	var input string
	fmt.Print("Press any key to terminate...")
	fmt.Scanln(&input)
}
