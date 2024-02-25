package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

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

var RookMagicDict [][]uint64
var BishMagicDict [][]uint64

func loadPreCalculations(searchForMagic bool) {
	RookMagicDict = make([][]uint64, 64)
	BishMagicDict = make([][]uint64, 64)

	var RookAlBlockers = [64][]uint64{}
	var BishAlBlockers = [64][]uint64{}

	for i := 0; i < 64; i++ {
		RookMagicDict[i] = make([]uint64, 1<<(64-rookShifts[i]))
		BishMagicDict[i] = make([]uint64, 1<<(64-bishopShifts[i]))
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

			RookMagicDict[i][(position*rookMagics[i])>>rookShifts[i]] = CurrentRookMoveMask
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

			BishMagicDict[i][(position*bishopMagics[i])>>bishopShifts[i]] = CurrentBishMoveMask
		}

		if searchForMagic {
			RookAlBlockers[i] = rookBlockersBitboard
			BishAlBlockers[i] = bishBlockersBitboard
		}
	}

	if !searchForMagic {
		return
	}

	currentMagic, currentShift, BcurrentMagic, BcurrentShift := loadMagics()
	/*
		currentMagic := rookMagics
		currentShift := rookShifts
		BcurrentMagic := bishopMagics
		BcurrentShift := bishopShifts
	*/
	/*
		currentMagic := make([]uint64, 64)
		currentShift := make([]int, 64)
		BcurrentMagic := make([]uint64, 64)
		BcurrentShift := make([]int, 64)

		for i := 0; i < 64; i++ {
			currentShift[i] = 50
			BcurrentShift[i] = 50
		}
	*/
	//Seaching for bettermagic
	fmt.Println("Starting Searchs")

	for {
		for i := 0; i < 64; i++ {
			seen := make(map[uint64]bool)
			testMagic := uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
			magicWorks := true
			for _, position := range RookAlBlockers[i] {

				if _, ok := seen[(position*testMagic)>>(currentShift[i]+1)]; ok {
					magicWorks = false
					break
				} else {
					seen[(position*testMagic)>>(currentShift[i]+1)] = true
				}

			}

			if magicWorks {
				currentShift[i] += 1
				fmt.Println("(Rook) Square: ", i, " Shifts: ", currentShift[i])
				currentMagic[i] = testMagic
				SaveMagics(currentMagic, currentShift, BcurrentMagic, BcurrentShift)
			}

			seen = make(map[uint64]bool)
			testMagic = uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
			magicWorks = true
			for _, position := range BishAlBlockers[i] {

				if _, ok := seen[(position*testMagic)>>(BcurrentShift[i]+1)]; ok {
					magicWorks = false
					break
				} else {
					seen[(position*testMagic)>>(BcurrentShift[i]+1)] = true
				}

			}

			if magicWorks {
				BcurrentShift[i] += 1
				fmt.Println("(Bish) Square: ", i, " Shifts: ", BcurrentShift[i])
				BcurrentMagic[i] = testMagic
				SaveMagics(currentMagic, currentShift, BcurrentMagic, BcurrentShift)
			}
		}
	}

}

func SaveMagics(Magics []uint64, Shifts []int, BMagics []uint64, BShifts []int) {
	// Two sample arrays

	// Marshalling arrays to JSON
	data := map[string]interface{}{
		"Magics":  Magics,
		"Shifts":  Shifts,
		"BMagics": BMagics,
		"BShifts": BShifts,
	}
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Writing JSON data to a file
	err = os.WriteFile("magics.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

}

func loadMagics() ([]uint64, []int, []uint64, []int) {
	jsonData, _ := os.ReadFile("magics.json")

	// Define structs to unmarshal JSON data into
	var data map[string]interface{}

	// Unmarshal JSON data into structs
	if err := json.Unmarshal(jsonData, &data); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)

	}

	// Extract arrays from the data map
	MagicsTMP := data["Magics"].([]interface{})
	ShiftsTMP := data["Shifts"].([]interface{})
	BMagicsTMP := data["BMagics"].([]interface{})
	BShiftsTMP := data["BShifts"].([]interface{})
	// Convert arrays to their respective types
	var Magics []uint64
	var Shifts []int
	var BMagics []uint64
	var BShifts []int

	for _, item := range MagicsTMP {
		Magics = append(Magics, uint64(item.(float64)))
	}

	for _, item := range ShiftsTMP {
		Shifts = append(Shifts, int(item.(float64)))
	}

	for _, item := range BMagicsTMP {
		BMagics = append(BMagics, uint64(item.(float64)))
	}

	for _, item := range BShiftsTMP {
		BShifts = append(BShifts, int(item.(float64)))
	}

	return Magics, Shifts, BMagics, BShifts
}
