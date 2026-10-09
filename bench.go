package main

import (
	"fmt"
	"time"
)

func searchFixedDepth(fen string, depth int) (uint64, uint64) {

	e := newBoard()
	e.loadFromFEN(fen)
	//fmt.Printf("%d \n", zobrist_black)
	//fmt.Printf("%d \n", zobrist_vals[0][1][0])
	e.searchDept(depth)
	return uint64(e.nodes), uint64(e.qnodes)

}

var benchFENs = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"r1bq1rk1/pp2ppbp/2np1np1/8/3NP3/2N1BP2/PPPQ2PP/R3KB1R w KQ - 0 9",
	"2rq1rk1/pp1bppbp/3p1np1/8/2PNP3/2N1BP2/PP1Q2PP/2KR1B1R w - - 0 12",
	"r2q1rk1/ppp2ppp/2n1bn2/3pp3/1b2P3/2NP1N2/PPP1BPPP/R1BQ1RK1 w - - 0 8",
	"6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
	"8/8/4k3/8/2p5/2K5/8/8 w - - 0 1",
	"4r1k1/1q3ppp/p1p5/3pP3/1P1P4/P3Q1P1/5P1P/4R1K1 w - - 0 1",
}

func RunBench(depth int) {
	loadT()
	if depth <= 0 {
		depth = 12
	}
	var total uint64
	var totalq uint64
	start := time.Now()
	for i, fen := range benchFENs {
		n, q := searchFixedDepth(fen, depth)
		total += n
		totalq += q
		fmt.Printf("position %2d: %d nodes and %d qNodes\n", i+1, n, q)

	}
	elapsed := time.Since(start)
	nps := uint64(float64(total) / elapsed.Seconds())
	qnps := uint64(float64(totalq) / elapsed.Seconds())
	fmt.Printf("\n%d nodes %d nps\n", total, nps)
	fmt.Printf("\n%d Qnodes %d qnps\n", totalq, qnps)
}
