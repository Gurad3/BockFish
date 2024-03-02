package main

import (
	"fmt"
	"time"
)

func benchsearch() {
	start := time.Now()
	// Code to measure
	currentBoard.searchMove()
	duration := time.Since(start)

	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	fmt.Println(duration)

}
