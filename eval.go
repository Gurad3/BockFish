package main

import "math/bits"

var pieceVals = []int{
	0, 100, 350, 410, 525, 900, 999999999,
}

func (board *board) EvalCurrentPos() int {
	score := 0

	for colID := 0; colID < 2; colID++ {

		opCol := 1 - colID
		colPieceMask := board.pieceMask[colID]
		opPieceMask := board.pieceMask[opCol]
		fullPieceMask := colPieceMask | opPieceMask

		ColorScore := 0

		for pieceType := 1; pieceType < 7; pieceType++ {

			pieceVal := pieceVals[pieceType]
			mask := board.pieceMasks[colID][pieceType]

			for mask > 0 {

				square := popLSB(mask)
				mask &= mask - 1
				ColorScore += pieceVal

				if pieceType == 1 { // Pawn
					ColorScore += eval_pawnTable[colID][square]

					//Isolated pawns
					if board.pieceMasks[colID][1]&mask_adjacentFiles[square] == 0 {
						ColorScore -= 20
					}

					// Doubled pawns
					if board.pieceMasks[colID][1]&(1<<square) > 0 {
						//ColorScore -= 10
					}

					// Passed pawns
					if board.pieceMasks[opCol][1]&mask_passedPawn[colID][square] == 0 {
						ColorScore += 30
					}

				} else if pieceType == 2 { // Knight
					ColorScore += eval_knightTable[colID][square]

				} else if pieceType == 3 { // Bishop
					ColorScore += eval_bishopTable[colID][square]
					ColorScore += 2 * bits.OnesCount64((BishMagicDict[square][((mask_BishAttack[square]&(fullPieceMask))*bishopMagics[square])>>bishopShifts[square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 4 { // Rook
					ColorScore += eval_rookTable[colID][square]
					ColorScore += 2 * bits.OnesCount64((RookMagicDict[square][((mask_RookAttack[square]&(fullPieceMask))*rookMagics[square])>>rookShifts[square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 5 { // Queen
					ColorScore += eval_queenTable[colID][square]
					ColorScore += 2 * bits.OnesCount64((RookMagicDict[square][((mask_RookAttack[square]&(fullPieceMask))*rookMagics[square])>>rookShifts[square]]|colPieceMask)^colPieceMask)
					ColorScore += 2 * bits.OnesCount64((BishMagicDict[square][((mask_BishAttack[square]&(fullPieceMask))*bishopMagics[square])>>bishopShifts[square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 6 { // King
					if board.isCheck(colID == 0) {
						ColorScore -= 50
					}

					ColorScore += 10 * bits.OnesCount64(mask_KingAttack[square]&board.pieceMasks[colID][1])

					opMaterial := bits.OnesCount64(board.pieceMasks[opCol][2])*2 + bits.OnesCount64(board.pieceMasks[opCol][3])*2
					opMaterial += bits.OnesCount64(board.pieceMasks[opCol][4])*3 + bits.OnesCount64(board.pieceMasks[opCol][5])*5

					if opMaterial <= 5 {
						ColorScore += eval_kingTable_end[colID][square]

					} else {
						ColorScore += eval_kingTable_middle[colID][square]
					}

				}
			}
		}

		if colID == 0 {
			score += ColorScore
		} else {
			score -= ColorScore
		}
	}

	return score
}
