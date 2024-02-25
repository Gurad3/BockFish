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

var bishop = [][]int{
	{1, 1}, {-1, 1}, {1, -1}, {-1, -1},
}

var KnightAttackMasks = [64]uint64{}
var KingAttackMasks = [64]uint64{}

var PawnAttackMasks = [64]uint64{}

var RookAttackMask = [64]uint64{}

var RookMagicDict []map[uint64]uint64

func loadPreCalculations() {
	RookMagicDict = make([]map[uint64]uint64, 64)

	for i := 0; i < 64; i++ {
		RookMagicDict[i] = make(map[uint64]uint64)

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

		CurrentRookMask := uint64(0)
		numRookBits := 0
		for _, dir := range rookDir {
			for step := 0; step < 6; step++ {
				tmpX := CurrentX + dir[0]*step
				tmpY := CurrentY + dir[1]*step
<<<<<<< HEAD
				if tmpX >= 0 && tmpX < 8 && tmpY >= 0 && tmpY < 8 {
					CurrentRookMask |= (1 << (tmpY*8 + tmpX))
=======
				if tmpX >= 0 && tmpX <= 7 && tmpY >= 0 && tmpY <= 7 {
					CurrentRookMask |= 1 << (tmpY*8 + tmpX)
					numRookBits++
				} else {
					break
>>>>>>> c53116d431794dfdbcfea996fb1296ade6fa883e
				}
			}
		}
		RookAttackMask[i] = CurrentRookMask

		CurrentBishMask := uint64(0)
		BishNumSetBits := 0
		for _, dir := range bishop {
			for step := 0; step < 6; step++ {
				tmpX := CurrentX + dir[0]*step
				tmpY := CurrentY + dir[1]*step
				if tmpX >= 1 && tmpX <= 6 && tmpY >= 1 && tmpY <= 6 {
					CurrentBishMask |= 1 << (tmpY*8 + tmpX)
					BishNumSetBits++
				} else {
					break
				}
			}
		}

		//Rook Num all = 10
		//Bish num diffrent.

		numRookPattern := 1 << numRookBits

		rookBlockersBitboard := make([]uint64, numRookPattern)

		for patternIndex := 0; patternIndex < numRookPattern; patternIndex++ {
			copyRookMask := CurrentRookMask
			bitIndex := 0
			for copyRookMask > 0 {

				ind := popLSB(copyRookMask)
				copyRookMask ^= 1 << ind

				rookBlockersBitboard[patternIndex] |= uint64((patternIndex>>bitIndex)&1) << ind

				bitIndex++
			}
		}

		for _, position := range rookBlockersBitboard {

			CurrentRookMoveMask := uint64(0)
			for _, dir := range rookDir {
				for step := 0; step < 6; step++ {
					tmpX := CurrentX + dir[0]*step
					tmpY := CurrentY + dir[1]*step
					if tmpX >= 1 && tmpX <= 6 && tmpY >= 1 && tmpY <= 6 {
						CurrentRookMoveMask |= 1 << (tmpY*8 + tmpX)

						if position&1<<(tmpY*8+tmpX) == 1 {
							break
						}

					} else {
						break
					}
				}
			}

			RookMagicDict[i][position] = CurrentRookMoveMask
		}
	}
}
