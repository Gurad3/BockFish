package main

var knightAttacks = [][]int{
	{1, 2}, {2, 1},
	{2, -1}, {1, -2},
	{-1, -2}, {-2, -1},
	{-2, 1}, {-1, 2},
}

var kingAttacks = [][]int{
	{0, 1}, {1, 1},
	{1, 0}, {1, -1},
	{0, -1}, {-1, -1},
	{-1, 0}, {-1, 1},
}

var pawnAttack = [][]int{ //Whit pawns, black << 16?
	{1, -1}, {-1, -1},
}

var rookDir = [][]int{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
}

var KnightAttackMasks = [64]uint64{}
var KingAttackMasks = [64]uint64{}

var PawnAttackMasks = [64]uint64{}

var RookAttackMask = [64]uint64{}

func loadPreCalculations() {
	for i := 0; i < 64; i++ {
		CurrentX := i % 8
		CurrentY := i / 8

		CurrentKnightMask := uint64(0)
		for _, offsets := range knightAttacks {
			tmpX := CurrentX + offsets[0]
			tmpY := CurrentY + offsets[1]
			if (tmpX < 8 && tmpX >= 0) && (tmpY < 8 && tmpY >= 0) {

				CurrentKnightMask |= (1 << (tmpY*8 + tmpX))
			}
		}
		KnightAttackMasks[i] = CurrentKnightMask

		CurrentKingMask := 0
		for _, offsets := range kingAttacks {
			tmpX := CurrentX + offsets[0]
			tmpY := CurrentY + offsets[1]
			if (tmpX < 8 && tmpX >= 0) && (tmpY < 8 && tmpY >= 0) {
				CurrentKingMask |= (1 << (tmpY*8 + tmpX))
			}
		}
		KingAttackMasks[i] = uint64(CurrentKingMask)

		CurrentPawnMask := 0
		for _, offsets := range pawnAttack {
			tmpX := CurrentX + offsets[0]
			tmpY := CurrentY + offsets[1]
			if (tmpX < 8 && tmpX >= 0) && (tmpY < 8 && tmpY >= 0) {
				CurrentPawnMask |= (1 << (tmpY*8 + tmpX))
			}
		}
		PawnAttackMasks[i] = uint64(CurrentPawnMask)

		CurrentRookMask := 0
		for _, dir := range rookDir {
			for step := 0; step < 5; step++ {
				tmpX := CurrentX + dir[0]*step
				tmpY := CurrentY + dir[1]*step
				if tmpX >= 0 && tmpX < 8 && tmpY >= 0 && tmpY < 8 {
					CurrentRookMask |= (1 << (tmpY*8 + tmpX))
				}

			}
		}

	}
}
