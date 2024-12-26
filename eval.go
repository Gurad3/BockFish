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
			pieceArray := board.pieceArrays[colID][pieceType]
			pieceCount := len(pieceArray)

			ColorScore += pieceVals[pieceType] * pieceCount

			for _, piece := range pieceArray {
				if pieceType == 1 { // Pawn
					ColorScore += eval_pawnTable[colID][piece.square]

					//Isolated pawns
					if board.pieceMasks[colID][1]&mask_adjacentFiles[piece.square] == 0 {
						ColorScore -= 20
					}

					// Doubled pawns
					if board.pieceMasks[colID][1]&(1<<piece.square) > 0 {
						//ColorScore -= 10
					}

					// Passed pawns
					if board.pieceMasks[opCol][1]&mask_passedPawn[colID][piece.square] == 0 {
						ColorScore += 30
					}

				} else if pieceType == 2 { // Knight
					ColorScore += eval_knightTable[colID][piece.square]

				} else if pieceType == 3 { // Bishop
					ColorScore += eval_bishopTable[colID][piece.square]
					ColorScore += 2 * bits.OnesCount64((BishMagicDict[piece.square][((mask_BishAttack[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 4 { // Rook
					ColorScore += eval_rookTable[colID][piece.square]
					ColorScore += 2 * bits.OnesCount64((RookMagicDict[piece.square][((mask_RookAttack[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 5 { // Queen
					ColorScore += eval_queenTable[colID][piece.square]
					ColorScore += 2 * bits.OnesCount64((RookMagicDict[piece.square][((mask_RookAttack[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]|colPieceMask)^colPieceMask)
					ColorScore += 2 * bits.OnesCount64((BishMagicDict[piece.square][((mask_BishAttack[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 6 { // King
					if board.isCheck(piece.isWhite) {
						ColorScore -= 50
					}

					ColorScore += 10 * bits.OnesCount64(mask_KingAttack[piece.square]&board.pieceMasks[colID][1])

					opMaterial := len(board.pieceArrays[opCol][2])*2 + len(board.pieceArrays[opCol][3])*2 + len(board.pieceArrays[opCol][4])*3 + len(board.pieceArrays[opCol][5])*5

					if opMaterial <= 5 {
						ColorScore += eval_kingTable_end[colID][piece.square]

					} else {
						ColorScore += eval_kingTable_middle[colID][piece.square]
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
