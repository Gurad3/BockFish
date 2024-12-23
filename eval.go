package main

import "math/bits"

var pieceVals = []int{
	0, 100, 350, 410, 525, 900, 999999999,
}

func (board *board) EvalCurrentPos() int {
	score := 0

	for colID := 0; colID < 2; colID++ {
		colMultiplier := -1

		if colID == 0 {
			colMultiplier = 1
		}

		opCol := 1 - colID
		colPieceMask := board.pieceMask[colID]
		opPieceMask := board.pieceMask[opCol]
		fullPieceMask := colPieceMask | opPieceMask

		for pieceType := 1; pieceType < 7; pieceType++ {
			pieceArray := board.pieceArrays[colID][pieceType]
			pieceCount := len(pieceArray)

			score += pieceVals[pieceType] * pieceCount * colMultiplier

			for _, piece := range pieceArray {
				if pieceType == 1 { // Pawn
					score += eval_pawnTable[colID][piece.square] * colMultiplier

					//Isolated pawns
					if board.pieceMasks[colID][1]&mask_adjacentFiles[piece.square] == 0 {
						score -= 20 * colMultiplier
					}

					// Doubled pawns
					if board.pieceMasks[colID][1]&(1<<piece.square) > 0 {
						score -= 10 * colMultiplier
					}

					// Passed pawns
					if board.pieceMasks[opCol][1]&mask_passedPawn[colID][piece.square] == 0 {
						score += 40 * colMultiplier
					}

				} else if pieceType == 2 { // Knight
					score += eval_knightTable[colID][piece.square] * colMultiplier

				} else if pieceType == 3 { // Bishop
					score += eval_bishopTable[colID][piece.square] * colMultiplier
					score += 2 * colMultiplier * bits.OnesCount64((BishMagicDict[piece.square][((mask_BishAttack[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 4 { // Rook
					score += eval_rookTable[colID][piece.square] * colMultiplier
					score += 2 * colMultiplier * bits.OnesCount64((RookMagicDict[piece.square][((mask_RookAttack[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 5 { // Queen
					score += eval_queenTable[colID][piece.square] * colMultiplier
					score += 2 * colMultiplier * bits.OnesCount64((RookMagicDict[piece.square][((mask_RookAttack[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]|colPieceMask)^colPieceMask)
					score += 2 * colMultiplier * bits.OnesCount64((BishMagicDict[piece.square][((mask_BishAttack[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]|colPieceMask)^colPieceMask)

				} else if pieceType == 6 { // King
					if board.isCheck(piece.isWhite) {
						score -= 50 * colMultiplier
					}

					score += 10 * colMultiplier * bits.OnesCount64(mask_KingAttack[piece.square]&board.pieceMasks[colID][1])

					opMaterial := len(board.pieceArrays[opCol][2])*2 + len(board.pieceArrays[opCol][3])*2 + len(board.pieceArrays[opCol][4])*3 + len(board.pieceArrays[opCol][5])*4
					srcMaterial := len(board.pieceArrays[colID][2])*2 + len(board.pieceArrays[colID][3])*2 + len(board.pieceArrays[colID][4])*3 + len(board.pieceArrays[colID][5])*4

					if opMaterial <= 4 && srcMaterial <= 4 {
						score += eval_kingTable_end[colID][piece.square] * colMultiplier

					} else {
						score += eval_kingTable_middle[colID][piece.square] * colMultiplier
					}

				}
			}
		}

	}

	return score
}
