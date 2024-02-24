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

	pieceMask  [2]uint64
	pieceMasks [2][7]uint64

	pieceArrays [2][7][]*piece
	squareArray [64]*piece

	kings [2]*piece

	deadPieces []*piece

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
	typeID             uint8
	square             uint8
	isWhite            bool
	dead               bool
	indexInPiecesArray uint8
}

func (board *board) loadAllMasks() {
	for ColrID, ColpieceArrays := range board.pieceArrays {
		for PieceType, pieceArray := range ColpieceArrays {
			for _, piece := range pieceArray {
				board.pieceMasks[ColrID][PieceType] |= 1 << piece.square
				board.pieceMask[ColrID] |= 1 << piece.square
			}
		}
	}

}

const MoveSourceMask move = 0b0000000000111111
const MoveTargetMask move = 0b0000111111000000
const FlagMask move = 0b1111000000000000

const CapFlag move = 0b1000000000000000
const PromFlag move = 0b0111000000000000

func (board *board) PlayMove(Move move) {

	//board.pieceMask |= 1 << piece.square
	sourceIndex := Move & MoveSourceMask
	targetIndex := (Move & MoveTargetMask) >> 6

	targetPiece := board.squareArray[targetIndex]

	if !targetPiece.dead {
		opCOl := 0
		if !targetPiece.isWhite {
			opCOl = 1
		}
		board.pieceMasks[opCOl][targetPiece.typeID] ^= 1 << targetIndex
		board.pieceMask[opCOl] ^= 1 << targetIndex

		//Tar sista itemet i sliceen sätter den på index positionen och resizar

		//fmt.Println(targetPiece.indexInPiecesArray)
		board.pieceArrays[opCOl][targetPiece.typeID][targetPiece.indexInPiecesArray] = board.pieceArrays[opCOl][targetPiece.typeID][len(board.pieceArrays[opCOl][targetPiece.typeID])-1]

		//board.pieceArrays[opCOl][targetPiece.typeID][targetPiece.indexInPiecesArray].indexInPiecesArray = targetPiece.indexInPiecesArray
		board.pieceArrays[opCOl][targetPiece.typeID][len(board.pieceArrays[opCOl][targetPiece.typeID])-1].indexInPiecesArray = targetPiece.indexInPiecesArray

		board.pieceArrays[opCOl][targetPiece.typeID] = board.pieceArrays[opCOl][targetPiece.typeID][:len(board.pieceArrays[opCOl][targetPiece.typeID])-1]

		targetPiece.dead = true
		board.deadPieces = append(board.deadPieces, targetPiece)
	}

	sourcePiece := board.squareArray[sourceIndex]

	colr := 1
	if sourcePiece.isWhite {
		colr = 0
	}

	board.squareArray[targetIndex] = sourcePiece
	board.squareArray[sourceIndex] = targetPiece

	board.pieceMasks[colr][sourcePiece.typeID] ^= 1 << sourceIndex

	if (Move&PromFlag)>>12 > 3 {
		sourcePiece.typeID = uint8((Move&PromFlag)>>12) - 2
		board.pieceArrays[colr][sourcePiece.typeID] = append(board.pieceArrays[colr][sourcePiece.typeID], sourcePiece)

		board.pieceArrays[colr][1][sourcePiece.indexInPiecesArray] = board.pieceArrays[colr][1][len(board.pieceArrays[colr][1])-1]

		board.pieceArrays[colr][1][len(board.pieceArrays[colr][1])-1].indexInPiecesArray = sourcePiece.indexInPiecesArray
		//board.pieceArrays[colr][1][sourcePiece.indexInPiecesArray].indexInPiecesArray = sourcePiece.indexInPiecesArray

		board.pieceArrays[colr][1] = board.pieceArrays[colr][1][:len(board.pieceArrays[colr][1])-1]

		sourcePiece.indexInPiecesArray = uint8(len(board.pieceArrays[colr][sourcePiece.typeID]) - 1)
	}

	board.pieceMasks[colr][sourcePiece.typeID] |= 1 << targetIndex

	board.pieceMask[colr] ^= 1 << sourceIndex
	board.pieceMask[colr] |= 1 << targetIndex

	sourcePiece.square = uint8(targetIndex)
	board.isWhiteTurn = !board.isWhiteTurn

}

func (board *board) UndoMove(Move move) {

	sourceIndex := (Move & MoveTargetMask) >> 6
	targetIndex := Move & MoveSourceMask

	sourcePiece := board.squareArray[sourceIndex]
	targetPiece := board.squareArray[targetIndex]

	board.squareArray[targetIndex] = sourcePiece

	board.isWhiteTurn = !board.isWhiteTurn

	colr := 1
	opCOL := 0
	if board.isWhiteTurn {
		colr = 0
		opCOL = 1
	}

	board.squareArray[sourceIndex] = targetPiece

	//writeToLogFile("Revived from dead: " + strconv.Itoa(int(targetPiece.typeID)))

	if Move&CapFlag > 0 {

		targetPiece := board.deadPieces[len(board.deadPieces)-1]

		board.pieceArrays[opCOL][targetPiece.typeID] = append(board.pieceArrays[opCOL][targetPiece.typeID], targetPiece)
		targetPiece.indexInPiecesArray = uint8(len(board.pieceArrays[opCOL][targetPiece.typeID]) - 1)

		board.pieceMask[opCOL] |= 1 << sourceIndex
		board.pieceMasks[opCOL][targetPiece.typeID] |= 1 << sourceIndex
		targetPiece.dead = false

		targetPiece.square = uint8(sourceIndex)
		board.deadPieces = board.deadPieces[:len(board.deadPieces)-1]
		//writeToLogFile("Revived from dead: " + strconv.Itoa(int(targetPiece.typeID)) + "Pos: " + strconv.Itoa(int(sourceIndex)))
	}

	if (Move&PromFlag)>>12 > 3 {

		board.pieceArrays[colr][sourcePiece.typeID][sourcePiece.indexInPiecesArray] = board.pieceArrays[colr][sourcePiece.typeID][len(board.pieceArrays[colr][sourcePiece.typeID])-1]
		board.pieceArrays[colr][sourcePiece.typeID][sourcePiece.indexInPiecesArray].indexInPiecesArray = sourcePiece.indexInPiecesArray
		board.pieceArrays[colr][sourcePiece.typeID] = board.pieceArrays[colr][sourcePiece.typeID][:len(board.pieceArrays[colr][sourcePiece.typeID])-1]

		sourcePiece.typeID = uint8(1)
		board.pieceArrays[colr][sourcePiece.typeID] = append(board.pieceArrays[colr][sourcePiece.typeID], sourcePiece)
		sourcePiece.indexInPiecesArray = uint8(len(board.pieceArrays[colr][sourcePiece.typeID]) - 1)
	}
	board.pieceMasks[colr][sourcePiece.typeID] ^= 1 << sourceIndex
	board.pieceMasks[colr][sourcePiece.typeID] |= 1 << targetIndex

	board.pieceMask[colr] ^= 1 << sourceIndex
	board.pieceMask[colr] |= 1 << targetIndex

	sourcePiece.square = uint8(targetIndex)

}

func (board *board) clearDeadArray() {
	board.deadPieces = board.deadPieces[:0]
}
