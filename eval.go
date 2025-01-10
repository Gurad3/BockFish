package main

import "math/bits"

var pieceVals = []int{
	0, 100, 350, 410, 525, 900, 999999999,
}

func (board *board) EvalCurrentPosA() int {
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

					//ChainedPawn
					if colID == 0 {
						if ((mask_PawnAttack[square] >> 8) & colPieceMask) > 0 {
							ColorScore += 10
						}
					} else {
						if ((mask_PawnAttack[square] << 8) & colPieceMask) > 0 {
							ColorScore += 10
						}
					}

					//Isolated pawns
					if board.pieceMasks[colID][1]&mask_adjacentFiles[square] == 0 {
						ColorScore -= 20
					}

					// Doubled pawns
					if board.pieceMasks[colID][1]&(mask_sameFile[square]^1<<square) > 0 {
						ColorScore -= 20
					}

					// Passed pawns
					if board.pieceMasks[opCol][1]&mask_passedPawn[colID*64+square] == 0 {
						if colID == 0 {
							ColorScore += (7 - (square / 8)) * 15
						} else {
							ColorScore += (square / 8) * 15
						}
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

var eval_mg_value [7]int = [7]int{0, 82, 337, 365, 477, 1025, 0}
var eval_eg_value [7]int = [7]int{0, 94, 281, 297, 512, 936, 0}

var eval_gamephaseInc [14]int = [14]int{0, 0, 1, 1, 2, 4, 0}
var eval_mg_table [14][64]int
var eval_eg_table [14][64]int

func eval_init_tables() {

	for TypID := 1; TypID < 7; TypID++ {
		for sq := 0; sq < 64; sq++ {
			eval_mg_table[TypID*2][sq] = eval_mg_value[TypID] + mg_pesto_table[TypID][sq]
			eval_eg_table[TypID*2][sq] = eval_eg_value[TypID] + eg_pesto_table[TypID][sq]
			eval_mg_table[TypID*2+1][sq] = eval_mg_value[TypID] + mg_pesto_table[TypID][sq^56]
			eval_eg_table[TypID*2+1][sq] = eval_eg_value[TypID] + eg_pesto_table[TypID][sq^56]
		}
	}
}

func (board *board) EvalCurrentPos() int {
	mg := 0
	eg := 0
	gamePhase := 0
	score := 0

	for colID := 0; colID < 2; colID++ {
		ColorScore := 0
		colPieceMask := board.pieceMask[colID]
		opCol := 1 - colID
		for typeID := 1; typeID < 7; typeID++ {
			mask := board.pieceMasks[colID][typeID]
			for mask != 0 {
				square := popLSB(mask)
				mask &= mask - 1
				if colID == 0 {
					mg += eval_mg_table[typeID*2+colID][square]
					eg += eval_eg_table[typeID*2+colID][square]
				} else {
					mg -= eval_mg_table[typeID*2+colID][square]
					eg -= eval_eg_table[typeID*2+colID][square]
				}
				gamePhase += eval_gamephaseInc[typeID]

				if typeID == 1 { // Pawn

					//ChainedPawn
					if colID == 0 {
						if ((mask_PawnAttack[square] >> 8) & colPieceMask) > 0 {
							ColorScore += 10
						}
					} else {
						if ((mask_PawnAttack[square] << 8) & colPieceMask) > 0 {
							ColorScore += 10
						}
					}

					//Isolated pawns
					if board.pieceMasks[colID][1]&mask_adjacentFiles[square] == 0 {
						ColorScore -= 20
					}

					// Doubled pawns
					if board.pieceMasks[colID][1]&(mask_sameFile[square]^1<<square) > 0 {
						ColorScore -= 20
					}

					// Passed pawns
					if board.pieceMasks[opCol][1]&mask_passedPawn[colID*64+square] == 0 {
						if colID == 0 {
							ColorScore += (7 - (square / 8)) * 15
						} else {
							ColorScore += (square / 8) * 15
						}
					}

				} else if typeID == 6 { // King
					if board.isCheck(colID == 0) {
						ColorScore -= 50
					}

					ColorScore += 10 * bits.OnesCount64(mask_KingAttack[square]&board.pieceMasks[colID][1])

				}

			}

		}

		if colID == 0 {
			score += ColorScore
		} else {
			score -= ColorScore
		}
	}

	/* tapered eval */

	if gamePhase > 24 {
		gamePhase = 24
	}

	egPhase := 24 - gamePhase

	return (mg*gamePhase + eg*egPhase + score) / 24
}
