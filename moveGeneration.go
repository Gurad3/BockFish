package main

import (
	"math"
)

func (board *board) getLegalMoves(onlyCaptures bool) []move {

	legalMoves := make([]move, 0, 218)
	colID := 1
	if board.isWhiteTurn {
		colID = 0
	}

	for TypeID, pieceArray := range board.pieceArrays[colID] {
		for _, piece := range pieceArray {
			if piece.dead {
				continue
			}

			var moveMask uint64
			pawnUpTwo := false
			opCOl := 0
			if colID == 0 {
				opCOl = 1
			}

			if TypeID == 1 { //Pawn
				if piece.isWhite {
					moveMask = (PawnAttackMasks[piece.square] & board.pieceMask[opCOl])

					if (1<<(piece.square-8))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
						moveMask |= (1 << (piece.square - 8))

						if piece.square > 47 { //not moved, check if dual is possible
							if (1<<(piece.square-16))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
								moveMask |= (1 << (piece.square - 16))
								pawnUpTwo = true
							}
						}

					}

				} else {
					moveMask = ((PawnAttackMasks[piece.square] << 16) & board.pieceMask[opCOl])

					if (1<<(piece.square+8))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
						moveMask |= (1 << (piece.square + 8))

						if piece.square < 16 { //not moved, check if dual is possible
							if (1<<(piece.square+16))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
								moveMask |= (1 << (piece.square + 16))
								pawnUpTwo = true
							}
						}
					}
				}

			} else if TypeID == 2 { //Knight
				moveMask = (KnightAttackMasks[piece.square] & board.pieceMask[colID]) ^ KnightAttackMasks[piece.square] //Cretae movemask

			} else if TypeID == 6 { //King
				moveMask = (KingAttackMasks[piece.square] & board.pieceMask[colID]) ^ KingAttackMasks[piece.square] //Cretae movemask
			}

			for moveMask > 0 {
				targetIndex := popLSB(moveMask)

				flag := move(0)
				newMove := move(targetIndex) << 6
				newMove += move(piece.square)

				if 1<<targetIndex&board.pieceMask[opCOl] > 0 {
					flag = CapFlag
				}

				if TypeID == 1 {

					if pawnUpTwo {
						if targetIndex <= 39 && targetIndex >= 24 {
							flag = PawnTwoUpFlag
						}

					} else if targetIndex < 8 || targetIndex > 55 { //Pawn is promoting

						for i := move(4); i < 8; i++ {
							legalMoves = append(legalMoves, newMove|i<<12|flag)
						}
						moveMask ^= 1 << targetIndex
						continue
					}
				}

				legalMoves = append(legalMoves, newMove|flag)
				moveMask ^= 1 << targetIndex
			}

		}

	}

	return legalMoves
}

func (board *board) isCheck(col bool) bool {
	colID := 1
	opCol := 0

	if col {
		colID = 0
		opCol = 1
	}

	kingIndex := board.kings[colID].square
	tmpMask := (board.pieceMasks[opCol][2] & KnightAttackMasks[kingIndex]) | (board.pieceMasks[opCol][6] & KingAttackMasks[kingIndex]) //King and Knights
	if col {
		tmpMask |= board.pieceMasks[opCol][1] & PawnAttackMasks[kingIndex] //Pawns
	} else {
		tmpMask |= board.pieceMasks[opCol][1] & (PawnAttackMasks[kingIndex] << 16) //Pawns
	}

	return tmpMask > 0
}

func (board *board) filterOutPsudoLegal(moves []move) []move {
	newList := make([]move, 0, cap(moves))
	for _, mo := range moves {

		board.PlayMove(mo)
		if !board.isCheck(!board.isWhiteTurn) {
			newList = append(newList, mo)
		}
		board.UndoMove(mo)

	}
	return newList
}

func popLSB(n uint64) int {
	// Take two's complement of the given number
	complement := n & ^(n - 1)
	// Take the log2 of the complement
	pos := int(math.Log2(float64(complement)))
	return pos
}

const NoFlag move = 0b0000 << 12
const EnPassant move = 0b0001 << 12
const CastleFlag move = 0b0010 << 12
const PawnTwoUpFlag move = 0b0011 << 12

const PromoteToKnightFlag move = 0b0100
const PromoteToBishopFlag move = 0b0101
const PromoteToRookFlag move = 0b0110
const PromoteToQueenFlag move = 0b0111

//var PromotFlags [4]move = [4]move{PromoteToKnightFlag, PromoteToBishopFlag, PromoteToRookFlag, PromoteToQueenFlag}

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king


*/
