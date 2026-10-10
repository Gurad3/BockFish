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
	mateScore    = 99999999
	lmrEnabled   = true
	deltaEnabled = true
	deltaMargin  = 200
	nullEnabled  = true
	nullMinDepth = 3
)

// Bench
func (board *board) searchDept(targetDepth int) uint16 {
	return board.iterate(targetDepth, time.Duration(50*time.Minute))
}

// UCI
func (board *board) searchMove(searchTime int64) uint16 {
	return board.iterate(64, time.Duration(searchTime)*time.Millisecond)
}

func (board *board) iterate(maxDepth int, duration time.Duration) uint16 {
	board.killers = [maxPly][2]uint16{}
	//fmt.Printf("info string budget %v\n", time.Until(endTime))
	endTime := time.Now().Add(duration)
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
	scores := make([]int32, len(moves))
	board.scoreMoves(moves, scores, 0, 0)
	for i := range moves {
		pickMove(moves, scores, i)
	}

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
				if depth <= 3 && iterBest < best-80 {
					depth--
					endTime = endTime.Add(duration)
					continue
				}
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
			//writeToLogFile("Found Forced mate")
			return best
		}
		// ms := time.Since(start).Milliseconds()
		// nps := 0
		// if ms > 0 {
		// 	nps = board.nodes * 1000 / int(ms)
		// }
		//fmt.Printf("info depth %d score cp %d nodes %d nps %d time %d pv %s\n", depth, alpha, board.nodes, nps, ms, engineMoveToUCIMove(best))

	}
	return best
}

func (board *board) NegaMax(endTime time.Time, prevMove uint16, ply, depth, alpha, beta, col, numExtensions int) int {
	if ply >= maxPly-1 {
		return board.EvalCurrentPos() * col
	}
	if depth <= 0 {
		return board.QuinNegaMax(endTime, ply, 0, alpha, beta, col)
	}

	board.nodes++
	if board.nodes&2047 == 0 && board.rootDepth > 1 && time.Now().After(endTime) {
		board.stop = true
	}
	if board.stop {
		return 0
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

	inCheck := board.isCheck(board.isWhiteTurn)

	s, cut := board.tryNullMove(endTime, prevMove, ply, depth, alpha, beta, col, numExtensions, inCheck)
	if cut {
		return s
	}

	extension := 0
	if numExtensions <= 3 && (inCheck || prevMove&isPromFlag > 0) {
		extension = 1
	}

	moves := board.moveBuf[ply][:0]
	moves = moves[:board.getLegalMoves(moves, false)]
	scores := board.scoreBuf[ply][:len(moves)]
	board.scoreMoves(moves, scores, tableMove, ply)

	bestScore := math.MinInt + 1
	bestMove := uint16(0)
	legal := 0
	newDepth := depth - 1 + extension
	newExt := numExtensions + extension

	var quiets [64]uint16
	nq := 0

	for i := range moves {
		Move := pickMove(moves, scores, i)

		board.PlayMove(Move)
		if board.isCheck(!board.isWhiteTurn) { // left own king attacked
			board.UndoMove(Move)
			continue
		}
		legal++
		isQuiet := Move&(CapFlag|isPromFlag) == 0

		var score int

		if lmrEnabled && legal > 4 && depth >= 3 && extension == 0 && isQuiet &&
			!board.isCheck(board.isWhiteTurn) {
			score = -board.NegaMax(endTime, Move, ply+1, newDepth-1, -alpha-1, -alpha, -col, newExt)
			if score > alpha && !board.stop {
				score = -board.NegaMax(endTime, Move, ply+1, newDepth, -beta, -alpha, -col, newExt)
			}
		} else {
			score = -board.NegaMax(endTime, Move, ply+1, newDepth, -beta, -alpha, -col, newExt)
		}

		board.UndoMove(Move)

		if board.stop {
			return 0
		}
		if score > bestScore {
			bestScore = score
			bestMove = Move
		}
		alpha = intMax(alpha, score)
		if alpha >= beta {
			if isQuiet {
				board.updateHistory(ply, depth, Move, quiets[:nq])
			}
			break
		}
		if isQuiet && nq < len(quiets) {
			quiets[nq] = Move
			nq++
		}
	}

	if legal == 0 {
		if inCheck {
			return -(mateScore + depth*10) // checkmated
		}
		return 0 // stalemate
	}

	bound := 1
	if bestScore <= origAlpha {
		bound = 2
	} else if bestScore >= origBeta {
		bound = 3
	}
	board.transpostionTable[idx] = tranEntry{
		bound: bound, depth: depth, score: bestScore, Move: bestMove, key: board.zobrist,
	}
	return bestScore
}

// ply indexes the move buffers, qply is only the depth below the main search.
func (board *board) QuinNegaMax(endTime time.Time, ply, qply, alpha, beta, col int) int {
	board.nodes++
	board.qnodes++
	if qply > board.maxQDepth {
		board.maxQDepth = qply
	}
	if board.nodes&2047 == 0 && board.rootDepth > 1 && time.Now().After(endTime) {
		board.stop = true
	}
	if board.stop {
		return 0
	}
	if qply >= 24 || ply >= maxPly-1 {
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

	moves := board.moveBuf[ply][:0]
	moves = moves[:board.getLegalMoves(moves, true)] // captures only
	scores := board.scoreBuf[ply][:len(moves)]
	board.scoreMoves(moves, scores, tableMove, ply)

	bestMove := uint16(0)
	for i := range moves {
		Move := pickMove(moves, scores, i)

		if deltaEnabled && Move&isPromFlag == 0 {
			victim := board.squareTypeArray[(Move&MoveTargetMask)>>6]
			if Move&FlagMask == EnPassant|CapFlag {
				victim = 1
			}
			if standPat+pieceValsMove[victim]+deltaMargin < alpha {
				continue // even winning the piece for free can't raise alpha
			}
		}

		board.PlayMove(Move)
		if board.isCheck(!board.isWhiteTurn) { // illegal capture
			board.UndoMove(Move)
			continue
		}
		score := -board.QuinNegaMax(endTime, ply+1, qply+1, -beta, -alpha, -col)
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
