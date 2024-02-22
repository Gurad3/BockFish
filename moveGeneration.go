package main

func (board *board) getLegalMoves(onlyCaptures bool) []move {

	legalMoves := make([]move, 0, 218)

	for colID, AllPieceArray := range board.pieceArrays {
		for TypeID, pieceArray := range AllPieceArray {
			for _, piece := range pieceArray {

				var moveMask uint64

				if TypeID == 1 { //Pawn
					if piece.isWhite {
						moveMask = (PawnAttackMasks[piece.square] & board.pieceMasks[colID][TypeID]) ^ PawnAttackMasks[piece.square] //Cretae movemask
					} else {
						moveMask = ((PawnAttackMasks[piece.square] << 16) & board.pieceMasks[colID][TypeID]) ^ (PawnAttackMasks[piece.square] << 16) //Cretae movemask
					}

				} else if TypeID == 2 { //Knight
					moveMask = (KnightAttackMasks[piece.square] & board.pieceMasks[colID][TypeID]) ^ KnightAttackMasks[piece.square] //Cretae movemask
				} else if TypeID == 6 { //King
					moveMask = (KingAttackMasks[piece.square] & board.pieceMasks[colID][TypeID]) ^ KingAttackMasks[piece.square] //Cretae movemask
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
