package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

var currentBoard board

func main() {

	errorLogFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open error log file:", err)
		os.Exit(1)
	}
	defer errorLogFile.Close()

	// Redirect standard error to the log file
	os.Stderr = errorLogFile

	// Recover from any panic and log the stack trace
	defer func() {
		if r := recover(); r != nil {
			// Write the panic reason and stack trace to the log file
			fmt.Fprintf(errorLogFile, "Program crashed: %v\n", r)
			errorLogFile.Write(debug.Stack())
			writeToLogFile("Erro " + Itoa(int(currentBoard.pieceMask[0])) + " " + Itoa(int(currentBoard.pieceMask[1])))

			for _, v := range currentBoard.gameStateStack {
				writeToLogFile(engineMoveToUCIMove(v.move))

			}
			writeToLogFile(Itoa(posEvaled))

		}
	}()

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
			t, _ := strconv.Atoi(parts[2])
			if currentBoard.ply <= 4 {
				t /= 3
			}
			move := currentBoard.searchMove(int64(t / 40))
			currentBoard.PlayMove(move)

			fmt.Println("bestmove " + engineMoveToUCIMove(move))
			writeToLogFile(Itoa(currentBoard.ply) + " " + Itoa(int(currentBoard.pieceMask[0])) + " " + Itoa(int(currentBoard.pieceMask[1])))

		case "getlegal":
			fmt.Println(moveArrayTOstring(currentBoard.getLegalMovesA()))

		case "getfullegal":
			fmt.Println(moveArrayTOstring(currentBoard.getLegalMovesA()))

			// fmt.Println(moveArrayTOstring(currentBoard.filterOutPsudoLegal(currentBoard.getLegalMovesA())))

		case "play":
			//currentBoard.PlayMove(StringMoveToMove(parts[1]))
			writeToLogFile("b " + strconv.Itoa(int(tmpStringToMove[parts[1]])))
			currentBoard.PlayMove(tmpStringToMove[parts[1]])
			if tmpStringToMove[parts[1]]&CapFlag > 0 {
				writeToLogFile("PlayedMove was a capture")
			}
			tmpPlayedMoveList = append(tmpPlayedMoveList, tmpStringToMove[parts[1]])
			fmt.Println("done")

		case "benchsearch":
			benchsearch()

		case "loaddef":
			loadTables()
			loadFen(strings.Split(startfen, " "))
			fmt.Println("legalmoves " + moveArrayTOstring(currentBoard.getLegalMovesA()))
			//fmt.Println(KnightAttackMasks[18])

		case "undo":
			currentBoard.UndoMove(tmpPlayedMoveList[len(tmpPlayedMoveList)-1])
			tmpPlayedMoveList = tmpPlayedMoveList[:len(tmpPlayedMoveList)-1]
			fmt.Println("done")

		case "getpos":
			writeToLogFile(Itoa(int(currentBoard.pieceMasks[0][4]|currentBoard.pieceMasks[1][4])) + " " + Itoa(int(currentBoard.pieceMasks[0][6]|currentBoard.pieceMasks[1][6])))
			fmt.Println("pos " + sendPosAsString())

		case "plabest":
			mov := currentBoard.searchMove(3000)
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
	currentBoard = board{transEntries: 1 << 20}
	currentBoard.transpostionTable = make([]tranEnrty, currentBoard.transEntries)
	currentBoard.castling = 0b1111
	// currentBoard.castling = 0b0000
	currentBoard.gameStateStack = make([]gameState, 0, 120)

	loadPreCalculations(false)
	eval_loadTables()

	currentBoard.initZobristArray()
}

func loadFen(fen []string) {
	currentBoard.loadFromFEN(fen)
}

var tmpPlayedMoveList []uint16 = make([]uint16, 0, 100)

func tmp() {
	loadTables()
	loadFen(strings.Split(startfen, " "))

	fmt.Println(currentBoard.pieceMasks[0][1])
	fmt.Println(currentBoard.pieceMasks[1][1])
}
