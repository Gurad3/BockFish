package main

import (
	"math/rand"
)

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

	pieceMask  [2]uint64
	pieceMasks [2][7]uint64

	pieceArrays [2][7][]*piece
	squareArray [64]*piece

	kings [2]*piece

	deadPieces []uint8

	castling       uint8
	gameStateStack []gameState

	halfMoves int
	fullMoves int
	ply       int

	zobrist      uint64
	zobristVals  [2][7][64]uint64 // empty +
	zobristBlack uint64

	transEntries      int
	transpostionTable []tranEnrty
}

type gameState struct {
	castling uint8
}

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

const WKCastleMask uint64 = 0b1111000000000000000000000000000000000000000000000000000000000000
const WKCastleFlag uint64 = 0b1001000000000000000000000000000000000000000000000000000000000000

const WQCastleMask uint64 = 0b1111100000000000000000000000000000000000000000000000000000000
const WQCastleFlag uint64 = 0b1000100000000000000000000000000000000000000000000000000000000

const BKCastleMask uint64 = 0b11110000
const BKCastleFlag uint64 = 0b10010000

const BQCastleMask uint64 = 0b11111
const BQCastleFlag uint64 = 0b10001

const MoveSourceMask uint16 = 0b0000000000111111
const MoveTargetMask uint16 = 0b0000111111000000
const FlagMask uint16 = 0b1111000000000000

const CapFlag uint16 = 0b1000000000000000

const isPromFlag uint16 = 0b0100000000000000
const PromFlag uint16 = 0b0111000000000000

const NoFlag uint16 = 0b0000 << 12
const EnPassant uint16 = 0b0001 << 12
const CastleFlag uint16 = 0b0010 << 12
const PawnTwoUpFlag uint16 = 0b0011 << 12

const PromoteToKnightFlag uint16 = 0b0100
const PromoteToBishopFlag uint16 = 0b0101
const PromoteToRookFlag uint16 = 0b0110
const PromoteToQueenFlag uint16 = 0b0111

