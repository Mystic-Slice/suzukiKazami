package main

import (
	"math/rand"
	"time"
)

const maxSleepTime = 10

func RandomSleep() {
	// Sleep for random amount of time
	var sleepTime int = rand.Intn(maxSleepTime)
	time.Sleep(time.Duration(sleepTime) * time.Second)
}
