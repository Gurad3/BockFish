package data_handler

//XOR_TrainDataLabel, XOR_TrainData, XOR_TestDataLabel, XOR_TestData := loadXor()

func Load_xor() ([][]float64, [][]float64, [][]float64, [][]float64) {

	data := [][]float64{
		{0, 1},
		{1, 0},
		{1, 1},
		{0, 0},
	}

	labels := [][]float64{
		{1, 0},
		{1, 0},
		{0, 1},
		{0, 1},
	}
	return labels, data, labels, data
}