func (board *board) PlayMove(Move uint16) {

	//board.pieceMask |= 1 << piece.square
	sourceIndex := Move & MoveSourceMask
	sourcePiece := board.squareArray[sourceIndex]

	targetIndex := (Move & MoveTargetMask) >> 6
	targetPiece := board.squareArray[targetIndex]

	newGameState := gameState{}
	newGameState.castling = board.castling

	board.gameStateStack = append(board.gameStateStack, newGameState)
	srcCol := 1
	if sourcePiece.isWhite {
		srcCol = 0

		if sourcePiece.typeID == 6 {
			board.castling &= 0b0011
		} else if sourcePiece.typeID == 4 {
			if sourceIndex == 56 {
				board.castling &= 0b0111
			} else if sourceIndex == 63 {
				board.castling &= 0b1011
			}
		}
		if targetPiece.typeID == 4 {
			if targetIndex == 0 {
				board.castling &= 0b1101
			} else if targetIndex == 7 {
				board.castling &= 0b1110
			}
		}
	} else {
		if sourcePiece.typeID == 6 {
			board.castling &= 0b1100
		} else if sourcePiece.typeID == 4 {
			if sourceIndex == 0 {
				board.castling &= 0b1101
			} else if sourceIndex == 7 {
				board.castling &= 0b1110
			}
		}
		if targetPiece.typeID == 4 {
			if targetIndex == 56 {
				board.castling &= 0b0111
			} else if targetIndex == 63 {
				board.castling &= 0b1011
			}
		}
	}
	board.zobrist ^= board.zobristVals[srcCol][sourcePiece.typeID][sourceIndex]
	board.zobrist ^= board.zobristVals[srcCol][sourcePiece.typeID][targetIndex]
	board.zobrist ^= board.zobristBlack

	//Remove SourceBits.

	if Move&CapFlag > 0 {
		opCOl := 0
		if !targetPiece.isWhite {
			opCOl = 1
		}

		board.pieceMasks[opCOl][targetPiece.typeID] ^= 1 << targetIndex
		board.pieceMask[opCOl] ^= 1 << targetIndex

		board.zobrist ^= board.zobristVals[opCOl][targetPiece.typeID][targetIndex]

		//Tar sista itemet i sliceen sätter den på index positionen och resizar
		if len(board.pieceArrays[opCOl][targetPiece.typeID])-1 == -1 {
			writeToLogFile("--")
			writeToLogFile("hum")
			writeToLogFile(Itoa(int(targetPiece.typeID)))
			writeToLogFile(Itoa(int(sourcePiece.typeID)))
			if targetPiece.dead {
				writeToLogFile("true")

			} else {
				writeToLogFile("false")

			}
			writeToLogFile(Itoa(int(sourceIndex)) + " " + Itoa(int(targetIndex)))
			writeToLogFile(Itoa(board.ply))
		}
		board.pieceArrays[opCOl][targetPiece.typeID][targetPiece.indexInPiecesArray] = board.pieceArrays[opCOl][targetPiece.typeID][len(board.pieceArrays[opCOl][targetPiece.typeID])-1]
		board.pieceArrays[opCOl][targetPiece.typeID][targetPiece.indexInPiecesArray].indexInPiecesArray = targetPiece.indexInPiecesArray

		//board.pieceArrays[opCOl][targetPiece.typeID][len(board.pieceArrays[opCOl][targetPiece.typeID])-1].indexInPiecesArray = targetPiece.indexInPiecesArray
		board.pieceArrays[opCOl][targetPiece.typeID] = board.pieceArrays[opCOl][targetPiece.typeID][:len(board.pieceArrays[opCOl][targetPiece.typeID])-1]

		targetPiece.dead = true
		board.deadPieces = append(board.deadPieces, targetPiece.typeID)

		targetPiece.typeID = 0
	}
	board.pieceMasks[srcCol][sourcePiece.typeID] ^= 1 << sourceIndex
	board.pieceMask[srcCol] ^= 1 << sourceIndex

	if (Move&PromFlag)>>12 > 3 {
		sourcePiece.typeID = uint8((Move&PromFlag)>>12) - 2
		board.pieceArrays[srcCol][sourcePiece.typeID] = append(board.pieceArrays[srcCol][sourcePiece.typeID], sourcePiece)
		board.pieceArrays[srcCol][1][len(board.pieceArrays[srcCol][1])-1].indexInPiecesArray = sourcePiece.indexInPiecesArray
		board.pieceArrays[srcCol][1][sourcePiece.indexInPiecesArray] = board.pieceArrays[srcCol][1][len(board.pieceArrays[srcCol][1])-1]
		board.pieceArrays[srcCol][1] = board.pieceArrays[srcCol][1][:len(board.pieceArrays[srcCol][1])-1]
		sourcePiece.indexInPiecesArray = uint8(len(board.pieceArrays[srcCol][sourcePiece.typeID]) - 1)
	}

	if (Move & FlagMask) == CastleFlag {
		if targetIndex == 0 {
			board.pieceMask[srcCol] ^= (1 << 0) // Rook, king already gone
			board.pieceMasks[srcCol][4] ^= (1 << 0)

			board.pieceMask[srcCol] |= (1 << 2)
			board.pieceMask[srcCol] |= (1 << 3)
			board.pieceMasks[srcCol][6] |= (1 << 2)
			board.pieceMasks[srcCol][4] |= (1 << 3)

			board.squareArray[4] = board.squareArray[2] // King square = empty TargetSqare
			board.squareArray[2] = sourcePiece
			board.squareArray[0] = board.squareArray[3]
			board.squareArray[3] = targetPiece

			sourcePiece.square = 2
			targetPiece.square = 3

			// src -> sq := 2
			// trg -> sq := 3
		} else if targetIndex == 7 {

			board.pieceMask[srcCol] ^= (1 << 7) // Rook, king already gone
			board.pieceMasks[srcCol][4] ^= (1 << 7)

			board.pieceMask[srcCol] |= (1 << 6)
			board.pieceMask[srcCol] |= (1 << 5)
			board.pieceMasks[srcCol][6] |= (1 << 6)
			board.pieceMasks[srcCol][4] |= (1 << 5)

			board.squareArray[4] = board.squareArray[6] // King square = empty TargetSqare
			board.squareArray[6] = sourcePiece
			board.squareArray[7] = board.squareArray[5]
			board.squareArray[5] = targetPiece

			sourcePiece.square = 6
			targetPiece.square = 5

			// src -> sq := 6
			// trg -> sq := 5
		} else if targetIndex == 56 {
			board.pieceMask[srcCol] ^= (1 << 56) // Rook, king already gone
			board.pieceMasks[srcCol][4] ^= (1 << 56)

			board.pieceMask[srcCol] |= (1 << 58)
			board.pieceMask[srcCol] |= (1 << 59)
			board.pieceMasks[srcCol][6] |= (1 << 58)
			board.pieceMasks[srcCol][4] |= (1 << 59)

			board.squareArray[60] = board.squareArray[58] // King square = empty TargetSqare
			board.squareArray[58] = sourcePiece
			board.squareArray[56] = board.squareArray[59]
			board.squareArray[59] = targetPiece

			sourcePiece.square = 58
			targetPiece.square = 59
			// src -> sq := 58
			// trg -> sq := 59

		} else if targetIndex == 63 {

			board.pieceMask[srcCol] ^= (1 << 63) // Rook, king already gone
			board.pieceMasks[srcCol][4] ^= (1 << 63)

			board.pieceMask[srcCol] |= (1 << 62)
			board.pieceMask[srcCol] |= (1 << 61)
			board.pieceMasks[srcCol][6] |= (1 << 62)
			board.pieceMasks[srcCol][4] |= (1 << 61)

			board.squareArray[60] = board.squareArray[62] // King square = empty TargetSqare
			board.squareArray[62] = sourcePiece
			board.squareArray[63] = board.squareArray[61]
			board.squareArray[61] = targetPiece

			sourcePiece.square = 62
			targetPiece.square = 61
			// src -> sq := 62
			// trg -> sq := 61
		}

	} else {
		board.squareArray[targetIndex] = sourcePiece
		board.squareArray[sourceIndex] = targetPiece

		//Update (add if empty) TargetBits
		board.pieceMasks[srcCol][sourcePiece.typeID] |= 1 << targetIndex
		board.pieceMask[srcCol] |= 1 << targetIndex

		sourcePiece.square = uint8(targetIndex)
		targetPiece.square = uint8(sourceIndex)
	}

	board.isWhiteTurn = !board.isWhiteTurn
	board.ply++

}

