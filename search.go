package main

import "math"

func (board *board) searchMove() move {
	depth := 2

	moves := board.getLegalMoves(false)
	currScore := math.MinInt
	currMove := moves[0]

	for _, Move := range moves {
		board.PlayMove(Move)
		score := board.NegaMax(depth, math.MinInt, math.MaxInt, false)
		board.UndoMove(Move)
		if score > currScore {
			currScore = score
			currMove = Move
		}
	}

	return currMove
}

func (board *board) NegaMax(depth int, alpha int, beta int, col bool) int {

	if board.isCheck(!col) {
		return 9999999999999
	}

	if depth == 0 {
		return board.EvalCurrentPos(col)
	}

	moves := board.getLegalMoves(false)
	//MoveOrdering

	score := math.MinInt
	for _, Move := range moves {
		board.PlayMove(Move)
		score = intMax(score, -board.NegaMax(depth-1, -beta, -alpha, !col))
		board.UndoMove(Move)
		alpha = intMax(alpha, score)
		if alpha >= beta {
			break
		}
	}
	return score
}
func intMax(a int, b int) int {
	if a > b {
		return a
	}

	return b
}
