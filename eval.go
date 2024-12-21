package main

import "math/bits"

var pieceVals = []int{
	0, 100, 300, 350, 500, 900, 999999999,
}

func (board *board) EvalCurrentPosLegacy() int {
	score := 0

	for colID, Allpieces := range board.pieceArrays {

		for typID, pieceArray := range Allpieces {

			if colID == 0 {
				score += pieceVals[typID] * len(pieceArray)
			} else {
				score -= pieceVals[typID] * len(pieceArray)
			}

		}
	}
	return score
}

func (board *board) EvalCurrentPos() int {
	score := 0

	for colID := 0; colID < 2; colID++ {
		ColMult := -1 //If black, sub (-1)

		opCol := 0
		if colID == 0 {
			opCol = 1
			ColMult = 1 //If white, add (+1)
		}

		colPieceMask := board.pieceMask[colID]
		opPieceMask := board.pieceMask[opCol]
		fullPieceMask := colPieceMask | opPieceMask

		for typeID := 1; typeID < 7; typeID++ {

			if colID == 0 {
				score += pieceVals[typeID] * len(board.pieceArrays[colID][typeID])
			} else {
				score -= pieceVals[typeID] * len(board.pieceArrays[colID][typeID])
			}

			arr := board.pieceArrays[colID][typeID]
			for pieceID := 0; pieceID < len(arr); pieceID++ {
				piece := arr[pieceID]

				if typeID == 1 { //pawn
					score += eval_pawnTable[colID][piece.square] //Black is negative in table (no Mult)
				}
				if typeID == 2 { //knight
					score += eval_knightTable[colID][piece.square] //Black is negative in table (no Mult)
				}
				if typeID == 3 { //bishop
					score += eval_bishopTable[colID][piece.square] //Black is negative in table (no Mult)

					score += 2 * ColMult * bits.OnesCount64((BishMagicDict[piece.square][((BishAttackMask[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]|colPieceMask)^colPieceMask)

				}
				if typeID == 4 { //rook
					score += eval_rookTable[colID][piece.square] //Black is negative in table (no Mult)

					score += 2 * ColMult * bits.OnesCount64((RookMagicDict[piece.square][((RookAttackMask[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]|colPieceMask)^colPieceMask)

				}
				if typeID == 5 { //queen
					score += eval_queenTable[colID][piece.square] //Black is negative in table (no Mult)

					score += 2 * ColMult * bits.OnesCount64((RookMagicDict[piece.square][((RookAttackMask[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]|colPieceMask)^colPieceMask)
					score += 2 * ColMult * bits.OnesCount64((BishMagicDict[piece.square][((BishAttackMask[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]|colPieceMask)^colPieceMask)

				}
				if typeID == 6 { //king
					opMaterial := len(board.pieceArrays[opCol][2])*2 + len(board.pieceArrays[opCol][3])*2 + len(board.pieceArrays[opCol][4])*3 + len(board.pieceArrays[opCol][5])*4
					srcMaterial := len(board.pieceArrays[colID][2])*2 + len(board.pieceArrays[colID][3])*2 + len(board.pieceArrays[colID][4])*3 + len(board.pieceArrays[colID][5])*4

					if opMaterial <= 4 && srcMaterial <= 4 {
						score += eval_kingTable_end[colID][piece.square] //Black is negative in table (no Mult)
						// score += eval_kingTable_middle[colID][piece.square] //Black is negative in table (no Mult)

					} else {
						score += eval_kingTable_middle[colID][piece.square] //Black is negative in table (no Mult)
					}

				}

			}

		}

	}

	return score
}
