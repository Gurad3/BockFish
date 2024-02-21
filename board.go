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
	pieceMasks [2][7]uint64

	pieceArrays [2][7][]*piece
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

func (board *board) loadAllMasks() {
	for ColrID, ColpieceArrays := range board.pieceArrays {
		for PieceType, pieceArray := range ColpieceArrays {
			for _, piece := range pieceArray {
				board.pieceMasks[ColrID][PieceType] |= 1 << piece.square
				board.pieceMask |= 1 << piece.square
			}
		}
	}
}

const MoveSourceMask move = 0b0000000000111111
const MoveTargetMask move = 0b0000111111000000
const FlagMask move = 0b1111000000000000

func (board *board) UpdateAllMasks(Move move) {

	//board.pieceMask |= 1 << piece.square

	sourceIndex := Move & MoveSourceMask
	targetIndex := (Move & MoveTargetMask) >> 6

	targetPiece := board.squareArray[targetIndex]
	if targetPiece.typeID != 0 {
		colr := 0
		if targetPiece.isWhite {
			colr = 1
		}
		board.pieceMasks[colr][targetPiece.typeID] ^= 1 << targetIndex
	}

	sourcePiece := board.squareArray[sourceIndex]
	board.squareArray[targetIndex] = sourcePiece
	board.squareArray[sourceIndex] = new(piece)

	colr := 0
	if sourcePiece.isWhite {
		colr = 1
	}
	board.pieceMasks[colr][sourcePiece.typeID] ^= 1 << sourceIndex
	board.pieceMasks[colr][sourcePiece.typeID] |= 1 << targetIndex

}
