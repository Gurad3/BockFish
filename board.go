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

	pieceMask  uint64
	pieceMasks [7]uint64

	pieceArray  [2][6][]*piece
	squareArray [64]*piece

	WQCastle bool
	BQCastle bool
	WKCastle bool
	BKCastle bool

	halfMoves int
	fullMoves int
	ply       int
}

type move uint16

type piece struct {
	typeID  uint8
	square  uint8
	isWhite bool
}

func (board *board) getLegalMoves(onlyCaptures bool) []move {
	return []move{}
}
