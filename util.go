package main

import (
	"strconv"
	"strings"
)

//rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1

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

func (board *board) loadFromFEN(fen string) {
	stringParts := strings.Split(fen, " ")

	board.isWhiteTurn = stringParts[1] == "w"

	board.halfMoves, _ = strconv.Atoi(stringParts[4])
	board.fullMoves, _ = strconv.Atoi(stringParts[5])
	board.ply = board.fullMoves * 2
	if !board.isWhiteTurn {
		board.ply++
	}

	x := 0
	y := 0

	FenPos := stringParts[0]
	for i := range FenPos {
		st := string([]byte{FenPos[i]})
		if st == "/" {
			x = 0
			y++
			continue
		}
		if isNumber(st) {
			val, _ := strconv.Atoi(st)
			x += val
			continue
		}

		newPeice := new(piece)
		newPeice.typeID = stringToType[strings.ToLower(st)]
		newPeice.isWhite = st == strings.ToUpper(st)
		newPeice.square = uint8(x + y*8)

		board.squareArray[newPeice.square] = newPeice

		colID := 0
		if !newPeice.isWhite {
			colID = 1
		}

		board.pieceArrays[colID][newPeice.typeID] = append(board.pieceArrays[colID][newPeice.typeID], newPeice)

	}
	board.loadAllMasks()
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
