package main

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

type Site struct {
	siteId                  int         // id of this site
	receivedSequenceNumbers []int       // Sequence number of the last message received from each site
	sendChannels            []chan int  // Channels to send messages to other sites
	listenChannels          []chan int  // Channels to listen to messages from other sites
	token                   *Token      // Token used for critical section
	internalLock            *sync.Mutex // Lock for the site -> Use when modifying any of site fields
}

// Initializes the site fields
func (s *Site) initialize(siteId int, sendChannels []chan int, listenChannels []chan int, token *Token) {
	s.siteId = siteId
	s.receivedSequenceNumbers = make([]int, numberOfSites)
	for i := 0; i < numberOfSites; i++ {
		s.receivedSequenceNumbers[i] = -1
	}
	s.sendChannels = sendChannels
	s.listenChannels = listenChannels
	s.token = token
	s.internalLock = &sync.Mutex{}
}

// Logs the state of the site every second
func (s *Site) logger() {
	for {
		log(
			fmt.Sprintf("receivedSequenceNumbers = %v\n", s.receivedSequenceNumbers),
			"site"+strconv.Itoa(s.siteId)+".log",
		)
		time.Sleep(1 * time.Second)
	}
}

func (s *Site) run() {
	// Start listening to all sites
	go s.listenerDaemonAll()

	for {
		// Execute non-critical section
		s.executeNonCriticalSection()

		// Request critical section
		s.requestCriticalSection()

		// Execute critical section
		s.WaitAndExecuteCriticalSection()

		// Release critical section
		s.releaseCriticalSection()

		log(
			"Completed cycle\n",
			"site"+strconv.Itoa(s.siteId)+".log",
		)
	}
}

// Executes the non-critical section
func (s *Site) executeNonCriticalSection() {
	log(
		"Executing non-critical section\n",
		"site"+strconv.Itoa(s.siteId)+".log",
	)
	RandomSleep()
}

// Requests the critical section by
// broadcasting sequenceNumber to all sites
func (s *Site) requestCriticalSection() {
	log(
		"Requesting critical section\n",
		"site"+strconv.Itoa(s.siteId)+".log",
	)

	s.internalLock.Lock()
	s.receivedSequenceNumbers[s.siteId]++
	s.internalLock.Unlock()

	// Send broadcast message to all sites
	for i := 0; i < numberOfSites; i++ {
		if i == s.siteId {
			continue
		}
		log(
			fmt.Sprintf("Sending message %d to site %d\n", s.receivedSequenceNumbers[s.siteId], i),
			"site"+strconv.Itoa(s.siteId)+".log",
		)
		s.sendChannels[i] <- s.receivedSequenceNumbers[s.siteId]
	}
}

// Waits for the token to be sent to this site
// Token lock not possessed on start
// Token lock possessed at the end
func (s *Site) WaitAndExecuteCriticalSection() {
	// Wait for token
	log(
		"Waiting for token\n",
		"site"+strconv.Itoa(s.siteId)+".log",
	)
	for {
		// Check if token is sent to this site
		if s.token.currentOwner == s.siteId {
			// If token is available, break out of loop
			s.token.lock.Lock()
			log(
				"Received token\n",
				"site"+strconv.Itoa(s.siteId)+".log",
			)
			break
		}
	}

	// Execute critical section
	log(
		"Executing critical section\n",
		"site"+strconv.Itoa(s.siteId)+".log",
	)
	RandomSleep()
}

// Releases the critical section by updating token
// and starting the tokenReleaseDaemon
// Token lock already possessed on start
// Token lock still possessed at the end
// (will be released later by tokenReleaseDaemon)
func (s *Site) releaseCriticalSection() {
	log(
		"Releasing critical section\n",
		"site"+strconv.Itoa(s.siteId)+".log",
	)
	s.internalLock.Lock()
	s.token.lastServicedSequenceNumber[s.siteId] = s.receivedSequenceNumbers[s.siteId]
	s.internalLock.Unlock()
	go s.tokenReleaseDaemon()
}

// Releases the token to the next site in the serviceQueue
// Token lock already possessed on start
// Token lock released at the end
//
// Reason for tokenReleaseDaemon:
//
//		Incase no other site is waiting for the token,
//		the `releaseCriticalSection` method will have to wait till a site
//		requests the token and then release it to that site.
//		This will prevent the same site from accesssing the critical section
//		multiple times in a row if it wants to
//	 	`tokenReleaseDaemon` solves this problem by working in the background
//		and not blocking the "main" execution path of the current site
func (s *Site) tokenReleaseDaemon() {
	// Release token on return
	defer func() {
		s.token.lock.Unlock()
		log(
			"Token lock released\n",
			"site"+strconv.Itoa(s.siteId)+".log",
		)
	}()

	for {
		s.internalLock.Lock()
		// Check if there are any other sites to service
		for i := 0; i < numberOfSites; i++ {
			if i == s.siteId {
				continue
			}

			if s.receivedSequenceNumbers[i] == s.token.lastServicedSequenceNumber[i]+1 {
				s.token.serviceQueue.PushUnique(i)
			}
		}
		// Checked at end, to ensure fairness
		if s.receivedSequenceNumbers[s.siteId] == s.token.lastServicedSequenceNumber[s.siteId]+1 {
			s.token.serviceQueue.PushUnique(s.siteId)
		}
		log(
			fmt.Sprintf("Updated serviceQueue = %v\n", s.token.serviceQueue.elements),
			"site"+strconv.Itoa(s.siteId)+".log",
		)

		if !s.token.serviceQueue.IsEmpty() {
			nextOwner := s.token.serviceQueue.Pop()
			s.token.currentOwner = nextOwner
			log(
				fmt.Sprintf("Sending token to site %d\n", nextOwner),
				"site"+strconv.Itoa(s.siteId)+".log",
			)
			s.internalLock.Unlock()
			return
		}
		s.internalLock.Unlock()
	}
}

// Starts a listener daemon for each sites
func (s *Site) listenerDaemonAll() {
	for i := 0; i < numberOfSites; i++ {
		if i == s.siteId {
			continue
		}
		go s.listenerDaemonSingle(i)
	}
}

// Listens for any messages from site with siteId `listenToSiteId`
// and updates the receivedSequenceNumbers accordingly
func (s *Site) listenerDaemonSingle(listenToSiteId int) {
	for {
		// Receive message from siteId
		message := <-s.listenChannels[listenToSiteId]
		log(
			fmt.Sprintf("Received message %d from site %d\n", message, listenToSiteId),
			"site"+strconv.Itoa(s.siteId)+".log",
		)

		s.internalLock.Lock()
		s.receivedSequenceNumbers[listenToSiteId] = int(
			math.Max(float64(s.receivedSequenceNumbers[listenToSiteId]), float64(message)),
		)
		s.internalLock.Unlock()

		log(
			fmt.Sprintf("Updated receivedSequenceNumbers = %v\n", s.receivedSequenceNumbers),
			"site"+strconv.Itoa(s.siteId)+".log",
		)
	}
}
