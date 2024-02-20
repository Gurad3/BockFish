package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Loop to read commands from the GUI
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(0)
			continue
		}

		command = strings.TrimSpace(command)

		// Parse the command
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "uci":
			fmt.Println("id name YourChessEngineName")
			fmt.Println("id author YourName")
			// Add any supported options here
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			// Initialize any necessary data structures for a new game
		case "position":
			// Parse position command and update internal board representation
		case "go":
			// Implement search algorithm and find the best move
			// Once found, send the best move using the "bestmove" command
			fmt.Println("e2e4")
		case "stop":
			// Stop searching (if applicable)
		case "quit":
			os.Exit(0)
		}
	}
}
