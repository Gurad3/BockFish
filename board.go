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

	squareTypeArray [64]uint8

	squareAttackNum [64]uint8

	deadPieces []uint8

	castling       uint8
	lastMove       uint16
	gameStateStack []gameState

	halfMoves int
	fullMoves int
	ply       int

	zobrist      uint64
	zobristVals  [2][7][64]uint64 // empty +
	zobristBlack uint64

	repetion     []uint64
	IrevPLYStack []int
	LastIrevPLY  int

	transEntries          uint64
	transpostionTable     []tranEntry
	quinTranspostionTable []tranEntry
}

type gameState struct {
	castling uint8
	move     uint16
}

const WKCastleMask uint64 = 0b1111000000000000000000000000000000000000000000000000000000000000
const WKCastleFlag uint64 = 0b1001000000000000000000000000000000000000000000000000000000000000
const WKCastleCheckMask uint64 = 0b111000000000000000000000000000000000000000000000000000000000000 //Cant castle thru check

const WQCastleMask uint64 = 0b1111100000000000000000000000000000000000000000000000000000000
const WQCastleFlag uint64 = 0b1000100000000000000000000000000000000000000000000000000000000
const WQCastleCheckMask uint64 = 0b1110000000000000000000000000000000000000000000000000000000000 //Cant castle thru check

const BKCastleMask uint64 = 0b11110000
const BKCastleFlag uint64 = 0b10010000
const BKCastleCheckMask uint64 = 0b1110000 //Cant castle thru check

const BQCastleMask uint64 = 0b11111
const BQCastleFlag uint64 = 0b10001
const BQCastleCheckMask uint64 = 0b11100 //Cant castle thr check

const MoveSourceMask uint16 = 0b0000000000111111
const MoveTargetMask uint16 = 0b0000111111000000
const FlagMask uint16 = 0b1111000000000000

const CapFlag uint16 = 0b1000000000000000

const isPromFlag uint16 = 0b0100000000000000
const PromMask uint16 = 0b0111000000000000

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

func (board *board) init() {
	board.initZobristArray()
	board.calcZobrist()
	board.IrevPLYStack = append(board.IrevPLYStack, board.ply)
	board.LastIrevPLY = board.ply
	board.repetion = append(board.repetion, board.zobrist)

	for colID := 0; colID < 2; colID++ {
		opCOl := 1 - colID
		//	opPieceMask := board.pieceMask[opCOl]
		colPieceMask := board.pieceMask[colID]
		fullPieceMask := board.pieceMask[opCOl] | board.pieceMask[colID]

		for TypeID := 1; TypeID < 7; TypeID++ {
			mask := board.pieceMasks[colID][TypeID]

			for mask != 0 {

				square := popLSB(mask)
				mask &= mask - 1

				moveMask := uint64(0)

				if TypeID == 1 { //Pawn
					if colID == 0 {
						moveMask |= (mask_PawnAttack[square] >> 8)
					} else {
						moveMask |= (mask_PawnAttack[square] << 8)
					}
				} else if TypeID == 2 { //Knight
					moveMask |= mask_KnightAttack[square] //Cretae movemask

				} else if TypeID == 6 { //King
					moveMask |= mask_KingAttack[square] //Cretae movemask
				} else if TypeID == 3 { //Bishop
					moveMask |= BishMagicDict[square][((mask_BishAttack[square]&(fullPieceMask))*bishopMagics[square])>>bishopShifts[square]]
				} else if TypeID == 4 { //Rook
					moveMask |= RookMagicDict[square][((mask_RookAttack[square]&(fullPieceMask))*rookMagics[square])>>rookShifts[square]]
				} else if TypeID == 5 { //Queen
					moveMask |= RookMagicDict[square][((mask_RookAttack[square]&(fullPieceMask))*rookMagics[square])>>rookShifts[square]]
					moveMask |= BishMagicDict[square][((mask_BishAttack[square]&(fullPieceMask))*bishopMagics[square])>>bishopShifts[square]]
				}

				moveMask |= colPieceMask
				moveMask ^= colPieceMask

				board.addAttackNum(colID == 0, moveMask)

			}
		}
	}

}

