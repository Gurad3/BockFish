package main

import (
	"math"
	"strconv"
)

func (board *board) searchMove() move {
	depth := 5

	moves := board.getLegalMoves()
	currScore := math.MinInt
	currMove := moves[0]
	//White = true
	col := -1
	if board.isWhiteTurn {
		col = 1
	}
	for _, Move := range moves {
		board.PlayMove(Move)
		score := -board.NegaMax(depth, math.MinInt, math.MaxInt, -col)
		board.UndoMove(Move)
		if score > currScore {
			currScore = score
			currMove = Move
		}
	}
	writeToLogFile(strconv.Itoa(int(currScore)) + " Col:" + strconv.Itoa(col))
	return currMove
}

func (board *board) NegaMax(depth int, alpha int, beta int, col int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 9999999999999
	}

	if depth == 0 {
		//return board.EvalCurrentPos() * col
		return currentBoard.QuinNegaMax(depth, alpha, beta, col)
	}

	moves := board.getLegalMoves()
	//MoveOrdering

	score := math.MinInt
	for _, Move := range moves {
		board.PlayMove(Move)
		score = intMax(score, -board.NegaMax(depth-1, -beta, -alpha, -col))
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

func (board *board) QuinNegaMax(depth int, alpha int, beta int, col int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 9999999999999
	}

	//MoveOrdering

	score := board.EvalCurrentPos() * col

	if score >= beta {
		return beta
	}
	alpha = intMax(alpha, score)

	moves := board.getLegalMoves()

	for _, Move := range moves {

		if Move&CapFlag == 0 {
			continue
		}

		board.PlayMove(Move)
		score = intMax(score, -board.QuinNegaMax(depth-1, -beta, -alpha, -col))
		board.UndoMove(Move)
		alpha = intMax(alpha, score)
		if alpha >= beta {
			break
		}
	}
	return score
}
