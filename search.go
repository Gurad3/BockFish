package main

import (
	"math"
	"strconv"
	"time"
)

type tranEntry struct {
	depth, bound, score int
	key                 uint64
	Move                uint16
}

var posEvaled = 0
var global_currScore = 0

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
	global_currScore = currScore

	depth := 0 //2 - 1
	for {
		depth++

		for _, Move := range moves {

			if time.Since(startTime).Milliseconds() > searchTime {
				// writeToLogFile("posEvaled: " + strconv.Itoa(posEvaled))

				writeToLogFile("currentDepth: " + strconv.Itoa(depth))
				writeToLogFile("PoseEvaled: " + Itoa(posEvaled))
				writeToLogFile("Eval: " + Itoa(global_currScore) + "(Bot perspecitve)")

				return currMove
			}

			board.PlayMove(Move)
			score := -board.NegaMax(startTime, searchTime, depth, math.MinInt+1, math.MaxInt, -col, 0)
			board.UndoMove(Move)
			if score > currScore {
				currScore = score
				currMove = Move
				global_currScore = score

			}
		}

		if currScore >= 99999 {
			writeToLogFile("Found Forced mate")

			return currMove
		}
		if currScore <= -99999 {
			writeToLogFile("Found Forced mate (bad)")

		}
	}
	// writeToLogFile(strconv.Itoa(int(currScore)) + " Col:" + strconv.Itoa(col))
	// board.tmp()

	return currMove
}

func (board *board) NegaMax(startTime time.Time, searchTime int64, depth, alpha, beta, col, numExtensions int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999 + depth*10
	}

	if depth == 0 {
		return board.EvalCurrentPos() * col
		//return board.QuinNegaMax(startTime, searchTime, depth, alpha, beta, col)
	}

	extension := 0
	if numExtensions <= 3 && (board.isCheck(board.isWhiteTurn)) {
		extension = 1
	}

	origAlpha := alpha
	entry := board.transpostionTable[board.zobrist%board.transEntries]
	if entry.key == board.zobrist && entry.depth >= depth && false {

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
	// if len(moves) == 0 {
	// 	return -(99999999 + depth*10)
	// }

	// quickSortStart(moves)

	bestMove := moves[0]
	//MoveOrdering

	score := math.MinInt + 1
	for _, Move := range moves {

		if time.Since(startTime).Milliseconds() > searchTime {
			break
		}

		board.PlayMove(Move)
		NewScore := -board.NegaMax(startTime, searchTime, depth-1+extension, -beta, -alpha, -col, numExtensions+extension)
		board.UndoMove(Move)

		if NewScore > score {
			bestMove = Move
			score = NewScore
		}

		alpha = intMax(alpha, score)
		if alpha >= beta {
			break
		}

	}

	if time.Since(startTime).Milliseconds() > searchTime {
		return score
	}

	bound := 1 //exakt

	if score <= origAlpha {
		bound = 2 //upper
	} else if score >= beta {
		bound = 3 //lower
	}

	board.transpostionTable[board.zobrist%board.transEntries] = tranEntry{
		bound: bound, depth: depth, score: score, Move: bestMove, key: board.zobrist,
	}

	return score
}

func (board *board) QuinNegaMax(startTime time.Time, searchTime int64, depth, alpha, beta, col int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999 + depth*10
	}

	origAlpha := alpha
	entry := board.quinTranspostionTable[board.zobrist%board.transEntries]
	if entry.key == board.zobrist && entry.depth >= depth && false {

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

	posEvaled++
	score := board.EvalCurrentPos() * col

	if score >= beta {
		return beta
	}
	alpha = intMax(alpha, score)

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves)]
	// moves := board.getLegalMoves()

	// quickSortStart(moves)

	// if len(moves) == 0 {
	// 	return -(99999999 + depth*10)
	// }

	bestMove := moves[0]
	for _, Move := range moves {

		if time.Since(startTime).Milliseconds() > searchTime {
			break
		}

		// if Move&CapFlag == 0 {
		// 	continue
		// }

		board.PlayMove(Move)
		if Move&CapFlag == 0 && !board.isCheck(board.isWhiteTurn) {
			board.UndoMove(Move)
			continue
		}

		NewScore := -board.QuinNegaMax(startTime, searchTime, depth-1, -beta, -alpha, -col)
		board.UndoMove(Move)

		if NewScore > score {
			bestMove = Move
			score = NewScore
		}

		alpha = intMax(alpha, score)
		if alpha >= beta {
			break
		}

	}

	if time.Since(startTime).Milliseconds() > searchTime {
		return score
	}

	bound := 1 //exakt
	if score <= origAlpha {
		bound = 2 //upper
	} else if score >= beta {
		bound = 3 //lower
	}

	board.quinTranspostionTable[board.zobrist%board.transEntries] = tranEntry{
		bound: bound, depth: depth, score: score, Move: bestMove, key: board.zobrist,
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

func partition(arr []uint16, low, high int) int {
	pivot := arr[high] // Choose the last element as pivot.
	i := low - 1       // Index of smaller element.

	for j := low; j < high; j++ {
		if arr[j] > pivot { // Descending order: greater elements come first.
			i++
			arr[i], arr[j] = arr[j], arr[i] // Swap elements.
		}
	}

	i++
	arr[i], arr[high] = arr[high], arr[i] // Place the pivot in the correct position.
	return i
}

// quickSort sorts an array in-place in descending order using the Quick Sort algorithm.
func quickSort(arr []uint16, low, high int) {
	if low < high {
		// Partition the array and get the pivot index.
		pi := partition(arr, low, high)

		// Recursively sort the subarrays.
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func quickSortStart(arr []uint16) {
	quickSort(arr, 0, len(arr)-1)
}
