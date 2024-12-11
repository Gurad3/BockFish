package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var startfen string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

var stringToType = map[string]uint8{
	"p": 1,
	"n": 2,
	"b": 3,
	"r": 4,
	"q": 5,
	"k": 6,
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (board *board) loadFromFEN(fen []string) {
	board.deadPieces = make([]uint8, 0, 30)

	board.isWhiteTurn = fen[1] == "w"

	board.halfMoves, _ = strconv.Atoi(fen[4])
	board.fullMoves, _ = strconv.Atoi(fen[5])
	board.ply = board.fullMoves * 2
	if !board.isWhiteTurn {
		board.ply++
	}

	x := 0
	y := 0

	FenPos := fen[0]
	for i := range FenPos {
		st := string([]byte{FenPos[i]})
		if st == "/" {
			x = 0
			y++
			continue
		}
		if isNumber(st) {
			val, _ := strconv.Atoi(st)

			for i := 0; i < val; i++ {
				newPeice := new(piece)
				newPeice.typeID = 0
				newPeice.isWhite = true
				newPeice.square = uint8(x + y*8)
				newPeice.dead = true
				board.squareArray[newPeice.square] = newPeice
				x++
			}

			continue
		}

		newPeice := new(piece)
		newPeice.typeID = stringToType[strings.ToLower(st)]
		newPeice.isWhite = st == strings.ToUpper(st)
		newPeice.square = uint8(x + y*8)
		newPeice.dead = false
		board.squareArray[newPeice.square] = newPeice

		if st == "K" {
			board.kings[0] = newPeice
		}
		if st == "k" {
			board.kings[1] = newPeice
		}

		colID := 1
		if newPeice.isWhite {
			colID = 0
		}

		board.pieceArrays[colID][newPeice.typeID] = append(board.pieceArrays[colID][newPeice.typeID], newPeice)
		newPeice.indexInPiecesArray = uint8(len(board.pieceArrays[colID][newPeice.typeID]) - 1)
		x++
	}
	board.loadAllMasks()
}

func StringMoveToMove(mo string) move {
	newMove := move(0)
	splited := strings.Split(mo, "")
	SX, _ := strconv.Atoi(splited[0])
	SY, _ := strconv.Atoi(splited[1])

	TX, _ := strconv.Atoi(splited[2])
	TY, _ := strconv.Atoi(splited[3])

	newMove += move(TX) + move(TY)*8
	newMove <<= 6

	newMove += move(SX) + move(SY)*8
	return newMove
}

func moveArrayTOstring(array []move) string {
	retString := ""
	for _, Move := range array {

		SIndex := (MoveSourceMask & Move)
		TIndex := (MoveTargetMask & Move) >> 6

		moveString := ""

		SX := SIndex % 8
		SY := SIndex / 8
		TX := TIndex % 8
		TY := TIndex / 8

		moveString += strconv.Itoa(int(SX))
		moveString += strconv.Itoa(int(SY))
		moveString += strconv.Itoa(int(TX))
		moveString += strconv.Itoa(int(TY))

		retString += moveString + " "
		tmpStringToMove[moveString] = Move
		//fmt.Println(SIndex, TIndex, strconv.Itoa(int(SX)), moveString)
	}
	return retString
}
func sendPosAsString() string {
	retString := ""
	for i := 0; i < 64; i++ {
		partString := ""
		SX := i % 8
		SY := i / 8
		partString += strconv.Itoa(int(SX))
		partString += strconv.Itoa(int(SY))
		if currentBoard.squareArray[i].isWhite {
			partString += "w"
		} else {
			partString += "b"
		}
		if currentBoard.squareArray[i].dead {
			partString += "0"
		} else {
			partString += strconv.Itoa(int(currentBoard.squareArray[i].typeID))
		}
		retString += partString + " "
	}

	return retString
}

var tmpStringToMove map[string]move = make(map[string]move, 1000)

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king

*/

func writeToLogFile(message string) {
	// Open the file in append mode. Create it if it doesn't exist.
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a logger that writes to the file
	logger := log.New(file, "", log.LstdFlags)

	// Write the log message
	logger.Println(message)
}

func intMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a int, b int) int {
	if a > b {
		return a
	}

	return b
}
