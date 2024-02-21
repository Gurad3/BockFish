package main

var KnightAttacks = [][]int{
	{1, 2}, {2, 1},
	{2, -1}, {1, -2},
	{-1, -2}, {-2, -1},
	{-2, 1}, {-1, 2},
}

var king = [][]int{
	{0, 1}, {1, 1},
	{1, 0}, {1, -1},
	{0, -1}, {-1, -1},
	{-1, 0}, {-1, 1},
}

func (board *board) getLegalMoves(onlyCaptures bool) []move {

	for colID, AllPieceArray := range board.pieceArrays {
		for TypeID, pieceArray := range AllPieceArray {
			for _, piece := range pieceArray {

			}
		}
	}

	return []move{}
}
