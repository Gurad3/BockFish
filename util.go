package main

import (
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

		colID := 0
		if newPeice.isWhite {
			colID = 1
		}

		board.pieceArrays[colID][newPeice.typeID] = append(board.pieceArrays[colID][newPeice.typeID], newPeice)
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

		//fmt.Println(SIndex, TIndex, strconv.Itoa(int(SX)), moveString)
	}
	return retString
}

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king

*/