func (board *board) subAttackNum(col bool, mask uint64) {
	for mask != 0 {
		targetIndex := popLSB(mask)
		targetMask := uint64(1) << targetIndex
		mask ^= targetMask
		if col {
			org := board.squareAttackNum[targetIndex]
			new := ((((org & 0b11110000) >> 4) - 1) << 4) | (org & 0b00001111)
			board.squareAttackNum[targetIndex] = new
		} else {
			org := board.squareAttackNum[targetIndex]
			new := ((org & 0b0000111) - 1) | (org & 0b11110000)
			board.squareAttackNum[targetIndex] = new
		}

	}

}

func (board *board) addAttackNum(col bool, mask uint64) {
	for mask != 0 {
		targetIndex := popLSB(mask)
		targetMask := uint64(1) << targetIndex
		mask ^= targetMask
		if col {
			org := board.squareAttackNum[targetIndex]
			new := ((((org & 0b11110000) >> 4) + 1) << 4) | (org & 0b00001111)
			board.squareAttackNum[targetIndex] = new
		} else {
			org := board.squareAttackNum[targetIndex]
			new := ((org & 0b0000111) + 1) | (org & 0b11110000)
			board.squareAttackNum[targetIndex] = new
		}

	}

}

func (board *board) PlayMove(Move uint16) {
	//board.pieceMask |= 1 << piece.square
	sourceIndex := Move & MoveSourceMask
	targetIndex := (Move & MoveTargetMask) >> 6

	sourceType := board.squareTypeArray[sourceIndex]
	targetType := board.squareTypeArray[targetIndex]

	newGameState := gameState{}
	newGameState.castling = board.castling
	newGameState.move = board.lastMove

	board.repetion = append(board.repetion, board.zobrist)

	board.gameStateStack = append(board.gameStateStack, newGameState)

	srcCol := 1
	opCol := 0
	if board.isWhiteTurn {
		srcCol = 0
		opCol = 1

		if sourceType == 6 {
			board.castling &= 0b0011
		} else if sourceType == 4 {
			if sourceIndex == castle_WRQS {
				board.castling &= 0b0111
			} else if sourceIndex == castle_WRKS {
				board.castling &= 0b1011
			}
		}
		if targetType == 4 {
			if targetIndex == castle_BRQS {
				board.castling &= 0b1101
			} else if targetIndex == castle_BRKS {
				board.castling &= 0b1110
			}
		}
	} else {
		if sourceType == 6 {
			board.castling &= 0b1100
		} else if sourceType == 4 {
			if sourceIndex == castle_BRQS {
				board.castling &= 0b1101
			} else if sourceIndex == castle_BRKS {
				board.castling &= 0b1110
			}
		}
		if targetType == 4 {
			if targetIndex == castle_WRQS {
				board.castling &= 0b0111
			} else if targetIndex == castle_WRKS {
				board.castling &= 0b1011
			}
		}
	}

	Irev := sourceType == 1

	board.zobrist ^= board.zobristVals[srcCol][sourceType][sourceIndex]
	board.zobrist ^= board.zobristVals[srcCol][sourceType][targetIndex]
	board.zobrist ^= board.zobristBlack

	//Remove SourceBits.

	board.pieceMasks[srcCol][sourceType] ^= 1 << sourceIndex
	board.pieceMask[srcCol] ^= 1 << sourceIndex

	//fullPieceMask := board.pieceMask[0] | board.pieceMask[1]

	if Move&CapFlag > 0 {
		Irev = true
		deadIndex := targetIndex

		if Move&FlagMask == EnPassant|CapFlag {
			deadIndex = (board.lastMove & MoveTargetMask) >> 6 //Last Move == DoublePawn
		}
		deadType := board.squareTypeArray[deadIndex]

		board.pieceMasks[opCol][deadType] ^= 1 << deadIndex
		board.pieceMask[opCol] ^= 1 << deadIndex

		board.zobrist ^= board.zobristVals[opCol][deadType][deadIndex]

		board.deadPieces = append(board.deadPieces, deadType)

		// deadPiece.typeID = 0

		board.squareTypeArray[deadIndex] = 0

	}

	if (Move & isPromFlag) > 0 {
		Irev = true
		sourceType = uint8((Move&PromMask)>>12) - 2
	}

	if (Move & FlagMask) == CastleFlag {
		Irev = true
		board.pieceMasks[srcCol][targetType] ^= 1 << targetIndex
		board.pieceMask[srcCol] ^= 1 << targetIndex

		if targetIndex == castle_BRQS {
			board.pieceMasks[srcCol][4] |= 1 << castle_BRQT
			board.pieceMask[srcCol] |= 1 << castle_BRQT
			board.pieceMasks[srcCol][6] |= 1 << castle_BKQT
			board.pieceMask[srcCol] |= 1 << castle_BKQT

			board.squareTypeArray[castle_BKS] = 0
			board.squareTypeArray[castle_BKQT] = 6
			board.squareTypeArray[castle_BRQS] = 0
			board.squareTypeArray[castle_BRQT] = 4

		} else if targetIndex == castle_BRKS {
			board.pieceMasks[srcCol][4] |= 1 << castle_BRKT
			board.pieceMask[srcCol] |= 1 << castle_BRKT
			board.pieceMasks[srcCol][6] |= 1 << castle_BKKT
			board.pieceMask[srcCol] |= 1 << castle_BKKT

			board.squareTypeArray[castle_BKS] = 0
			board.squareTypeArray[castle_BKKT] = 6
			board.squareTypeArray[castle_BRKS] = 0
			board.squareTypeArray[castle_BRKT] = 4

		} else if targetIndex == castle_WRQS {
			board.pieceMasks[srcCol][4] |= 1 << castle_WRQT
			board.pieceMask[srcCol] |= 1 << castle_WRQT
			board.pieceMasks[srcCol][6] |= 1 << castle_WKQT
			board.pieceMask[srcCol] |= 1 << castle_WKQT

			board.squareTypeArray[castle_WKS] = 0
			board.squareTypeArray[castle_WKQT] = 6
			board.squareTypeArray[castle_WRQS] = 0
			board.squareTypeArray[castle_WRQT] = 4

		} else if targetIndex == castle_WRKS {
			board.pieceMasks[srcCol][4] |= 1 << castle_WRKT
			board.pieceMask[srcCol] |= 1 << castle_WRKT
			board.pieceMasks[srcCol][6] |= 1 << castle_WKKT
			board.pieceMask[srcCol] |= 1 << castle_WKKT

			board.squareTypeArray[castle_WKS] = 0
			board.squareTypeArray[castle_WKKT] = 6
			board.squareTypeArray[castle_WRKS] = 0
			board.squareTypeArray[castle_WRKT] = 4

		}

	} else {
		board.squareTypeArray[targetIndex] = sourceType
		board.squareTypeArray[sourceIndex] = 0

		//Update (add if empty) TargetBits
		board.pieceMasks[srcCol][sourceType] |= 1 << targetIndex
		board.pieceMask[srcCol] |= 1 << targetIndex

	}

	board.lastMove = Move
	board.isWhiteTurn = !board.isWhiteTurn
	board.ply++

	if Irev {
		// board.IrevPLYStack = append(board.IrevPLYStack, board.ply)
		// board.LastIrevPLY = board.ply
	}

}

