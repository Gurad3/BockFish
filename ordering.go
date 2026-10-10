package main

const maxPly = 128

var pieceValsMove = [7]int{0, 100, 320, 330, 500, 900, 0}

// Move ordering:
// TT move > captures (MVV-LVA) and queen promotions > killers > history.  Rest is sorted lazyly
func (board *board) scoreMoves(moves []uint16, scores []int32, ttMove uint16, ply int) {
	side := 0
	if !board.isWhiteTurn {
		side = 1
	}
	k0, k1 := board.killers[ply][0], board.killers[ply][1]

	for i, m := range moves {
		if m == ttMove && m != 0 {
			scores[i] = 1 << 30
			continue
		}

		from := m & MoveSourceMask
		to := (m & MoveTargetMask) >> 6
		var s int32

		switch {
		case m&CapFlag != 0:
			victim := board.squareTypeArray[to]
			if m&FlagMask == EnPassant|CapFlag {
				victim = 1 // enpassent  victim = pawn
			}
			attacker := board.squareTypeArray[from]
			// most valuable victim first, then least valuable attacker,  <- formel not made by me
			s = 100000 + int32(victim)*8 + int32(7-attacker)
		case m == k0:
			s = 90000
		case m == k1:
			s = 80000
		default:
			s = board.history[side][from][to]
		}

		if m&isPromFlag != 0 {
			if (m&PromMask)>>12 == PromoteToQueenFlag {
				s += 100060
			} else {
				s = -1000 // underpromotions last
			}
		}
		scores[i] = s
	}
}

// pickMove gör "lazy sort"
func pickMove(moves []uint16, scores []int32, i int) uint16 {
	best := i
	for j := i + 1; j < len(moves); j++ {
		if scores[j] > scores[best] {
			best = j
		}
	}
	moves[i], moves[best] = moves[best], moves[i]
	scores[i], scores[best] = scores[best], scores[i]
	return moves[i]
}

// Formel for hisotry move ordering, found online. basicly gives each sq to sq combination a score over the entire search, while fading values placed long ago.
func (board *board) addHistory(side int, m uint16, bonus int32) {
	h := &board.history[side][m&MoveSourceMask][(m&MoveTargetMask)>>6]
	ab := bonus
	if ab < 0 {
		ab = -ab
	}
	*h += bonus - *h*ab/16384 // keeps values bounded by +-16384
}

func (board *board) updateHistory(ply, depth int, cut uint16, tried []uint16) {
	if board.killers[ply][0] != cut {
		board.killers[ply][1] = board.killers[ply][0]
		board.killers[ply][0] = cut
	}
	side := 0
	if !board.isWhiteTurn {
		side = 1
	}
	bonus := int32(depth * depth)
	if bonus > 400 {
		bonus = 400
	}
	board.addHistory(side, cut, bonus)
	for _, q := range tried {
		board.addHistory(side, q, -bonus)
	}
}
