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

var bishDir = [][]int{
	{1, 1}, {-1, 1}, {1, -1}, {-1, -1},
}

var KnightAttackMasks = [64]uint64{}
var KingAttackMasks = [64]uint64{}

var PawnAttackMasks = [64]uint64{}

var RookAttackMask = [64]uint64{}
var BishAttackMask = [64]uint64{}

var RookMagicDict []map[uint64]uint64
var BishMagicDict []map[uint64]uint64

func loadPreCalculations() {
	RookMagicDict = make([]map[uint64]uint64, 64)
	BishMagicDict = make([]map[uint64]uint64, 64)
	for i := 0; i < 64; i++ {
		RookMagicDict[i] = make(map[uint64]uint64)
		BishMagicDict[i] = make(map[uint64]uint64)
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

		//Rooks
		CurrentRookMask := uint64(0)
		numRookBits := 0
		for _, dir := range rookDir {
			for step := 1; step < 9; step++ {
				tmpX := CurrentX + dir[0]*step
				tmpY := CurrentY + dir[1]*step

				//absX := int(math.Abs(float64(dir[0])))
				//	absY := int(math.Abs(float64(dir[1])))
				if (tmpX == CurrentX || (tmpX >= 1 && tmpX <= 6)) && ((tmpY >= 1 && tmpY <= 6) || tmpY == CurrentY) {
					CurrentRookMask |= 1 << (tmpY*8 + tmpX)
					numRookBits++
				} else {
					break
				}
			}
		}
		RookAttackMask[i] = CurrentRookMask

		numRookPattern := uint64(1) << numRookBits
		rookBlockersBitboard := make([]uint64, numRookPattern)
		for patternIndex := uint64(0); patternIndex < numRookPattern; patternIndex++ {
			rookBlockersBitboard[patternIndex] = 0
			copyRookMask := CurrentRookMask
			bitIndex := 0
			for copyRookMask > 0 {
				ind := popLSB(copyRookMask)
				copyRookMask ^= 1 << ind
				rookBlockersBitboard[patternIndex] |= ((patternIndex >> bitIndex) & 1) << ind
				bitIndex++
			}
		}

		for _, position := range rookBlockersBitboard {

			CurrentRookMoveMask := uint64(0)
			for _, dir := range rookDir {
				for step := 1; step < 9; step++ {
					tmpX := CurrentX + dir[0]*step
					tmpY := CurrentY + dir[1]*step
					if (tmpX >= 0 && tmpX <= 7) && (tmpY >= 0 && tmpY <= 7) {
						CurrentRookMoveMask |= 1 << (tmpY*8 + tmpX)

						if position&(1<<(tmpY*8+tmpX)) > 0 {
							break
						}

					} else {
						break
					}
				}
			}

			RookMagicDict[i][position] = CurrentRookMoveMask
		}

		//Bishop
		CurrentBishMask := uint64(0)
		numBishBits := 0
		for _, dir := range bishDir {
			for step := 1; step < 9; step++ {
				tmpX := CurrentX + dir[0]*step
				tmpY := CurrentY + dir[1]*step

				if tmpX >= 1 && tmpX <= 6 && tmpY >= 1 && tmpY <= 6 {
					CurrentBishMask |= 1 << (tmpY*8 + tmpX)
					numBishBits++
				} else {
					break
				}
			}
		}
		BishAttackMask[i] = CurrentBishMask

		numBishPattern := uint64(1) << numBishBits
		bishBlockersBitboard := make([]uint64, numBishPattern)
		for patternIndex := uint64(0); patternIndex < numBishPattern; patternIndex++ {
			bishBlockersBitboard[patternIndex] = 0
			copyBishMask := CurrentBishMask
			bitIndex := 0
			for copyBishMask > 0 {
				ind := popLSB(copyBishMask)
				copyBishMask ^= 1 << ind
				bishBlockersBitboard[patternIndex] |= ((patternIndex >> bitIndex) & 1) << ind
				bitIndex++
			}
		}

		for _, position := range bishBlockersBitboard {

			CurrentBishMoveMask := uint64(0)
			for _, dir := range bishDir {
				for step := 1; step < 9; step++ {
					tmpX := CurrentX + dir[0]*step
					tmpY := CurrentY + dir[1]*step
					if (tmpX >= 0 && tmpX <= 7) && (tmpY >= 0 && tmpY <= 7) {
						CurrentBishMoveMask |= 1 << (tmpY*8 + tmpX)

						if position&(1<<(tmpY*8+tmpX)) > 0 {
							break
						}

					} else {
						break
					}
				}
			}

			BishMagicDict[i][position] = CurrentBishMoveMask
		}
	}

}
