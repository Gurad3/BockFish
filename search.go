package main

import (
	"math"
	"time"
)

type tranEntry struct {
	depth, bound, score int
	key                 uint64
	Move                uint16
}

var global_currScore = 0

const (
	mateScore  = 99999999
	lmrEnabled = true
)

// Bench
func (board *board) searchDept(targetDepth int) uint16 {
	return board.iterate(targetDepth, time.Now().Add(50*time.Minute))
}

// UCI
func (board *board) searchMove(searchTime int64) uint16 {
	endTime := time.Now().Add(time.Duration(searchTime) * time.Millisecond)
	return board.iterate(64, endTime)
}

func (board *board) iterate(maxDepth int, endTime time.Time) uint16 {
	//fmt.Printf("info string budget %v\n", time.Until(endTime))
	board.nodes, board.qnodes, board.maxQDepth = 0, 0, 0
	board.stop = false
	//start := time.Now()
	pseudo := make([]uint16, 0, 218)
	pseudo = pseudo[:board.getLegalMoves(pseudo, false)]
	moves := make([]uint16, 0, len(pseudo))
	for _, m := range pseudo {
		board.PlayMove(m)
		if !board.isCheck(!board.isWhiteTurn) { // mover's king not attacked
			moves = append(moves, m)
		}
		board.UndoMove(m)
	}
	if len(moves) == 0 {
		return 0
	}
	quickSortStart(moves)

	col := -1
	if board.isWhiteTurn {
		col = 1
	}

	best := moves[0]
	for depth := 1; depth <= maxDepth; depth++ {
		board.rootDepth = depth

		alpha, beta := math.MinInt+1, math.MaxInt
		iterBest, iterBestID := moves[0], 0

		for id, m := range moves {
			board.PlayMove(m)
			score := -board.NegaMax(endTime, m, 1, depth-1, -beta, -alpha, -col, 0)
			board.UndoMove(m)

			if board.stop {
				if id > 0 {
					best = iterBest
				}
				return best
			}
			if score > alpha {
				alpha = score
				iterBest, iterBestID = m, id
			}
		}

		best = iterBest
		global_currScore = alpha
		moves[iterBestID], moves[0] = moves[0], moves[iterBestID]
		if alpha >= 99999 {
			writeToLogFile("Found Forced mate")
			return best
		}
		// ms := time.Since(start).Milliseconds()
		// nps := 0
		// if ms > 0 {
		// 	nps = board.nodes * 1000 / int(ms)
		// }
		//fmt.Printf("info depth %d score cp %d nodes %d nps %d time %d pv %s\n",depth, alpha, board.nodes, nps, ms, engineMoveToUCIMove(best))

	}
	return best
}

func (board *board) NegaMax(endTime time.Time, prevMove uint16, ply, depth, alpha, beta, col, numExtensions int) int {

	if depth <= 0 {
		//return board.EvalCurrentPos() * col
		return board.QuinNegaMax(endTime, 0, alpha, beta, col)
	}
	board.nodes++

	if board.nodes&2047 == 0 && board.rootDepth > 1 && time.Now().After(endTime) {
		board.stop = true
	}
	if board.stop {
		return 0
	}

	if board.isCheck(!board.isWhiteTurn) {
		return mateScore + depth*10
	}

	if board.isRepeated() {
		return 0
	}

	origAlpha, origBeta := alpha, beta
	idx := board.zobrist & (board.transEntries - 1)
	entry := board.transpostionTable[idx]
	tableMove := uint16(0)

	if entry.key == board.zobrist {
		tableMove = entry.Move

		if entry.depth >= depth {
			switch entry.bound {
			case 1: // exact
				return entry.score
			case 3: // lower bound
				alpha = intMax(alpha, entry.score)
			default: // upper bound
				beta = intMin(beta, entry.score)
			}
			if alpha >= beta {
				return entry.score
			}
		}
	}
	isCheck := board.isCheck(board.isWhiteTurn)
	extension := 0
	if numExtensions <= 3 && (prevMove&isPromFlag > 0 || isCheck) {
		extension = 1
	}

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, false)]

	if tableMove != 0 {
		moveOrderWithBestMove(moves, tableMove)

	} else {
		quickSortStart(moves)
	}
	bestScore := math.MinInt + 1
	bestMove := uint16(0)
	legal := 0
	newDepth := depth - 1 + extension
	newExt := numExtensions + extension

	for _, Move := range moves {

		board.PlayMove(Move)
		if board.isCheck(!board.isWhiteTurn) {
			board.UndoMove(Move)
			continue
		}
		legal++
		var score int
		score = -board.NegaMax(endTime, Move, ply+1, newDepth, -beta, -alpha, -col, newExt)
		board.UndoMove(Move)

		if board.stop {
			return 0
		}

		if score > bestScore {
			bestMove = Move
			bestScore = score
		}
		alpha = intMax(alpha, score)
		if alpha >= beta {
			break
		}

	}

	if legal == 0 {
		if isCheck {
			return -(mateScore + depth*10)
		}
		return 0
	}

	bound := 1 //exakt
	if bestScore <= origAlpha {
		bound = 2 //upper
	} else if bestScore >= origBeta {
		bound = 3 //lower
	}

	board.transpostionTable[idx] = tranEntry{
		bound: bound, depth: depth, score: bestScore, Move: bestMove, key: board.zobrist,
	}

	return bestScore
}

