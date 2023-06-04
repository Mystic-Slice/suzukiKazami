package main

import (
	"fmt"
	"sync"
	"time"
)

type Token struct {
	lastServicedSequenceNumber []int       // Sequence number of the last request serviced for each site
	serviceQueue               Queue       // Queue of sites waiting for the token
	currentOwner               int         // Site id of the current owner of the token
	lock                       *sync.Mutex // Lock for the token -> Use when modifying any of the above fields
}

// Initializes the token fields
func (t *Token) initialize() {
	t.lastServicedSequenceNumber = make([]int, numberOfSites)
	for i := 0; i < numberOfSites; i++ {
		t.lastServicedSequenceNumber[i] = -1
	}
	t.serviceQueue = Queue{}
	t.currentOwner = -1
	t.lock = &sync.Mutex{}
}

// Logs the state of the token every second
func (t *Token) logger() {
	for {
		log(fmt.Sprintf("currentOwner = %v\n", t.currentOwner), "token.log")
		log(fmt.Sprintf("lastServicedSequenceNumber = %v\n", t.lastServicedSequenceNumber), "token.log")
		log(fmt.Sprintf("serviceQueue = %v\n", t.serviceQueue.elements), "token.log")
		time.Sleep(1 * time.Second)
	}
}
