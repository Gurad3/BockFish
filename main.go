package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
		// writeToLogFile(command)

		// Parse the command
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "uci":
			fmt.Println("id name Bock_Fish")
			fmt.Println("id author Gurad")
			// Add any supported options here
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			loadTables()
		case "position":

			if len(parts) > 3 && parts[2] == "moves" {

				StringMove := parts[len(parts)-1]
				Move := UCIMoveToEngineMove(StringMove)
				currentBoard.PlayMove(currentBoard.generateCapFlag(Move))

			} else {
				if parts[1] == "startpos" {
					loadFen(strings.Split(startfen, " "))
				} else {
					loadFen(parts[1:])
				}
			}

		case "go":
			fmt.Println(currentBoard.isWhiteTurn)

			move := currentBoard.searchMove()
			currentBoard.PlayMove(move)

			fmt.Println("bestmove " + engineMoveToUCIMove(move))

		case "getlegal":
			fmt.Println(moveArrayTOstring(currentBoard.getLegalMoves()))

		case "getfullegal":
			fmt.Println(moveArrayTOstring(currentBoard.filterOutPsudoLegal(currentBoard.getLegalMoves())))

		case "play":
			//currentBoard.PlayMove(StringMoveToMove(parts[1]))
			writeToLogFile("b " + strconv.Itoa(int(tmpStringToMove[parts[1]])))
			currentBoard.PlayMove(tmpStringToMove[parts[1]])
			if tmpStringToMove[parts[1]]&CapFlag > 0 {
				writeToLogFile("PlayedMove was a capture")
			}
			tmpPlayedMoveList = append(tmpPlayedMoveList, tmpStringToMove[parts[1]])
			fmt.Println("done")

		case "loaddef":
			loadTables()
			loadFen(strings.Split(startfen, " "))
			fmt.Println("legalmoves " + moveArrayTOstring(currentBoard.getLegalMoves()))
			//fmt.Println(KnightAttackMasks[18])

		case "benchsearch":
			benchsearch()

		case "undo":
			currentBoard.UndoMove(tmpPlayedMoveList[len(tmpPlayedMoveList)-1])
			tmpPlayedMoveList = tmpPlayedMoveList[:len(tmpPlayedMoveList)-1]
			fmt.Println("done")

		case "getpos":
			fmt.Println("pos " + sendPosAsString())

		case "plabest":
			mov := currentBoard.searchMove()
			currentBoard.PlayMove(mov)
			tmpPlayedMoveList = append(tmpPlayedMoveList, mov)
			fmt.Println("done")

		case "searchmagic":
			fmt.Println("g")
			loadPreCalculations(true)

		case "stop":
			// Stop searching (if applicable)
		case "quit":
			os.Exit(0)

		case "exit":
			os.Exit(0)

		case "tmp":
			tmp()
		}
	}

}

func loadTables() {
	currentBoard = board{transEntries: 50_000}
	currentBoard.transpostionTable = make([]tranEnrty, currentBoard.transEntries)
	loadPreCalculations(false)
	eval_loadTables()

	currentBoard.initZobristArray()
}

func loadFen(fen []string) {
	currentBoard.loadFromFEN(fen)
}

var tmpPlayedMoveList []move = make([]move, 0, 100)

func tmp() {
	loadTables()
	loadFen(strings.Split(startfen, " "))

	fmt.Println(currentBoard.pieceMasks[0][1])
	fmt.Println(currentBoard.pieceMasks[1][1])
}
