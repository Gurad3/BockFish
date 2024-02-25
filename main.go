package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var currentBoard board

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
			loadTables()
			fmt.Println("done")
		case "position":
			loadFen(parts[1:])
			fmt.Println("done")
		case "getlegal":
			fmt.Println(moveArrayTOstring(currentBoard.getLegalMoves(false)))

		case "getfullegal":
			fmt.Println(moveArrayTOstring(currentBoard.filterOutPsudoLegal(currentBoard.getLegalMoves(false))))

		case "play":
			//currentBoard.PlayMove(StringMoveToMove(parts[1]))
			currentBoard.PlayMove(tmpStringToMove[parts[1]])
			if tmpStringToMove[parts[1]]&CapFlag > 0 {
				writeToLogFile("PlayedMove was a capture")
			}
			tmpPlayedMoveList = append(tmpPlayedMoveList, tmpStringToMove[parts[1]])
			fmt.Println("done")

		case "test":
			loadTables()
			loadFen(strings.Split(startfen, " "))
			fmt.Println("legalmoves " + moveArrayTOstring(currentBoard.getLegalMoves(false)))
			//fmt.Println(KnightAttackMasks[18])

		case "undo":
			currentBoard.UndoMove(tmpPlayedMoveList[len(tmpPlayedMoveList)-1])
			tmpPlayedMoveList = tmpPlayedMoveList[:len(tmpPlayedMoveList)-1]
			fmt.Println("done")

		case "getpos":
			fmt.Println("pos " + sendPosAsString())

		case "searchmagic":
			fmt.Println("g")
			loadPreCalculations(true)
		case "go":
			// Implement search algorithm and find the best move
			// Once found, send the best move using the "bestmove" command
			fmt.Println("e2e4")

		case "stop":
			// Stop searching (if applicable)
		case "quit":
			os.Exit(0)

		case "exit":
			os.Exit(0)
		}
	}

}

func loadTables() {
	currentBoard = board{}
	loadPreCalculations(false)
}

func loadFen(fen []string) {
	currentBoard.loadFromFEN(fen)
}

var tmpPlayedMoveList []move = make([]move, 0, 100)
