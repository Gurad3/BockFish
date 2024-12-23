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

	transEntries          uint64
	transpostionTable     []tranEntry
	quinTranspostionTable []tranEntry
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

// Castling
const castle_WKS uint16 = 60  //White King Source
const castle_WRKS uint16 = 63 //White Rook KingSide Source
const castle_WRQS uint16 = 56 //White Rook QueenSide Source

const castle_WKKT uint16 = 62 //White King KingSide Target
const castle_WKQT uint16 = 58 //White King QueenSide Target

const castle_WRKT uint16 = 61 //White Rook KingSide Target
const castle_WRQT uint16 = 59 //White Rook QueenSide Target

const castle_BKS uint16 = 4  //Black King Source
const castle_BRKS uint16 = 7 //Black Rook KingSide Source
const castle_BRQS uint16 = 0 //Black Rook QueenSide Source

const castle_BKKT uint16 = 6 //Black King KingSide Target
const castle_BKQT uint16 = 2 //Black King QueenSide Target

const castle_BRKT uint16 = 5 //Black Rook KingSide Target
const castle_BRQT uint16 = 3 //Black Rook QueenSide Target

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
	opCol := 0
	if sourcePiece.isWhite {
		srcCol = 0
		opCol = 1

		if sourcePiece.typeID == 6 {
			board.castling &= 0b0011
		} else if sourcePiece.typeID == 4 {
			if sourceIndex == castle_WRQS {
				board.castling &= 0b0111
			} else if sourceIndex == castle_WRKS {
				board.castling &= 0b1011
			}
		}
		if targetPiece.typeID == 4 {
			if targetIndex == castle_BRQS {
				board.castling &= 0b1101
			} else if targetIndex == castle_BRKS {
				board.castling &= 0b1110
			}
		}
	} else {
		if sourcePiece.typeID == 6 {
			board.castling &= 0b1100
		} else if sourcePiece.typeID == 4 {
			if sourceIndex == castle_BRQS {
				board.castling &= 0b1101
			} else if sourceIndex == castle_BRKS {
				board.castling &= 0b1110
			}
		}
		if targetPiece.typeID == 4 {
			if targetIndex == castle_WRQS {
				board.castling &= 0b0111
			} else if targetIndex == castle_WRKS {
				board.castling &= 0b1011
			}
		}
	}
	board.zobrist ^= board.zobristVals[srcCol][sourcePiece.typeID][sourceIndex]
	board.zobrist ^= board.zobristVals[srcCol][sourcePiece.typeID][targetIndex]
	board.zobrist ^= board.zobristBlack

	//Remove SourceBits.

	board.pieceMasks[srcCol][sourcePiece.typeID] ^= 1 << sourceIndex
	board.pieceMask[srcCol] ^= 1 << sourceIndex

	if Move&CapFlag > 0 {

		board.pieceMasks[opCol][targetPiece.typeID] ^= 1 << targetIndex
		board.pieceMask[opCol] ^= 1 << targetIndex

		board.zobrist ^= board.zobristVals[opCol][targetPiece.typeID][targetIndex]

		// if len(board.pieceArrays[opCol][targetPiece.typeID])-1 == -1 {
		// 	writeToLogFile("--")
		// 	writeToLogFile("hum")
		// 	writeToLogFile(Itoa(int(targetPiece.typeID)))
		// 	writeToLogFile(Itoa(int(sourcePiece.typeID)))
		// }

		//Tar sista itemet i sliceen sätter den på index positionen och resizar
		board.pieceArrays[opCol][targetPiece.typeID][targetPiece.indexInPiecesArray] = board.pieceArrays[opCol][targetPiece.typeID][len(board.pieceArrays[opCol][targetPiece.typeID])-1]
		board.pieceArrays[opCol][targetPiece.typeID][targetPiece.indexInPiecesArray].indexInPiecesArray = targetPiece.indexInPiecesArray

		//board.pieceArrays[opCOl][targetPiece.typeID][len(board.pieceArrays[opCOl][targetPiece.typeID])-1].indexInPiecesArray = targetPiece.indexInPiecesArray
		board.pieceArrays[opCol][targetPiece.typeID] = board.pieceArrays[opCol][targetPiece.typeID][:len(board.pieceArrays[opCol][targetPiece.typeID])-1]

		targetPiece.dead = true
		board.deadPieces = append(board.deadPieces, targetPiece.typeID)

		targetPiece.typeID = 0
	}

	if (Move&PromFlag)>>12 > 3 {
		sourcePiece.typeID = uint8((Move&PromFlag)>>12) - 2
		board.pieceArrays[srcCol][sourcePiece.typeID] = append(board.pieceArrays[srcCol][sourcePiece.typeID], sourcePiece)
		board.pieceArrays[srcCol][1][len(board.pieceArrays[srcCol][1])-1].indexInPiecesArray = sourcePiece.indexInPiecesArray
		board.pieceArrays[srcCol][1][sourcePiece.indexInPiecesArray] = board.pieceArrays[srcCol][1][len(board.pieceArrays[srcCol][1])-1]
		board.pieceArrays[srcCol][1] = board.pieceArrays[srcCol][1][:len(board.pieceArrays[srcCol][1])-1]
		sourcePiece.indexInPiecesArray = uint8(len(board.pieceArrays[srcCol][sourcePiece.typeID]) - 1)
	}

	if (Move & FlagMask) == CastleFlag {

		board.pieceMasks[srcCol][targetPiece.typeID] ^= 1 << targetIndex
		board.pieceMask[srcCol] ^= 1 << targetIndex

		if targetIndex == castle_BRQS {
			board.pieceMasks[srcCol][4] |= 1 << castle_BRQT
			board.pieceMask[srcCol] |= 1 << castle_BRQT
			board.pieceMasks[srcCol][6] |= 1 << castle_BKQT
			board.pieceMask[srcCol] |= 1 << castle_BKQT

			board.squareArray[castle_BKS] = board.squareArray[castle_BKQT]
			board.squareArray[castle_BKQT] = sourcePiece
			board.squareArray[castle_BRQS] = board.squareArray[castle_BRQT]
			board.squareArray[castle_BRQT] = targetPiece

			sourcePiece.square = uint8(castle_BKQT)
			targetPiece.square = uint8(castle_BRQT)

			board.squareArray[castle_BKS].square = uint8(castle_BKS)
			board.squareArray[castle_BRQS].square = uint8(castle_BRQS)

		} else if targetIndex == castle_BRKS {
			board.pieceMasks[srcCol][4] |= 1 << castle_BRKT
			board.pieceMask[srcCol] |= 1 << castle_BRKT
			board.pieceMasks[srcCol][6] |= 1 << castle_BKKT
			board.pieceMask[srcCol] |= 1 << castle_BKKT

			board.squareArray[castle_BKS] = board.squareArray[castle_BKKT]
			board.squareArray[castle_BKKT] = sourcePiece
			board.squareArray[castle_BRKS] = board.squareArray[castle_BRKT]
			board.squareArray[castle_BRKT] = targetPiece

			sourcePiece.square = uint8(castle_BKKT)
			targetPiece.square = uint8(castle_BRKT)

			board.squareArray[castle_BKS].square = uint8(castle_BKS)
			board.squareArray[castle_BRKS].square = uint8(castle_BRKS)

		} else if targetIndex == castle_WRQS {
			board.pieceMasks[srcCol][4] |= 1 << castle_WRQT
			board.pieceMask[srcCol] |= 1 << castle_WRQT
			board.pieceMasks[srcCol][6] |= 1 << castle_WKQT
			board.pieceMask[srcCol] |= 1 << castle_WKQT

			board.squareArray[castle_WKS] = board.squareArray[castle_WKQT]
			board.squareArray[castle_WKQT] = sourcePiece
			board.squareArray[castle_WRQS] = board.squareArray[castle_WRQT]
			board.squareArray[castle_WRQT] = targetPiece

			sourcePiece.square = uint8(castle_WKQT)
			targetPiece.square = uint8(castle_WRQT)

			board.squareArray[castle_WKS].square = uint8(castle_WKS)
			board.squareArray[castle_WRQS].square = uint8(castle_WRQS)

		} else if targetIndex == castle_WRKS {
			board.pieceMasks[srcCol][4] |= 1 << castle_WRKT
			board.pieceMask[srcCol] |= 1 << castle_WRKT
			board.pieceMasks[srcCol][6] |= 1 << castle_WKKT
			board.pieceMask[srcCol] |= 1 << castle_WKKT

			board.squareArray[castle_WKS] = board.squareArray[castle_WKKT]
			board.squareArray[castle_WKKT] = sourcePiece
			board.squareArray[castle_WRKS] = board.squareArray[castle_WRKT]
			board.squareArray[castle_WRKT] = targetPiece

			sourcePiece.square = uint8(castle_WKKT)
			targetPiece.square = uint8(castle_WRKT)

			board.squareArray[castle_WKS].square = uint8(castle_WKS)
			board.squareArray[castle_WRKS].square = uint8(castle_WRKS)
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

		if sourceIndex == castle_BRQS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_BRQT
			board.pieceMask[srcCol] ^= 1 << castle_BRQT
			board.pieceMasks[srcCol][6] ^= 1 << castle_BKQT
			board.pieceMask[srcCol] ^= 1 << castle_BKQT

			board.pieceMasks[srcCol][4] |= 1 << castle_BRQS
			board.pieceMask[srcCol] |= 1 << castle_BRQS
			board.pieceMasks[srcCol][6] |= 1 << castle_BKS
			board.pieceMask[srcCol] |= 1 << castle_BKS

			board.squareArray[castle_BRQS] = board.squareArray[castle_BRQT]
			board.squareArray[castle_BRQT] = sourcePiece
			board.squareArray[castle_BKS] = board.squareArray[castle_BKQT]
			board.squareArray[castle_BKQT] = targetPiece

			board.squareArray[castle_BRQS].square = uint8(castle_BRQS)
			board.squareArray[castle_BKS].square = uint8(castle_BKS)

			board.squareArray[castle_BRQT].square = uint8(castle_BRQT)
			board.squareArray[castle_BKQT].square = uint8(castle_BKQT)

		} else if sourceIndex == castle_BRKS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_BRKT
			board.pieceMask[srcCol] ^= 1 << castle_BRKT
			board.pieceMasks[srcCol][6] ^= 1 << castle_BKKT
			board.pieceMask[srcCol] ^= 1 << castle_BKKT

			board.pieceMasks[srcCol][4] |= 1 << castle_BRKS
			board.pieceMask[srcCol] |= 1 << castle_BRKS
			board.pieceMasks[srcCol][6] |= 1 << castle_BKS
			board.pieceMask[srcCol] |= 1 << castle_BKS

			board.squareArray[castle_BRKS] = board.squareArray[castle_BRKT]
			board.squareArray[castle_BRKT] = sourcePiece
			board.squareArray[castle_BKS] = board.squareArray[castle_BKKT]
			board.squareArray[castle_BKKT] = targetPiece

			board.squareArray[castle_BRKS].square = uint8(castle_BRKS)
			board.squareArray[castle_BKS].square = uint8(castle_BKS)

			board.squareArray[castle_BRKT].square = uint8(castle_BRKT)
			board.squareArray[castle_BKKT].square = uint8(castle_BKKT)

		} else if sourceIndex == castle_WRQS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_WRQT
			board.pieceMask[srcCol] ^= 1 << castle_WRQT
			board.pieceMasks[srcCol][6] ^= 1 << castle_WKQT
			board.pieceMask[srcCol] ^= 1 << castle_WKQT

			board.pieceMasks[srcCol][4] |= 1 << castle_WRQS
			board.pieceMask[srcCol] |= 1 << castle_WRQS
			board.pieceMasks[srcCol][6] |= 1 << castle_WKS
			board.pieceMask[srcCol] |= 1 << castle_WKS

			board.squareArray[castle_WRQS] = board.squareArray[castle_WRQT]
			board.squareArray[castle_WRQT] = sourcePiece
			board.squareArray[castle_WKS] = board.squareArray[castle_WKQT]
			board.squareArray[castle_WKQT] = targetPiece

			board.squareArray[castle_WRQS].square = uint8(castle_WRQS)
			board.squareArray[castle_WKS].square = uint8(castle_WKS)

			board.squareArray[castle_WRQT].square = uint8(castle_WRQT)
			board.squareArray[castle_WKQT].square = uint8(castle_WKQT)

		} else if sourceIndex == castle_WRKS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_WRKT
			board.pieceMask[srcCol] ^= 1 << castle_WRKT
			board.pieceMasks[srcCol][6] ^= 1 << castle_WKKT
			board.pieceMask[srcCol] ^= 1 << castle_WKKT

			board.pieceMasks[srcCol][4] |= 1 << castle_WRKS
			board.pieceMask[srcCol] |= 1 << castle_WRKS
			board.pieceMasks[srcCol][6] |= 1 << castle_WKS
			board.pieceMask[srcCol] |= 1 << castle_WKS

			board.squareArray[castle_WRKS] = board.squareArray[castle_WRKT]
			board.squareArray[castle_WRKT] = sourcePiece
			board.squareArray[castle_WKS] = board.squareArray[castle_WKKT]
			board.squareArray[castle_WKKT] = targetPiece

			board.squareArray[castle_WRKS].square = uint8(castle_WRKS)
			board.squareArray[castle_WKS].square = uint8(castle_WKS)

			board.squareArray[castle_WRKT].square = uint8(castle_WRKT)
			board.squareArray[castle_WKKT].square = uint8(castle_WKKT)
		}

	} else {
		board.pieceMasks[srcCol][sourcePiece.typeID] |= 1 << targetIndex
		board.pieceMask[srcCol] |= 1 << targetIndex

		board.squareArray[targetIndex] = sourcePiece
		board.squareArray[sourceIndex] = targetPiece

		sourcePiece.square = uint8(targetIndex)
		targetPiece.square = uint8(sourceIndex)

	}
	// board.squareArray[targetIndex] = sourcePiece
	// board.squareArray[sourceIndex] = targetPiece
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
