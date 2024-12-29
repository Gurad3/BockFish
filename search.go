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
	endTime := startTime.Add(time.Duration(searchTime) * time.Millisecond)

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, false)]
	// moves := board.getLegalMoves()
	quickSortStart(moves)

	currMove := moves[0]

	//White = true
	col := -1
	if board.isWhiteTurn {
		col = 1
	}

	depth := 0 //2 - 1

	currScore := 0 //* col
	bestMoveID := 0
	for {
		depth++
		currScore = (math.MinInt + 1)
		for MoveID, Move := range moves {

			if depth > 1 && time.Now().After(endTime) {
				// writeToLogFile("posEvaled: " + strconv.Itoa(posEvaled))

				writeToLogFile("currentDepth: " + strconv.Itoa(depth))
				writeToLogFile("PoseEvaled: " + Itoa(posEvaled))
				writeToLogFile("Eval: " + Itoa(global_currScore) + "(Bot perspecitve)")

				return currMove
			}

			board.PlayMove(Move)
			score := -board.NegaMax(startTime, endTime, Move, 0, depth, math.MinInt+1, math.MaxInt, -col, 0)
			board.UndoMove(Move)
			if score > currScore {
				currScore = score
				currMove = Move
				global_currScore = score
				bestMoveID = MoveID
			}
		}

		if moves[0] != currMove {
			moves[bestMoveID], moves[0] = moves[0], moves[bestMoveID]
		}

		if currScore >= 99999 {
			writeToLogFile("Found Forced mate")

			return currMove
		}

	}
	// writeToLogFile(strconv.Itoa(int(currScore)) + " Col:" + strconv.Itoa(col))
	// board.tmp()

}

func (board *board) NegaMax(startTime time.Time, endTime time.Time, prevMove uint16, reverseDepth, depth, alpha, beta, col, numExtensions int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999 + depth*10
	}

	if board.isRepeated() {
		return 0
	}

	if depth <= 0 {
		//return board.EvalCurrentPos() * col
		return board.QuinNegaMax(startTime, endTime, reverseDepth, depth, alpha, beta, col)
	}

	origAlpha := alpha
	entry := board.transpostionTable[board.zobrist&(board.transEntries-1)]
	tableMove := entry.Move

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

	extension := 0
	if numExtensions <= 3 && (prevMove&isPromFlag > 0 || board.isCheck(board.isWhiteTurn)) {
		extension = 1
	}

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, false)]
	// moves := board.getLegalMoves()
	if len(moves) == 0 {
		return 0
	}

	if tableMove != 0 {
		moveOrderWithBestMove(moves, tableMove)

	} else {
		quickSortStart(moves)
	}
	bestMove := moves[0]

	//MoveOrdering

	depthReductionID := 7000
	if reverseDepth > 3 {
		if len(moves) > 6 {
			depthReductionID = len(moves) - len(moves)/4
		}
	}

	score := math.MinInt + 1
	for id, Move := range moves {

		if reverseDepth >= 1 && time.Now().After(endTime) {
			break
		}
		depthReduction := 0
		if id >= depthReductionID {
			depthReduction = 1
		}

		board.PlayMove(Move)
		NewScore := -board.NegaMax(startTime, endTime, Move, reverseDepth+1, depth-1+extension-depthReduction, -beta, -alpha, -col, numExtensions+extension)
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

	if time.Now().After(endTime) {
		return score
	}

	bound := 1 //exakt

	if score <= origAlpha {
		bound = 2 //upper
	} else if score >= beta {
		bound = 3 //lower
	}

	board.transpostionTable[board.zobrist&(board.transEntries-1)] = tranEntry{
		bound: bound, depth: depth, score: score, Move: bestMove, key: board.zobrist,
	}

	return score
}

func (board *board) QuinNegaMax(startTime time.Time, endTime time.Time, reverseDepth, depth, alpha, beta, col int) int {

	if board.isCheck(!board.isWhiteTurn) {
		return 99999999 + depth*10
	}

	origAlpha := alpha
	entry := board.quinTranspostionTable[board.zobrist&(board.transEntries-1)]
	tableMove := entry.Move

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

	posEvaled++
	score := board.EvalCurrentPos() * col

	if score >= beta {
		return beta
	}
	alpha = intMax(alpha, score)

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, true)]
	// moves := board.getLegalMoves()

	// quickSortStart(moves)

	if len(moves) == 0 {
		return score
	}

	if tableMove != 0 {
		for id, Move := range moves {
			if Move == tableMove {
				moves[id], moves[0] = moves[0], moves[id]
				break
			}
		}
	}
	bestMove := moves[0]

	for _, Move := range moves {

		if time.Now().After(endTime) {
			break
		}

		board.PlayMove(Move)

		// if Move&CapFlag == 0 && !board.isCheck(board.isWhiteTurn) {
		// 	board.UndoMove(Move)
		// 	continue
		// }

		NewScore := -board.QuinNegaMax(startTime, endTime, reverseDepth+1, depth-1, -beta, -alpha, -col)
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

	if time.Now().After(endTime) {
		return score
	}

	bound := 1 //exakt
	if score <= origAlpha {
		bound = 2 //upper
	} else if score >= beta {
		bound = 3 //lower
	}

	board.quinTranspostionTable[board.zobrist&(board.transEntries-1)] = tranEntry{
		bound: bound, depth: depth, score: score, Move: bestMove, key: board.zobrist,
	}

	return score
}

func (board *board) isRepeated() bool {

	// rep := board.repetion[board.ply-board.LastIrevPLY:]
	// for _, v := range rep {
	// 	if v == board.zobrist {
	// 		return true
	// 	}
	// }
	return false

	// i := len(board.repetion) - 1
	// for i >= board.ply-board.LastIrevPLY {
	// 	if board.repetion[i] == board.zobrist {
	// 		return true
	// 	}

	// 	i--
	// }
	// return false
}

// func (board *board) tmp() {

// 	for colID, Allpieces := range board.pieceArrays {

// 		for typID, pieceArray := range Allpieces {

// 			for _, piece := range pieceArray {

// 				d := "false"
// 				if piece.dead {
// 					d = "true"
// 				}
// 				writeToLogFile(strconv.Itoa(colID) + " " + strconv.Itoa(typID) + " " + d)
// 			}

// 		}
// 	}
// }

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

func moveOrderWithBestMove(arr []uint16, bestMove uint16) {
	for i := 0; i < len(arr); i++ {
		if arr[i] == bestMove {
			arr[i] = 0b1111111111111111
			break
		}
	}
	quickSortStart(arr)
	if arr[0] == 0b1111111111111111 {
		arr[0] = bestMove
	}
}
