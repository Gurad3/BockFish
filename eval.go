package main

var pieceVals = []int{
	0, 100, 300, 350, 500, 900, 999999999,
}

func (board *board) EvalCurrentPos() int {

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
