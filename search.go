package main

import (
	"math"
	"strconv"
)

func (board *board) searchMove() move {
	depth := 3

	moves := board.getLegalMoves()
	currMove := moves[0]
	//White = true
	col := -1
	if board.isWhiteTurn {
		col = 1
	}
	currScore := (math.MinInt + 1) //* col

	for _, Move := range moves {
		board.PlayMove(Move)
		score := -board.NegaMax(depth, math.MinInt+1, math.MaxInt, -col, col)
		board.UndoMove(Move)
		if score > currScore {
			currScore = score
			currMove = Move
		}
	}

	// writeToLogFile(strconv.Itoa(int(currScore)) + " Col:" + strconv.Itoa(col))

	// board.tmp()

	return currMove
}

func (board *board) NegaMax(depth int, alpha int, beta int, col int, root int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999
	}
	origAlpha := alpha
	entry := board.transpostionTable[board.zobrist%uint64(board.transEntries)]

	if entry.key == board.zobrist && entry.depth >= depth {

		score := entry.score
		// Exact
		if entry.bound == 1 {
			// posEvaled++
			return score
		}

		// Lowerbound
		if entry.bound == 3 {
			alpha = intMax(alpha, score)
		} else { // Upperbound
			beta = intMin(beta, score)
		}

		if alpha >= beta {
			// posEvaled++
			return score
		}
	}

	if depth == 0 {
		//return board.EvalCurrentPos() * col
		return board.QuinNegaMax(depth, alpha, beta, col, root)
	}

	moves := board.getLegalMoves()
	var bestMove move
	//MoveOrdering

	score := math.MinInt + 1
	for _, Move := range moves {
		board.PlayMove(Move)
		NewScore := -board.NegaMax(depth-1, -beta, -alpha, -col, root)
		board.UndoMove(Move)

		if NewScore > score {
			bestMove = Move
			score = NewScore
			alpha = intMax(alpha, score)
			if alpha >= beta {
				break
			}
		}

	}

	bound := 0
	if score >= beta {
		bound = 3
	} else {
		if score <= origAlpha {
			bound = 2
		} else {
			bound = 1
		}
	}
	ttEntry := tranEnrty{bound: bound, depth: depth, score: score, Move: bestMove, key: board.zobrist}
	board.transpostionTable[board.zobrist%uint64(board.transEntries)] = ttEntry

	return score
}

func (board *board) QuinNegaMax(depth int, alpha int, beta int, col int, root int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999
	}

	//MoveOrdering
	if depth <= -3 {
		return -900
	}

	// score := 0
	score := board.EvalCurrentPos() * col
	// if root == 1 {
	// 	score = board.EvalCurrentPos() * col
	// } else {
	// 	score = board.EvalCurrentPosLegacy() * col
	// }

	// writeToLogFile(strconv.Itoa(int(col)))

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
		score = intMax(score, -board.QuinNegaMax(depth-1, -beta, -alpha, -col, root))
		board.UndoMove(Move)
		alpha = intMax(alpha, score)
		if alpha >= beta {
			break
		}
	}
	return score
}

func (board *board) tmp() {

	for colID, Allpieces := range board.pieceArrays {

		for typID, pieceArray := range Allpieces {

			for _, piece := range pieceArray {

				d := "false"
				if piece.dead {
					d = "true"
				}
				writeToLogFile(strconv.Itoa(colID) + " " + strconv.Itoa(typID) + " " + d)
			}

		}
	}
}