func (board *board) UndoMove(Move uint16) {

	sourceIndex := (Move & MoveTargetMask) >> 6
	targetIndex := Move & MoveSourceMask

	sourceType := board.squareTypeArray[sourceIndex]

	srcCol := 1
	opCOL := 0
	if !board.isWhiteTurn {
		srcCol = 0
		opCOL = 1

	}
	board.castling = board.gameStateStack[len(board.gameStateStack)-1].castling
	board.lastMove = board.gameStateStack[len(board.gameStateStack)-1].move
	board.gameStateStack = board.gameStateStack[:len(board.gameStateStack)-1]

	board.zobrist = board.repetion[len(board.repetion)-1]
	board.repetion = board.repetion[:len(board.repetion)-1]

	// board.zobrist ^= board.zobristVals[srcCol][sourceType][sourceIndex]
	// board.zobrist ^= board.zobristVals[srcCol][sourceType][targetIndex]
	// board.zobrist ^= board.zobristBlack

	board.pieceMasks[srcCol][sourceType] ^= 1 << sourceIndex
	board.pieceMask[srcCol] ^= 1 << sourceIndex

	board.squareTypeArray[sourceIndex] = 0

	if Move&CapFlag > 0 {
		deadIndex := sourceIndex

		if Move&FlagMask == EnPassant|CapFlag {
			deadIndex = (board.lastMove & MoveTargetMask) >> 6 //Last Move == DoublePawn
		}

		deadType := board.deadPieces[len(board.deadPieces)-1]
		board.deadPieces = board.deadPieces[:len(board.deadPieces)-1]

		// board.zobrist ^= board.zobristVals[opCOL][deadType][deadIndex]

		board.pieceMask[opCOL] |= 1 << deadIndex
		board.pieceMasks[opCOL][deadType] |= 1 << deadIndex

		board.squareTypeArray[deadIndex] = deadType
	}

	if (Move & isPromFlag) > 0 {
		sourceType = uint8(1)
	}

	if (Move & FlagMask) == CastleFlag {
		//Undo
		board.pieceMasks[srcCol][sourceType] ^= 1 << sourceIndex
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

			board.squareTypeArray[castle_BRQS] = 4
			board.squareTypeArray[castle_BRQT] = 0
			board.squareTypeArray[castle_BKS] = 6
			board.squareTypeArray[castle_BKQT] = 0

		} else if sourceIndex == castle_BRKS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_BRKT
			board.pieceMask[srcCol] ^= 1 << castle_BRKT
			board.pieceMasks[srcCol][6] ^= 1 << castle_BKKT
			board.pieceMask[srcCol] ^= 1 << castle_BKKT

			board.pieceMasks[srcCol][4] |= 1 << castle_BRKS
			board.pieceMask[srcCol] |= 1 << castle_BRKS
			board.pieceMasks[srcCol][6] |= 1 << castle_BKS
			board.pieceMask[srcCol] |= 1 << castle_BKS

			board.squareTypeArray[castle_BRKS] = 4
			board.squareTypeArray[castle_BRKT] = 0
			board.squareTypeArray[castle_BKS] = 6
			board.squareTypeArray[castle_BKKT] = 0

		} else if sourceIndex == castle_WRQS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_WRQT
			board.pieceMask[srcCol] ^= 1 << castle_WRQT
			board.pieceMasks[srcCol][6] ^= 1 << castle_WKQT
			board.pieceMask[srcCol] ^= 1 << castle_WKQT

			board.pieceMasks[srcCol][4] |= 1 << castle_WRQS
			board.pieceMask[srcCol] |= 1 << castle_WRQS
			board.pieceMasks[srcCol][6] |= 1 << castle_WKS
			board.pieceMask[srcCol] |= 1 << castle_WKS

			board.squareTypeArray[castle_WRQS] = 4
			board.squareTypeArray[castle_WRQT] = 0
			board.squareTypeArray[castle_WKS] = 6
			board.squareTypeArray[castle_WKQT] = 0

		} else if sourceIndex == castle_WRKS {
			board.pieceMasks[srcCol][4] ^= 1 << castle_WRKT
			board.pieceMask[srcCol] ^= 1 << castle_WRKT
			board.pieceMasks[srcCol][6] ^= 1 << castle_WKKT
			board.pieceMask[srcCol] ^= 1 << castle_WKKT

			board.pieceMasks[srcCol][4] |= 1 << castle_WRKS
			board.pieceMask[srcCol] |= 1 << castle_WRKS
			board.pieceMasks[srcCol][6] |= 1 << castle_WKS
			board.pieceMask[srcCol] |= 1 << castle_WKS

			board.squareTypeArray[castle_WRKS] = 4
			board.squareTypeArray[castle_WRKT] = 0
			board.squareTypeArray[castle_WKS] = 6
			board.squareTypeArray[castle_WKKT] = 0
		}

	} else {
		board.pieceMasks[srcCol][sourceType] |= 1 << targetIndex
		board.pieceMask[srcCol] |= 1 << targetIndex

		board.squareTypeArray[targetIndex] = sourceType

	}

	// if board.ply == board.LastIrevPLY {
	// 	board.IrevPLYStack = board.IrevPLYStack[:len(board.IrevPLYStack)-1]
	// 	board.LastIrevPLY = board.IrevPLYStack[len(board.IrevPLYStack)-1]
	// }

	board.isWhiteTurn = !board.isWhiteTurn
	board.ply--

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

			mask := board.pieceMasks[col][i]

			for mask > 0 {
				sq := popLSB(mask)
				mask &= mask - 1
				board.zobrist ^= board.zobristVals[col][i][sq]
			}

		}
	}
}
