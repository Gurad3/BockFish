package main

var pieceVals = []int{
	0, 100, 300, 350, 500, 900, 9999999,
}

func (board *board) EvalCurrentPos(perSpect bool) int {

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

	if perSpect {
		return score
	} else {
		return -score
	}

}
