package main

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

	for colID, Allpieces := range board.pieceArrays {

		for typeID, pieceArray := range Allpieces {

			if colID == 0 {
				score += pieceVals[typeID] * len(pieceArray)
			} else {
				score -= pieceVals[typeID] * len(pieceArray)
			}

			for _, piece := range pieceArray {

				if typeID == 1 { //pawn
					score += eval_pawnTable[colID][piece.square] //Black is negative in table
				}
				if typeID == 2 { //knight
					score += eval_knightTable[colID][piece.square]
				}
				if typeID == 3 { //bishop
					score += eval_bishopTable[colID][piece.square]
				}
				if typeID == 4 { //rook
					score += eval_rookTable[colID][piece.square]
				}
				if typeID == 5 { //queen
					score += eval_queenTable[colID][piece.square]
				}
				if typeID == 6 { //king
					score += eval_kingTable_middle[colID][piece.square]
				}

			}

		}

	}

	return score
}
