package main

func (board *board) getLegalMoves(onlyCaptures bool) []move {

	legalMoves := make([]move, 0, 218)

	for colID, AllPieceArray := range board.pieceArrays {
		for TypeID, pieceArray := range AllPieceArray {
			for _, piece := range pieceArray {

				var moveMask uint64

				opCOl := 0
				if colID == 0 {
					opCOl = 1
				}

				if TypeID == 1 { //Pawn
					if piece.isWhite {
						moveMask = (PawnAttackMasks[piece.square] & board.pieceMask[opCOl])

						if (1<<(piece.square-8))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
							moveMask |= (1 << (piece.square - 8))

							if piece.square > 47 { //not moved, check if dual is possible
								if (1<<(piece.square-16))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
									moveMask |= (1 << (piece.square - 16))
								}
							}

						}

					} else {
						moveMask = ((PawnAttackMasks[piece.square] << 16) & board.pieceMask[opCOl])

						if (1<<(piece.square+8))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
							moveMask |= (1 << (piece.square + 8))

							if piece.square < 16 { //not moved, check if dual is possible
								if (1<<(piece.square+16))&(board.pieceMask[opCOl]|board.pieceMask[colID]) == 0 { //Nothin infront og pawn
									moveMask |= (1 << (piece.square + 16))
								}
							}
						}
					}

				} else if TypeID == 2 { //Knight
					moveMask = (KnightAttackMasks[piece.square] & board.pieceMask[colID]) ^ KnightAttackMasks[piece.square] //Cretae movemask
				} else if TypeID == 6 { //King
					moveMask = (KingAttackMasks[piece.square] & board.pieceMask[colID]) ^ KingAttackMasks[piece.square] //Cretae movemask
				}

				for i := move(0); i < 64; i++ {
					if moveMask&1 == 1 {
						var newMove move

						newMove = i << 6
						newMove += move(piece.square)

						legalMoves = append(legalMoves, newMove)
					}
					moveMask >>= 1
				}

			}
		}
	}

	return legalMoves
}

/*

0 = empty
1 = pawn
2 = knight
3 = bishop
4 = rook
5 = queen
6 = king

*/
