package main

func (board *board) getLegalMovesA() []uint16 {

	moves := make([]uint16, 0, 218)
	moves = moves[:board.getLegalMoves(moves)]
	return moves

	//return board.getLegalMoves()
}

//func (board *board) getLegalMoves() []uint16 {

func (board *board) getLegalMoves(legalMoves []uint16) int {

	//legalMoves := make([]uint16, 0, 218) //max 218
	colID := 1
	opCOl := 0
	if board.isWhiteTurn {
		colID = 0
		opCOl = 1
	}
	opPieceMask := board.pieceMask[opCOl]
	colPieceMask := board.pieceMask[colID]

	fullPieceMask := board.pieceMask[opCOl] | board.pieceMask[colID]

	for TypeID, pieceArray := range board.pieceArrays[colID] {
		for _, piece := range pieceArray {
			if piece.dead {
				continue
			}

			moveMask := uint64(0)
			pawnUpTwo := false

			if TypeID == 1 { //Pawn
				if piece.isWhite {
					moveMask |= ((mask_PawnAttack[piece.square] >> 8) & opPieceMask)

					if (1<<(piece.square-8))&(fullPieceMask) == 0 { //Nothin infront og pawn
						moveMask |= (1 << (piece.square - 8))

						if piece.square > 47 { //not moved, check if dual is possible
							if (1<<(piece.square-16))&(fullPieceMask) == 0 { //Nothin infront og pawn
								moveMask |= (1 << (piece.square - 16))
								pawnUpTwo = true
							}
						}

					}

				} else {
					moveMask |= ((mask_PawnAttack[piece.square] << 8) & opPieceMask)

					if (1<<(piece.square+8))&(fullPieceMask) == 0 { //Nothin infront og pawn
						moveMask |= (1 << (piece.square + 8))

						if piece.square < 16 { //not moved, check if dual is possible
							if (1<<(piece.square+16))&(fullPieceMask) == 0 { //Nothin infront og pawn
								moveMask |= (1 << (piece.square + 16))
								pawnUpTwo = true
							}
						}
					}
				}

			} else if TypeID == 2 { //Knight
				moveMask |= (mask_KnightAttack[piece.square] | colPieceMask) ^ colPieceMask //Cretae movemask

			} else if TypeID == 6 { //King
				moveMask |= (mask_KingAttack[piece.square] | colPieceMask) ^ colPieceMask //Cretae movemask

				if board.castling != 0 {
					if colID == 0 {
						if board.castling&0b0100 > 0 {
							if fullPieceMask&WKCastleMask == WKCastleFlag && !board.isAttacked(opCOl, WKCastleCheckMask) {
								newMove := uint16(63) << 6
								newMove += uint16(piece.square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}
						if board.castling&0b1000 > 0 {
							if fullPieceMask&WQCastleMask == WQCastleFlag && !board.isAttacked(opCOl, WQCastleCheckMask) {
								newMove := uint16(56) << 6
								newMove += uint16(piece.square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}
					} else {

						if board.castling&0b0001 > 0 {
							if fullPieceMask&BKCastleMask == BKCastleFlag && !board.isAttacked(opCOl, BKCastleCheckMask) {
								newMove := uint16(7) << 6
								newMove += uint16(piece.square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}
						if board.castling&0b0010 > 0 {
							if fullPieceMask&BQCastleMask == BQCastleFlag && !board.isAttacked(opCOl, BQCastleCheckMask) {
								newMove := uint16(0) << 6
								newMove += uint16(piece.square)
								legalMoves = append(legalMoves, newMove|CastleFlag)
							}
						}

					}
				}

			} else if TypeID == 3 { //Bishop
				moveMask |= (BishMagicDict[piece.square][((mask_BishAttack[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]] | colPieceMask) ^ colPieceMask
			} else if TypeID == 4 { //Rook
				moveMask |= (RookMagicDict[piece.square][((mask_RookAttack[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]] | colPieceMask) ^ colPieceMask
			} else if TypeID == 5 { //Queen
				moveMask |= RookMagicDict[piece.square][((mask_RookAttack[piece.square]&(fullPieceMask))*rookMagics[piece.square])>>rookShifts[piece.square]]
				moveMask |= BishMagicDict[piece.square][((mask_BishAttack[piece.square]&(fullPieceMask))*bishopMagics[piece.square])>>bishopShifts[piece.square]]
				moveMask |= colPieceMask
				moveMask ^= colPieceMask
			}

			for moveMask > 0 {
				targetIndex := popLSB(moveMask)

				flag := uint16(0)
				newMove := uint16(targetIndex) << 6
				newMove += uint16(piece.square)

				if (1<<targetIndex)&opPieceMask > 0 {
					flag = CapFlag
				}

				if TypeID == 1 {

					if pawnUpTwo {
						if targetIndex <= 39 && targetIndex >= 24 {
							flag = PawnTwoUpFlag
						}

					} else if targetIndex < 8 || targetIndex > 55 { //Pawn is promoting

						for i := uint16(4); i < 8; i++ {
							//	legalMoves = append(legalMoves, newMove|(i<<12)|flag)
						}

						// fmt.Println(piece.square, targetIndex, TypeID, board.pieceMask[0], board.pieceMasks[0][1], board.pieceMask[1], board.pieceMasks[1][1])
						moveMask ^= 1 << targetIndex
						continue
					}
				}

				legalMoves = append(legalMoves, newMove|flag)
				moveMask ^= 1 << targetIndex
			}

		}

	}

	//return legalMoves

	return len(legalMoves)
}

func (board *board) isCheck(col bool) bool {
	colID := 1
	opCol := 0

	if col {
		colID = 0
		opCol = 1
	}

	kingIndex := board.kings[colID].square

	tmpMask := (board.pieceMasks[opCol][2] & mask_KnightAttack[kingIndex]) | (board.pieceMasks[opCol][6] & mask_KingAttack[kingIndex]) //King and Knights
	if col {
		tmpMask |= board.pieceMasks[opCol][1] & (mask_PawnAttack[kingIndex] >> 8) //Pawns
	} else {
		tmpMask |= board.pieceMasks[opCol][1] & (mask_PawnAttack[kingIndex] << 8) //Pawns
	}

	if tmpMask > 0 {
		return true
	}

	//Bish + Queen
	if BishMagicDict[kingIndex][((mask_BishAttack[kingIndex]&(board.pieceMask[0]|board.pieceMask[1]))*bishopMagics[kingIndex])>>bishopShifts[kingIndex]]&(board.pieceMasks[opCol][3]|board.pieceMasks[opCol][5]) > 0 {
		return true
	}

	//Rook + Queen
	if RookMagicDict[kingIndex][((mask_RookAttack[kingIndex]&(board.pieceMask[0]|board.pieceMask[1]))*rookMagics[kingIndex])>>rookShifts[kingIndex]]&(board.pieceMasks[opCol][4]|board.pieceMasks[opCol][5]) > 0 {
		return true
	}

	return false
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

func (board *board) isAttacked(opCol int, mask uint64) bool {
	isAttacked := false
	for mask > 0 && !isAttacked {
		targetIndex := popLSB(mask)
		mask ^= 1 << targetIndex

		tmpMask := (board.pieceMasks[opCol][2] & mask_KnightAttack[targetIndex]) | (board.pieceMasks[opCol][6] & mask_KingAttack[targetIndex]) //King and Knights

		if opCol == 1 { // Col == White == 0
			tmpMask |= board.pieceMasks[opCol][1] & (mask_PawnAttack[targetIndex] >> 8) //Pawns
		} else {
			tmpMask |= board.pieceMasks[opCol][1] & (mask_PawnAttack[targetIndex] << 8) //Pawns
		}

		tmpMask |= RookMagicDict[targetIndex][((mask_RookAttack[targetIndex]&(board.pieceMask[0]|board.pieceMask[1]))*rookMagics[targetIndex])>>rookShifts[targetIndex]] & (board.pieceMasks[opCol][4] | board.pieceMasks[opCol][5])     //Rook + Queen
		tmpMask |= BishMagicDict[targetIndex][((mask_BishAttack[targetIndex]&(board.pieceMask[0]|board.pieceMask[1]))*bishopMagics[targetIndex])>>bishopShifts[targetIndex]] & (board.pieceMasks[opCol][3] | board.pieceMasks[opCol][5]) //Bish + Queen

		isAttacked = tmpMask > 0
	}

	return isAttacked
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
