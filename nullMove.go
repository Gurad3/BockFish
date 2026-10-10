package main

import "time"

func (board *board) PlayNullMove() {
	board.gameStateStack = append(board.gameStateStack,
		gameState{castling: board.castling, move: board.lastMove})
	board.repetion = append(board.repetion, board.zobrist)

	board.zobrist ^= zobrist_black
	board.lastMove = 0 // no en passant, wich I read from lastMove
	board.isWhiteTurn = !board.isWhiteTurn
	board.ply++

	//repetion need it to be irrev
	board.IrevPLYStack = append(board.IrevPLYStack, board.ply)
	board.LastIrevPLY = board.ply
}

func (board *board) UndoNullMove() {
	st := board.gameStateStack[len(board.gameStateStack)-1]
	board.gameStateStack = board.gameStateStack[:len(board.gameStateStack)-1]
	board.castling = st.castling
	board.lastMove = st.move

	board.zobrist = board.repetion[len(board.repetion)-1]
	board.repetion = board.repetion[:len(board.repetion)-1]

	if board.ply == board.LastIrevPLY {
		board.IrevPLYStack = board.IrevPLYStack[:len(board.IrevPLYStack)-1]
		board.LastIrevPLY = board.IrevPLYStack[len(board.IrevPLYStack)-1]
	}
	board.isWhiteTurn = !board.isWhiteTurn
	board.ply--
}

// Zugzwang, connept that can arrive with null move,   if only king and pawns passing might be best move.
func (board *board) onlyKingOrPawn() bool {
	side := 0
	if !board.isWhiteTurn {
		side = 1
	}
	m := &board.pieceMasks[side]
	return m[2]|m[3]|m[4]|m[5] != 0
}

// returns (score, true) if the node can be cut off.

func (board *board) tryNullMove(endTime time.Time, prevMove uint16, ply, depth, alpha, beta, col, numExt int, inCheck bool) (int, bool) {
	// false instant
	if !nullEnabled || inCheck || prevMove == 0 || depth < nullMinDepth || !board.onlyKingOrPawn() {
		return 0, false
	}

	// if standing still is bad
	if board.EvalCurrentPos()*col < beta {
		return 0, false
	}

	r := 2
	if depth >= 6 {
		r = 3
	}

	board.PlayNullMove()
	score := -board.NegaMax(endTime, 0, ply+1, depth-1-r, -beta, -beta+1, -col, numExt)
	board.UndoNullMove()

	if board.stop {
		return 0, true
	}
	if score >= beta {
		if score > mateScore/2 { //have no guarentte mate. cuase this is not a real move on the line
			score = beta
		}
		return score, true
	}
	return 0, false
}
