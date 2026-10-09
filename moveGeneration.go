package main

func (board *board) getLegalMovesA() []uint16 {

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves, false)]
	return moves

	//return board.getLegalMoves()
}

//func (board *board) getLegalMoves() []uint16 {

func (board *board) getLegalMoves(legalMoves []uint16, quin bool) int {

	//legalMoves := make([]uint16, 0, 218) //max 218
	colID := 1
	if board.isWhiteTurn {
		colID = 0
	}
	opCOl := 1 - colID

	opPieceMask := board.pieceMask[opCOl]
	colPieceMask := board.pieceMask[colID]
	fullPieceMask := board.pieceMask[opCOl] | board.pieceMask[colID]

	for TypeID := 1; TypeID < 7; TypeID++ {

		mask := board.pieceMasks[colID][TypeID]

		for mask != 0 {

			square := popLSB(mask)
			mask &= mask - 1

			moveMask := uint64(0)

			if TypeID == 1 { //Pawn
				if colID == 0 {
					moveMask |= ((mask_PawnAttack[square] >> 8) & opPieceMask)

					if square < 32 && board.lastMove&FlagMask == PawnTwoUpFlag { //Enpassent
						mMask := (mask_PawnAttack[square]) & (1 << ((board.lastMove & MoveTargetMask) >> 6))
						if mMask > 0 {
							newMove := uint16(popLSB(mMask>>8)) << 6
							newMove += uint16(square)
							legalMoves = append(legalMoves, newMove|EnPassant|CapFlag)
						}
					}

					if !quin {

						if (1<<(square-8))&(fullPieceMask) == 0 { //Nothin infront og pawn
							moveMask |= (1 << (square - 8))

							if square > 47 { //not moved, check if dual is possible
								if (1<<(square-16))&(fullPieceMask) == 0 { //Nothin infront og pawn
									newMove := uint16(square-16) << 6
									newMove += uint16(square)
									legalMoves = append(legalMoves, newMove|PawnTwoUpFlag)
								}
							}
						}

					} else {
						if square < 8 && (1<<(square-8))&(fullPieceMask) == 0 {
							newMove := uint16(square-8) << 6
							newMove += uint16(square)
							for i := uint16(4); i < 8; i++ {
								legalMoves = append(legalMoves, newMove|(i<<12))
							}
						}
					}

				} else {
					moveMask |= ((mask_PawnAttack[square] << 8) & opPieceMask)

					if square > 31 && board.lastMove&FlagMask == PawnTwoUpFlag {
						mMask := (mask_PawnAttack[square]) & (1 << ((board.lastMove & MoveTargetMask) >> 6))
						if mMask > 0 {
							newMove := uint16(popLSB(mMask<<8)) << 6
							newMove += uint16(square)
							legalMoves = append(legalMoves, newMove|EnPassant|CapFlag)
						}
					}

					if !quin {

						if (1<<(square+8))&(fullPieceMask) == 0 { //Nothin infront og pawn
							moveMask |= (1 << (square + 8))

							if square < 16 { //not moved, check if dual is possible
								if (1<<(square+16))&(fullPieceMask) == 0 { //Nothin infront og pawn
									newMove := uint16(square+16) << 6
									newMove += uint16(square)
									legalMoves = append(legalMoves, newMove|PawnTwoUpFlag)
								}
							}
						}

					} else {
						if square > 55 && (1<<(square+8))&(fullPieceMask) == 0 {
							newMove := uint16(square+8) << 6
							newMove += uint16(square)
							for i := uint16(4); i < 8; i++ {
								legalMoves = append(legalMoves, newMove|(i<<12))
							}
						}
					}

				}

			} else if TypeID == 2 { //Knight
				moveMask |= mask_KnightAttack[square] //Cretae movemask

			} else if TypeID == 6 { //King
				moveMask |= mask_KingAttack[square] //Cretae movemask

				if !quin && board.castling != 0 {
					if colID == 0 {
						if board.castling&0b0100 > 0 {
							if fullPieceMask&WKCastleMask == WKCastleFlag && !board.isAttacked(opCOl, WKCastleCheckMask) {
								newMove := uint16(63) << 6
								newMove += uint16(square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}
						if board.castling&0b1000 > 0 {
							if fullPieceMask&WQCastleMask == WQCastleFlag && !board.isAttacked(opCOl, WQCastleCheckMask) {
								newMove := uint16(56) << 6
								newMove += uint16(square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}
					} else {

						if board.castling&0b0001 > 0 {
							if fullPieceMask&BKCastleMask == BKCastleFlag && !board.isAttacked(opCOl, BKCastleCheckMask) {
								newMove := uint16(7) << 6
								newMove += uint16(square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}
						if board.castling&0b0010 > 0 {
							if fullPieceMask&BQCastleMask == BQCastleFlag && !board.isAttacked(opCOl, BQCastleCheckMask) {
								newMove := uint16(0) << 6
								newMove += uint16(square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}

					}
				}

			} else if TypeID == 3 { //Bishop
				moveMask |= BishMagicDict[square][((mask_BishAttack[square]&(fullPieceMask))*bishopMagics[square])>>bishopShifts[square]]
			} else if TypeID == 4 { //Rook
				moveMask |= RookMagicDict[square][((mask_RookAttack[square]&(fullPieceMask))*rookMagics[square])>>rookShifts[square]]
			} else if TypeID == 5 { //Queen
				moveMask |= RookMagicDict[square][((mask_RookAttack[square]&(fullPieceMask))*rookMagics[square])>>rookShifts[square]]
				moveMask |= BishMagicDict[square][((mask_BishAttack[square]&(fullPieceMask))*bishopMagics[square])>>bishopShifts[square]]
			}

			moveMask |= colPieceMask
			moveMask ^= colPieceMask

			if quin {
				moveMask &= opPieceMask
			}

			for moveMask != 0 {
				targetIndex := popLSB(moveMask)
				targetMask := uint64(1) << targetIndex
				moveMask ^= targetMask

				flag := uint16(0)
				newMove := uint16(targetIndex) << 6
				newMove += uint16(square)

				if targetMask&opPieceMask > 0 {
					flag = CapFlag
				}

				if TypeID == 1 {

					if targetIndex < 8 || targetIndex > 55 { //Pawn is promoting

						//ALL
						for i := uint16(4); i < 8; i++ {
							legalMoves = append(legalMoves, newMove|(i<<12)|flag)
						}

						//Knight and Queen
						// legalMoves = append(legalMoves, newMove|(4<<12)|flag)
						// legalMoves = append(legalMoves, newMove|(7<<12)|flag)

						continue
					}
				}

				legalMoves = append(legalMoves, newMove|flag)
			}

		}

	}

	//return legalMoves

	return len(legalMoves)
}

func (board *board) filterOutPsudoLegal(moves []uint16) []uint16 {
	newList := make([]uint16, 0, cap(moves))
	for _, mo := range moves {

		board.PlayMove(mo)
		if !board.isCheck(!board.isWhiteTurn) {
			newList = append(newList, mo)
		}
		board.UndoMove(mo)

	}
	return newList
}

func (board *board) isCheckA(col bool) bool {

	colID := 1
	if col {
		colID = 0
	}
	// Get target king's position
	targetIndex := popLSB(board.pieceMasks[colID][6])

	if targetIndex == 64 {
		return false
	}

	if col {
		return (board.squareAttackNum[targetIndex] & 0b00001111) > 0
	}
	return (board.squareAttackNum[targetIndex] & 0b11110000) > 0
}

func (board *board) isCheck(col bool) bool {
	colID := 1
	if col {
		colID = 0
	}
	opCol := 1 - colID
	// Get target king's position
	targetIndex := popLSB(board.pieceMasks[colID][6])
	if targetIndex == 64 {
		return false
	}

	// Precompute combined opponent piece masks
	allPieces := board.pieceMask[0] | board.pieceMask[1]
	opponentKnights := board.pieceMasks[opCol][2]
	opponentKings := board.pieceMasks[opCol][6]
	opponentPawns := board.pieceMasks[opCol][1]
	opponentBishopsQueens := board.pieceMasks[opCol][3] | board.pieceMasks[opCol][5]
	opponentRooksQueens := board.pieceMasks[opCol][4] | board.pieceMasks[opCol][5]

	// Knight and King attacks
	if (opponentKnights&mask_KnightAttack[targetIndex])|(opponentKings&mask_KingAttack[targetIndex]) > 0 {
		return true
	}

	// Pawn attacks
	if opCol == 1 { // Black pawns attacking
		if opponentPawns&(mask_PawnAttack[targetIndex]>>8) > 0 {
			return true
		}
	} else { // White pawns attacking
		if opponentPawns&(mask_PawnAttack[targetIndex]<<8) > 0 {
			return true
		}
	}

	// Bishop and Queen attacks
	bishopIndex := ((mask_BishAttack[targetIndex] & allPieces) * bishopMagics[targetIndex]) >> bishopShifts[targetIndex]
	if BishMagicDict[targetIndex][bishopIndex]&opponentBishopsQueens > 0 {
		return true

	}

	// Rook and Queen attacks
	rookIndex := ((mask_RookAttack[targetIndex] & allPieces) * rookMagics[targetIndex]) >> rookShifts[targetIndex]
	if RookMagicDict[targetIndex][rookIndex]&opponentRooksQueens > 0 {
		return true

	}
	return false
}

func (board *board) isAttacked(opCol int, mask uint64) bool {
	for mask > 0 {
		targetIndex := popLSB(mask)
		mask &= mask - 1

		tmpMask := (board.pieceMasks[opCol][2] & mask_KnightAttack[targetIndex]) | (board.pieceMasks[opCol][6] & mask_KingAttack[targetIndex]) //King and Knights

		if opCol == 1 { // Col == White == 0
			tmpMask |= board.pieceMasks[opCol][1] & (mask_PawnAttack[targetIndex] >> 8) //Pawns
		} else {
			tmpMask |= board.pieceMasks[opCol][1] & (mask_PawnAttack[targetIndex] << 8) //Pawns
		}
		if tmpMask > 0 {
			return true
		}

		if BishMagicDict[targetIndex][((mask_BishAttack[targetIndex]&(board.pieceMask[0]|board.pieceMask[1]))*bishopMagics[targetIndex])>>bishopShifts[targetIndex]]&(board.pieceMasks[opCol][3]|board.pieceMasks[opCol][5]) > 0 {
			return true
		}

		//Rook + Queen
		if RookMagicDict[targetIndex][((mask_RookAttack[targetIndex]&(board.pieceMask[0]|board.pieceMask[1]))*rookMagics[targetIndex])>>rookShifts[targetIndex]]&(board.pieceMasks[opCol][4]|board.pieceMasks[opCol][5]) > 0 {
			return true
		}
	}

	return false
}

//var PromotFlags [4]move = [4]move{PromoteToKnightFlag, PromoteToBishopFlag, PromoteToRookFlag, PromoteToQueenFlag}

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king


*/
