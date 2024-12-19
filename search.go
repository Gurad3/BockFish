package main

import (
	"math"
	"strconv"
	"time"
)

type tranEnrty struct {
	depth, bound, score int
	key                 uint64
	Move                uint16
}

var posEvaled = 0

func (board *board) searchMove(searchTime int64) uint16 {
	posEvaled = 0
	startTime := time.Now()

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves)]
	// moves := board.getLegalMoves()
	currMove := moves[0]
	//White = true
	col := -1
	if board.isWhiteTurn {
		col = 1
	}
	currScore := (math.MinInt + 1) //* col

	depth := 2 - 1
	for time.Since(startTime).Milliseconds() <= searchTime {
		depth++

		for _, Move := range moves {

			if time.Since(startTime).Milliseconds() > searchTime {
				// writeToLogFile("posEvaled: " + strconv.Itoa(posEvaled))
				// writeToLogFile("currentDepth: " + strconv.Itoa(depth))

				return currMove
			}

			board.PlayMove(Move)
			score := -board.NegaMax(startTime, searchTime, depth, math.MinInt+1, math.MaxInt, -col, col)
			board.UndoMove(Move)
			if score > currScore {
				currScore = score
				currMove = Move
			}
		}
	}
	// writeToLogFile(strconv.Itoa(int(currScore)) + " Col:" + strconv.Itoa(col))
	writeToLogFile("posEvaled: " + strconv.Itoa(posEvaled))
	// board.tmp()

	return currMove
}

func (board *board) NegaMax(startTime time.Time, searchTime int64, depth int, alpha int, beta int, col int, root int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999 + depth*2
	}

	if depth == 0 {
		//return board.EvalCurrentPos() * col
		return board.QuinNegaMax(startTime, searchTime, depth, alpha, beta, col, root)
	}

	origAlpha := alpha
	entry := board.transpostionTable[board.zobrist%uint64(board.transEntries)]
	if entry.key == board.zobrist && entry.depth >= depth {

		score := entry.score
		// Exact
		if entry.bound == 1 {
			posEvaled++
			return score
		}

		// Lowerbound
		if entry.bound == 3 {
			alpha = intMax(alpha, score)
		} else { // Upperbound
			beta = intMin(beta, score)
		}

		if alpha >= beta {
			posEvaled++
			return score
		}
	}

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves)]
	// moves := board.getLegalMoves()

	var bestMove uint16
	//MoveOrdering

	score := math.MinInt + 1
	for _, Move := range moves {

		if time.Since(startTime).Milliseconds() > searchTime {
			return 99999999 + depth*2
		}

		board.PlayMove(Move)
		NewScore := -board.NegaMax(startTime, searchTime, depth-1, -beta, -alpha, -col, root)
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

func (board *board) QuinNegaMax(startTime time.Time, searchTime int64, depth int, alpha int, beta int, col int, root int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999 + depth*2
	}

	origAlpha := alpha
	entry := board.transpostionTable[board.zobrist%uint64(board.transEntries)]
	if entry.key == board.zobrist && entry.depth >= depth {

		score := entry.score
		// Exact
		if entry.bound == 1 {
			posEvaled++
			return score
		}

		// Lowerbound
		if entry.bound == 3 {
			alpha = intMax(alpha, score)
		} else { // Upperbound
			beta = intMin(beta, score)
		}

		if alpha >= beta {
			posEvaled++
			return score
		}
	}

	score := board.EvalCurrentPos() * col
	posEvaled++

	if score >= beta {
		return beta
	}
	alpha = intMax(alpha, score)

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves)]
	// moves := board.getLegalMoves()

	var bestMove uint16
	for _, Move := range moves {

		if time.Since(startTime).Milliseconds() > searchTime {
			return 99999999 + depth*2
		}

		if Move&CapFlag == 0 {
			continue
		}

		board.PlayMove(Move)
		NewScore := -board.QuinNegaMax(startTime, searchTime, depth-1, -beta, -alpha, -col, root)
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
