package neural

type MiM struct {
	data_flat                *[]float64
	layers_out               [][]float64 //First = input data
	layers_out_non_activated [][]float64
	layers_dimentions        [][]int
}

func (shelf *MiM) init(net *Network) {
	length := len(net.layers) + 1

	shelf.layers_dimentions = make([][]int, length)
	shelf.layers_dimentions[0] = []int{1, 28, 28} // TODO GLÖM INTE BORT

	shelf.layers_out = make([][]float64, length)
	shelf.layers_out_non_activated = make([][]float64, length)

	for fake_layer_id, layer := range net.layers {
		layerDim := 1
		for _, v := range layer.get_size() {
			layerDim *= v
		}

		layerID := fake_layer_id + 1

		shelf.layers_out[layerID] = make([]float64, layerDim)
		shelf.layers_out_non_activated[layerID] = make([]float64, layerDim)
		shelf.layers_dimentions[layerID] = layer.get_size()
	}
}
