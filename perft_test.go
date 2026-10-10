package main

import (
	"fmt"
	"os"
	"sort"
	"testing"
	"time"
)

//Generated test script. to help find bugs. However most my bugs passed it :(

func legalMoves(b *board) []uint16 {
	pseudo := b.getLegalMovesA()
	legal := make([]uint16, 0, len(pseudo))
	for _, m := range pseudo {
		b.PlayMove(m)
		if !b.isCheck(!b.isWhiteTurn) {
			legal = append(legal, m)
		}
		b.UndoMove(m)
	}
	return legal
}

func newBoardFromFEN(fen string) *board {
	loadT()
	b := newBoard()
	b.loadFromFEN(fen)

	return &b
}

func perft(b *board, depth int) uint64 {
	if depth == 0 {
		return 1
	}
	moves := legalMoves(b)
	if depth == 1 {
		return uint64(len(moves)) // bulk counting
	}
	var nodes uint64
	for _, m := range moves {
		b.PlayMove(m)
		nodes += perft(b, depth-1)
		b.UndoMove(m)
	}
	return nodes
}

func perftDivide(b *board, depth int) map[string]uint64 {
	res := map[string]uint64{}
	for _, m := range legalMoves(b) {
		b.PlayMove(m)
		res[engineMoveToUCIMove(m)] = perft(b, depth-1)
		b.UndoMove(m)
	}
	return res
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type perftCase struct {
	name  string
	fen   string
	nodes []uint64 // nodes[i] = expected count at depth i+1
}

var perftCases = []perftCase{
	{"startpos", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		[]uint64{20, 400, 8902, 197281, 4865609, 119060324}},
	{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		[]uint64{48, 2039, 97862, 4085603, 193690690}},
	{"pos3", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		[]uint64{14, 191, 2812, 43238, 674624, 11030083}},
	{"pos4", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		[]uint64{6, 264, 9467, 422333, 15833292}},
	{"pos5", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		[]uint64{44, 1486, 62379, 2103487, 89941194}},
	{"pos6", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		[]uint64{46, 2079, 89890, 3894594, 164075551}},
}

func TestBench(t *testing.T) {
	RunBench(6)
}
func TestPerft(t *testing.T) {
	maxDepth := 6
	if testing.Short() {
		maxDepth = 4
	}
	for _, tc := range perftCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for d := 1; d <= len(tc.nodes) && d <= maxDepth; d++ {
				b := newBoardFromFEN(tc.fen)
				start := time.Now()
				got := perft(b, d)
				want := tc.nodes[d-1]
				if got != want {
					t.Errorf("depth %d: got %d, want %d (diff %+d)",
						d, got, want, int64(got)-int64(want))
					break // deeper depths will be wrong too
				}
				t.Logf("depth %d ok: %d nodes in %v", d, got, time.Since(start).Round(time.Millisecond))
			}
		})
	}
}

func TestDivide(t *testing.T) {
	fen := getenv("PERFT_FEN", "")
	if fen == "" {
		t.Skip("set PERFT_FEN and PERFT_DEPTH to use")
	}
	depth := 1
	fmt.Sscanf(getenv("PERFT_DEPTH", "1"), "%d", &depth)
	res := perftDivide(newBoardFromFEN(fen), depth)
	keys := make([]string, 0, len(res))
	var total uint64
	for k, v := range res {
		keys = append(keys, k)
		total += v
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %d\n", k, res[k])
	}
	fmt.Printf("\nTotal: %d\n", total)
}

func TestMakeUnmakeRestoresState(t *testing.T) {
	for _, tc := range perftCases {
		b := newBoardFromFEN(tc.fen)
		b.calcZobrist()
		before := b.zobrist
		for _, m := range legalMoves(b) {
			b.PlayMove(m)
			b.UndoMove(m)
			b.calcZobrist()
			if b.zobrist != before {
				t.Fatalf("%s: hash changed after make/unmake of %s", tc.name, engineMoveToUCIMove(m))
			}
		}
	}

}

func BenchmarkPerftKiwipete4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		perft(newBoardFromFEN(perftCases[1].fen), 4)
	}
}

func TestEPDirect(t *testing.T) {
	b := newBoardFromFEN("4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1")
	var ep uint16
	moves := make([]uint16, 0, 218)
	moves = moves[:b.getLegalMoves(moves, false)]
	for _, m := range moves {
		if m&MoveSourceMask == 28 && (m&MoveTargetMask)>>6 == 19 {
			ep = m
		}
	}
	if ep == 0 {
		t.Fatal("generator does not produce e5d6")
	}
	t.Logf("flags: %x (want %x)", ep&FlagMask, EnPassant|CapFlag)
	b.PlayMove(ep)
	if b.squareTypeArray[27] != 0 {
		t.Fatal("ghost pawn on d5 after PlayMove")
	}
	b.UndoMove(ep)
	if b.squareTypeArray[27] != 1 {
		t.Fatal("pawn not restored on d5 after UndoMove")
	}
}

func TestNullMoveRestores(t *testing.T) {
	fens := []string{
		"r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
		"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1",
		"rnbqkbnr/ppp1pppp/8/8/3pP3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 3", // en passant available
	}
	for _, tc := range perftCases {
		fens = append(fens, tc.fen)
	}

	for _, fen := range fens {
		b := newBoardFromFEN(fen)

		types, masks, mask, castling := b.squareTypeArray, b.pieceMasks, b.pieceMask, b.castling
		hash, last, ply, irev, white := b.zobrist, b.lastMove, b.ply, b.LastIrevPLY, b.isWhiteTurn
		nRep, nIrev, nState := len(b.repetion), len(b.IrevPLYStack), len(b.gameStateStack)

		b.PlayNullMove()

		if b.lastMove != 0 {
			t.Errorf("%s: lastMove not cleared (en passant would leak through the pass)", fen)
		}
		if b.isWhiteTurn == white {
			t.Errorf("%s: side to move not flipped", fen)
		}
		if b.squareTypeArray != types || b.pieceMasks != masks || b.pieceMask != mask || b.castling != castling {
			t.Errorf("%s: a pass changed the pieces or castling rights", fen)
		}
		// The hash after a pass must equal a from-scratch recompute.
		// (needs calcZobrist to include zobrist_black when black is to move)
		saved := b.zobrist
		b.calcZobrist()
		if b.zobrist != saved {
			t.Errorf("%s: hash after pass %016x != recomputed %016x", fen, saved, b.zobrist)
		}
		b.zobrist = saved

		b.UndoNullMove()

		if b.zobrist != hash || b.lastMove != last || b.isWhiteTurn != white {
			t.Errorf("%s: hash, lastMove or side not restored", fen)
		}
		if b.squareTypeArray != types || b.pieceMasks != masks || b.pieceMask != mask || b.castling != castling {
			t.Errorf("%s: pieces or castling not restored", fen)
		}
		if len(b.repetion) != nRep || len(b.IrevPLYStack) != nIrev || len(b.gameStateStack) != nState {
			t.Errorf("%s: a stack length changed", fen)
		}
		if b.ply != ply || b.LastIrevPLY != irev {
			t.Errorf("%s: ply bookkeeping not restored", fen)
		}
	}
}
