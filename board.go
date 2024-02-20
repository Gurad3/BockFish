package main

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king

*/

type board struct {
	isWhiteTurn bool
	zobrist     uint64

	pieceMask uint64

	pieceMasks [7]uint64

	pieceArray [32]piece
}

type move struct {
	isCapture bool

	sourceSquare uint64
	targetSquare uint64
}

type piece struct {
	typeID  uint8
	square  uint64
	isWhite bool
}

func (board *board) getLegalMoves(onlyCaptures bool) []move {
	return []move{}
}
