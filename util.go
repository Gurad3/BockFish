package main

import (
	"log"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

var startfen string = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

var typeToString string = "apnbrqk"
var stringToType = map[string]uint8{
	"p": 1,
	"n": 2,
	"b": 3,
	"r": 4,
	"q": 5,
	"k": 6,
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (board *board) loadFromFEN(fenS string) {

	board.pieceMask = [2]uint64{}
	board.pieceMasks = [2][7]uint64{}
	board.squareTypeArray = [64]uint8{}
	board.squareAttackNum = [64]uint8{}
	board.castling, board.lastMove, board.zobrist = 0, 0, 0
	board.gameStateStack = board.gameStateStack[:0]
	board.repetion = board.repetion[:0]
	board.IrevPLYStack = board.IrevPLYStack[:0]

	fen := strings.Split(fenS, " ")

	board.deadPieces = make([]uint8, 0, 30)

	board.isWhiteTurn = fen[1] == "w"

	for i := 0; i < len(fen[2]); i++ {
		if fen[2][i] == "K"[0] {
			board.castling |= 0b0100
		}
		if fen[2][i] == "Q"[0] {
			board.castling |= 0b1000
		}
		if fen[2][i] == "k"[0] {
			board.castling |= 0b0001
		}
		if fen[2][i] == "q"[0] {
			board.castling |= 0b0010
		}
	}

	if fen[3] != "-" {
		lastMove := uint16(0)
		X := strings.IndexRune(numMoveToUCIMove, rune(fen[3][0]))
		Y := 8 - int(fen[3][1]-'0')

		if Y > 4 { //WhitePawn
			lastMove = uint16((Y-1)*8+X) << 6
			lastMove += uint16((Y+1)*8 + X)
		} else { //BlackPawn
			lastMove = uint16((Y+1)*8+X) << 6
			lastMove += uint16((Y-1)*8 + X)
		}

		board.lastMove = lastMove | PawnTwoUpFlag
	}

	board.halfMoves, _ = strconv.Atoi(fen[4])
	board.fullMoves, _ = strconv.Atoi(fen[5])
	board.ply = board.fullMoves * 2
	if !board.isWhiteTurn {
		board.ply++
	}

	x := 0
	y := 0

	FenPos := fen[0]
	for i := range FenPos {
		st := string([]byte{FenPos[i]})
		if st == "/" {
			x = 0
			y++
			continue
		}
		if isNumber(st) {
			val, _ := strconv.Atoi(st)

			for i := 0; i < val; i++ {
				x++
			}

			continue
		}

		typeID := stringToType[strings.ToLower(st)]
		square := uint8(x + y*8)
		board.squareTypeArray[square] = typeID
		colID := 1
		if st == strings.ToUpper(st) {
			colID = 0
		}
		board.pieceMask[colID] |= 1 << square
		board.pieceMasks[colID][typeID] |= 1 << square

		x++
	}
	board.calcZobrist()
	board.init()
}

func StringMoveToMove(mo string) uint16 {
	newMove := uint16(0)
	splited := strings.Split(mo, "")
	SX, _ := strconv.Atoi(splited[0])
	SY, _ := strconv.Atoi(splited[1])

	TX, _ := strconv.Atoi(splited[2])
	TY, _ := strconv.Atoi(splited[3])

	newMove += uint16(TX) + uint16(TY)*8
	newMove <<= 6

	newMove += uint16(SX) + uint16(SY)*8
	return newMove
}

func moveArrayTOstring(array []uint16) string {
	retString := ""
	for _, Move := range array {

		SIndex := (MoveSourceMask & Move)
		TIndex := (MoveTargetMask & Move) >> 6

		moveString := ""

		SX := SIndex % 8
		SY := SIndex / 8
		TX := TIndex % 8
		TY := TIndex / 8

		moveString += strconv.Itoa(int(SX))
		moveString += strconv.Itoa(int(SY))
		moveString += strconv.Itoa(int(TX))
		moveString += strconv.Itoa(int(TY))

		retString += moveString + " "
		tmpStringToMove[moveString] = Move
		//fmt.Println(SIndex, TIndex, strconv.Itoa(int(SX)), moveString)
	}
	return retString
}
func sendPosAsString() string {
	retString := ""
	for i := 0; i < 64; i++ {
		partString := ""
		SX := i % 8
		SY := i / 8
		partString += strconv.Itoa(int(SX))
		partString += strconv.Itoa(int(SY))

		if currentBoard.pieceMask[1]&(1<<i) > 0 {
			partString += "b"
		} else {
			partString += "w"
		}

		partString += Itoa(int(currentBoard.squareTypeArray[i]))
		retString += partString + " "

	}

	return retString
}

var tmpStringToMove map[string]uint16 = make(map[string]uint16, 1000)

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king

*/

const numMoveToUCIMove string = "abcdefgh"

func engineMoveToUCIMove(Move uint16) string {
	SIndex := (MoveSourceMask & Move)
	TIndex := (MoveTargetMask & Move) >> 6

	SX := string(numMoveToUCIMove[(SIndex % 8)])
	SY := strconv.Itoa(8 - int(SIndex/8))
	TX := string(numMoveToUCIMove[(TIndex % 8)])
	TY := strconv.Itoa(8 - int(TIndex/8))
	// fmt.Println(Move, SIndex, TIndex, SX, SY, TX, TY)

	if (Move & isPromFlag) > 0 {
		TY += string(typeToString[((Move&PromMask)>>12)-2])
	}

	return SX + SY + TX + TY

}

func UCIMoveToEngineMove(uciMove string) uint16 {

	SX := strings.IndexRune(numMoveToUCIMove, rune(uciMove[0]))
	SY := 8 - int(uciMove[1]-'0')
	TX := strings.IndexRune(numMoveToUCIMove, rune(uciMove[2]))
	TY := 8 - int(uciMove[3]-'0')

	SIndex := (SY)*8 + SX
	TIndex := (TY)*8 + TX
	var Move uint16 = uint16(SIndex | (TIndex << 6))
	// fmt.Println(Move, SIndex, TIndex, SX, SY, TX, TY)

	if len(uciMove) > 4 {
		Move |= uint16(stringToType[string(uciMove[4])]+2) << 12
	}

	if currentBoard.squareTypeArray[SIndex] == 6 {
		if SIndex-TIndex == 2 {
			Move = uint16(SIndex|((SIndex-4)<<6)) | CastleFlag
		}

		if SIndex-TIndex == -2 {
			Move = uint16(SIndex|((SIndex+3)<<6)) | CastleFlag
		}
	}

	if currentBoard.squareTypeArray[SIndex] == 1 {
		if SIndex-TIndex == 16 || SIndex-TIndex == -16 {
			Move |= PawnTwoUpFlag
		}

		if currentBoard.lastMove&FlagMask == PawnTwoUpFlag {
			if currentBoard.isWhiteTurn && TIndex-int(currentBoard.lastMove&MoveSourceMask) == 8 {
				Move |= EnPassant | CapFlag
			}

			if !currentBoard.isWhiteTurn && TIndex-int(currentBoard.lastMove&MoveSourceMask) == -8 {
				Move |= EnPassant | CapFlag
			}

		}
	}

	return Move
}

func (board *board) generateCapFlag(Move uint16) uint16 {
	OpColID := 0
	if board.isWhiteTurn {
		OpColID = 1
	}
	if (1<<((Move&MoveTargetMask)>>6))&board.pieceMask[OpColID] > 0 {
		Move |= CapFlag
	}

	return Move
}

func writeToLogFile(message string) {
	// Open the file in append mode. Create it if it doesn't exist.

	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a logger that writes to the file
	logger := log.New(file, "", log.LstdFlags)

	// Write the log message
	logger.Println(message)
}

func intMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func popLSB(n uint64) int {
	return bits.TrailingZeros64(n)
}

func Itoa(i int) string {
	return strconv.Itoa(i)
}

var ep = 0
var capt = 0
var prom = 0

func (board *board) MoveGenTest(depth int) int {
	if depth == 0 {
		return 1
	}

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, false)]

	pos := 0
	for _, mov := range moves {

		board.PlayMove(mov)

		if !board.isCheck(!board.isWhiteTurn) {
			if depth == 1 {
				if mov&CapFlag > 0 {
					if mov&FlagMask == EnPassant|CapFlag {
						ep++
						capt++
					} else {

						capt++

					}
				}
				if mov&isPromFlag > 0 {
					prom++
				}

			}

			pos += board.MoveGenTest(depth - 1)
		}

		board.UndoMove(mov)
	}

	return pos
}
