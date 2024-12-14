package main

import (
	"fmt"
	"strings"
	"time"
)

func benchsearch() {

	loadTables()
	loadFen(strings.Split(startfen, " "))

	start := time.Now()
	// Code to measure
	currentBoard.searchMove(9000)
	duration := time.Since(start)

	// Formatted string, such as "2h3m0.5s" or "4.503μs"
	fmt.Println(duration)

}