func (board *board) UndoMove(Move uint16) {

	sourceIndex := (Move & MoveTargetMask) >> 6
	targetIndex := Move & MoveSourceMask

	sourcePiece := board.squareArray[sourceIndex]
	targetPiece := board.squareArray[targetIndex]

	srcCol := 1
	opCOL := 0
	if !board.isWhiteTurn {
		srcCol = 0
		opCOL = 1

	}
	board.castling = board.gameStateStack[len(board.gameStateStack)-1].castling
	board.gameStateStack = board.gameStateStack[:len(board.gameStateStack)-1]

	board.zobrist ^= board.zobristVals[srcCol][sourcePiece.typeID][sourceIndex]
	board.zobrist ^= board.zobristVals[srcCol][sourcePiece.typeID][targetIndex]
	board.zobrist ^= board.zobristBlack

	if Move&CapFlag > 0 {

		targetPiece.typeID = board.deadPieces[len(board.deadPieces)-1]

		board.zobrist ^= board.zobristVals[opCOL][targetPiece.typeID][sourceIndex]

		board.deadPieces = board.deadPieces[:len(board.deadPieces)-1]

		board.pieceArrays[opCOL][targetPiece.typeID] = append(board.pieceArrays[opCOL][targetPiece.typeID], targetPiece)
		targetPiece.indexInPiecesArray = uint8(len(board.pieceArrays[opCOL][targetPiece.typeID]) - 1)

		board.pieceMask[opCOL] |= 1 << sourceIndex
		board.pieceMasks[opCOL][targetPiece.typeID] |= 1 << sourceIndex
		targetPiece.dead = false

		//writeToLogFile("Revived from dead: " + strconv.Itoa(int(targetPiece.typeID)) + "Pos: " + strconv.Itoa(int(sourceIndex)))
	}

	board.pieceMasks[srcCol][sourcePiece.typeID] ^= 1 << sourceIndex
	board.pieceMask[srcCol] ^= 1 << sourceIndex

	if (Move&PromFlag)>>12 > 3 {

		board.pieceArrays[srcCol][sourcePiece.typeID][sourcePiece.indexInPiecesArray] = board.pieceArrays[srcCol][sourcePiece.typeID][len(board.pieceArrays[srcCol][sourcePiece.typeID])-1]
		board.pieceArrays[srcCol][sourcePiece.typeID][sourcePiece.indexInPiecesArray].indexInPiecesArray = sourcePiece.indexInPiecesArray
		board.pieceArrays[srcCol][sourcePiece.typeID] = board.pieceArrays[srcCol][sourcePiece.typeID][:len(board.pieceArrays[srcCol][sourcePiece.typeID])-1]

		sourcePiece.typeID = uint8(1)
		board.pieceArrays[srcCol][sourcePiece.typeID] = append(board.pieceArrays[srcCol][sourcePiece.typeID], sourcePiece)
		sourcePiece.indexInPiecesArray = uint8(len(board.pieceArrays[srcCol][sourcePiece.typeID]) - 1)
	}

	if (Move & FlagMask) == CastleFlag {

		//Undo
		board.pieceMasks[srcCol][sourcePiece.typeID] ^= 1 << sourceIndex
		board.pieceMask[srcCol] ^= 1 << sourceIndex

		if sourceIndex == 0 { // Rook == Source

			board.pieceMask[srcCol] |= (1 << 4)
			board.pieceMasks[srcCol][6] |= (1 << 4)
			board.pieceMask[srcCol] |= (1 << 0)
			board.pieceMasks[srcCol][4] |= (1 << 0)

			board.pieceMask[srcCol] ^= (1 << 2)
			board.pieceMask[srcCol] ^= (1 << 3)
			board.pieceMasks[srcCol][6] ^= (1 << 2)
			board.pieceMasks[srcCol][4] ^= (1 << 3)

			board.squareArray[4] = board.squareArray[2] //King
			board.squareArray[2] = targetPiece          //NotKing King
			board.squareArray[0] = board.squareArray[3] //Rook
			board.squareArray[3] = sourcePiece          //NotRook

			board.squareArray[0].square = 0
			board.squareArray[4].square = 4

		} else if sourceIndex == 7 {

			board.pieceMask[srcCol] |= (1 << 4)
			board.pieceMasks[srcCol][6] |= (1 << 4)
			board.pieceMask[srcCol] |= (1 << 7)
			board.pieceMasks[srcCol][4] |= (1 << 7)

			board.pieceMask[srcCol] ^= (1 << 6)
			board.pieceMask[srcCol] ^= (1 << 5)
			board.pieceMasks[srcCol][6] ^= (1 << 6)
			board.pieceMasks[srcCol][4] ^= (1 << 5)

			board.squareArray[4] = board.squareArray[6] //King
			board.squareArray[6] = targetPiece          //NotKing King
			board.squareArray[7] = board.squareArray[5] //Rook
			board.squareArray[5] = sourcePiece          //NotRook

			board.squareArray[7].square = 7
			board.squareArray[4].square = 4

			// src -> sq := 6
			// trg -> sq := 5
		} else if sourceIndex == 56 {
			board.pieceMask[srcCol] |= (1 << 60)
			board.pieceMasks[srcCol][6] |= (1 << 60)
			board.pieceMask[srcCol] |= (1 << 56)
			board.pieceMasks[srcCol][4] |= (1 << 56)

			board.pieceMask[srcCol] ^= (1 << 58)
			board.pieceMask[srcCol] ^= (1 << 59)
			board.pieceMasks[srcCol][6] ^= (1 << 58)
			board.pieceMasks[srcCol][4] ^= (1 << 59)

			board.squareArray[60] = board.squareArray[58] //King
			board.squareArray[58] = targetPiece           //NotKing King
			board.squareArray[56] = board.squareArray[59] //Rook
			board.squareArray[59] = sourcePiece           //NotRook

			board.squareArray[56].square = 56
			board.squareArray[60].square = 60

			// src -> sq := 58
			// trg -> sq := 59
		} else if sourceIndex == 63 {

			board.pieceMask[srcCol] |= (1 << 60)
			board.pieceMasks[srcCol][6] |= (1 << 60)
			board.pieceMask[srcCol] |= (1 << 63)
			board.pieceMasks[srcCol][4] |= (1 << 63)

			board.pieceMask[srcCol] ^= (1 << 62)
			board.pieceMask[srcCol] ^= (1 << 61)
			board.pieceMasks[srcCol][6] ^= (1 << 62)
			board.pieceMasks[srcCol][4] ^= (1 << 61)

			board.squareArray[60] = board.squareArray[62] //King
			board.squareArray[62] = targetPiece           //NotKing King
			board.squareArray[63] = board.squareArray[61] //Rook
			board.squareArray[61] = sourcePiece           //NotRook

			board.squareArray[63].square = 63
			board.squareArray[60].square = 60

			// src -> sq := 62
			// trg -> sq := 61
		}

	} else {
		board.pieceMasks[srcCol][sourcePiece.typeID] |= 1 << targetIndex
		board.pieceMask[srcCol] |= 1 << targetIndex

		board.squareArray[targetIndex] = sourcePiece
		board.squareArray[sourceIndex] = targetPiece

		sourcePiece.square = uint8(targetIndex)
		targetPiece.square = uint8(sourceIndex)

	}

	board.ply--
	board.isWhiteTurn = !board.isWhiteTurn

	// writeToLogFile(strconv.Itoa(int(board.pieceMask[0])))
	// writeToLogFile(strconv.Itoa(int(board.pieceMask[1])))

}

func (board *board) initZobristArray() {
	for col := 0; col < 2; col++ {
		for p := 1; p < 7; p++ { //first = empty
			for sq := 0; sq < 64; sq++ {
				board.zobristVals[col][p][sq] = rand.Uint64()
			}
		}
	}
	board.zobristBlack = rand.Uint64()
}

func (board *board) calcZobrist() {
	for col := 0; col < 2; col++ {
		for i := 1; i < 7; i++ {
			for pID := 0; pID < len(board.pieceArrays[col][i]); pID++ {
				currPiece := board.pieceArrays[col][i][pID]
				piecCol := 1
				if currPiece.isWhite {
					piecCol = 0
				}
				board.zobrist ^= board.zobristVals[piecCol][currPiece.typeID][currPiece.square]
			}
		}
	}
}