func (board *board) QuinNegaMax(endTime time.Time, qply, alpha, beta, col int) int {
	board.nodes++
	board.qnodes++

	if qply >= board.maxQDepth {
		board.maxQDepth = qply
	}

	if board.nodes&2047 == 0 && board.rootDepth > 1 && time.Now().After(endTime) {
		board.stop = true
	}
	if board.stop {
		return 0
	}
	if qply >= 24 {
		return board.EvalCurrentPos() * col
	}
	origAlpha, origBeta := alpha, beta
	idx := board.zobrist & (board.transEntries - 1)
	entry := board.quinTranspostionTable[idx]
	tableMove := uint16(0)

	if entry.key == board.zobrist {
		tableMove = entry.Move
		switch entry.bound {
		case 1:
			return entry.score
		case 3:
			alpha = intMax(alpha, entry.score)
		default:
			beta = intMin(beta, entry.score)
		}
		if alpha >= beta {
			return entry.score
		}
	}

	standPat := board.EvalCurrentPos() * col
	if standPat >= beta {
		return standPat
	}

	best := standPat
	if best > alpha {
		alpha = best
	}

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, true)]

	if tableMove != 0 {
		moveOrderWithBestMove(moves, tableMove)
	} else {
		quickSortStart(moves)
	}

	bestMove := uint16(0)
	for _, Move := range moves {
		board.PlayMove(Move)
		if board.isCheck(!board.isWhiteTurn) {
			board.UndoMove(Move)
			continue
		}
		// if Move&CapFlag == 0 && !board.isCheck(board.isWhiteTurn) {

		score := -board.QuinNegaMax(endTime, qply+1, -beta, -alpha, -col)
		board.UndoMove(Move)

		if board.stop {
			return 0
		}
		if score > best {
			best = score
			bestMove = Move
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}

	bound := 1
	if best <= origAlpha {
		bound = 2
	} else if best >= origBeta {
		bound = 3
	}
	board.quinTranspostionTable[idx] = tranEntry{
		bound: bound, depth: 0, score: best, Move: bestMove, key: board.zobrist,
	}

	return best
}

func (board *board) isRepeated() bool {
	n := len(board.repetion)
	sinceIrrev := board.ply - board.LastIrevPLY
	limit := n - sinceIrrev
	if limit < 0 {
		limit = 0
	}
	for i := n - 2; i >= limit; i -= 2 {
		if board.repetion[i] == board.zobrist {
			return true
		}
	}
	return false
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

func moveOrderWithBestMove(arr []uint16, bestMove uint16) {
	for i := 0; i < len(arr); i++ {
		if arr[i] == bestMove {
			arr[i] = 0b1111111111111111
			break
		}
	}
	quickSortStart(arr)
	if len(arr) != 0 && arr[0] == 0b1111111111111111 {
		arr[0] = bestMove
	}
}
