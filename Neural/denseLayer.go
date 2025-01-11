package neural

import (
	"fmt"
	"math/rand"
)

type DenseLayer struct {
	act_interface   Activation
	layerID         int
	size            int
	prev_layer_size int
	layer_type      string

	weights [][]float64
	bias    []float64

	weight_gradients [][]float64
	bias_gradients   []float64

	weight_velocities [][]float64
	bias_velocities   []float64
}

func (shelf *DenseLayer) init(layerID int, prev_layer_size []int) {
	//Init all layer arrays sizes, Sets bias AND WEIGHTS to 0
	shelf.layer_type = "DenseLayer"
	shelf.layerID = layerID

	shelf.prev_layer_size = 1
	for _, v := range prev_layer_size {
		shelf.prev_layer_size *= v
	}

	shelf.bias = make([]float64, shelf.size)
	shelf.weights = make([][]float64, shelf.size)

	shelf.bias_gradients = make([]float64, shelf.size)
	shelf.bias_velocities = make([]float64, shelf.size)
	shelf.weight_gradients = make([][]float64, shelf.size)
	shelf.weight_velocities = make([][]float64, shelf.size)

	for i := 0; i < shelf.size; i++ {
		shelf.weights[i] = make([]float64, shelf.prev_layer_size)
		shelf.weight_gradients[i] = make([]float64, shelf.prev_layer_size)
		shelf.weight_velocities[i] = make([]float64, shelf.prev_layer_size)
	}
}

func (shelf *DenseLayer) forward(mim *MiM) {
	data := *mim.data_flat

	layer_out_non_activated := mim.layers_out_non_activated[shelf.layerID]
	layer_out := mim.layers_out[shelf.layerID]
	for neuronID, neuronVal := range shelf.bias {

		for weightID, weight := range shelf.weights[neuronID] {
			neuronVal += data[weightID] * weight
		}

		layer_out_non_activated[neuronID] = neuronVal
		layer_out[neuronID] = shelf.act_interface.call(neuronVal)
	}
	mim.data_flat = &mim.layers_out[shelf.layerID]

}

func (shelf *DenseLayer) backprop(mim *MiM, prev_act_interface Activation) {
	//Grad = mim.request_data(), Gradiants is Cost/out not Cost/Act(Out)
	out_grade := *mim.data_flat

	prevLayerOut := mim.layers_out[shelf.layerID-1]

	for neuronID, neuronGradient := range out_grade {

		neuronWeights := shelf.weight_gradients[neuronID]

		for weightID, prevLayer := range prevLayerOut {

			neuronWeights[weightID] += neuronGradient * prevLayer

		}

		shelf.bias_gradients[neuronID] += neuronGradient
	}

	if shelf.layerID > 1 { //First layerID == 1, behövr inte räkna ut nästa lagers: Cost/Out om vi är på första lagret.

		for prev_neuronID := 0; prev_neuronID < shelf.prev_layer_size; prev_neuronID++ {
			prevLayerOut[prev_neuronID] = 0
		}

		for neuronID, neuronGradient := range out_grade {
			neuronWeights := shelf.weights[neuronID]
			for prev_neuronID := 0; prev_neuronID < shelf.prev_layer_size; prev_neuronID++ {
				prevLayerOut[prev_neuronID] += neuronWeights[prev_neuronID] * neuronGradient
			}
		}
		for prev_neuronID, v := range mim.layers_out_non_activated[shelf.layerID-1] {
			prevLayerOut[prev_neuronID] *= prev_act_interface.ddx(v)
		}

		mim.data_flat = &mim.layers_out[shelf.layerID-1]
	}

}

func (shelf *DenseLayer) apply_gradients(learn_rate float64, batch_size int, regularization float64, momentum float64) {
	mult := learn_rate / float64(batch_size)
	weight_decay := (1 - regularization*mult)

	for neuronID := range shelf.bias {
		velocity := shelf.bias_velocities[neuronID]*momentum - shelf.bias_gradients[neuronID]*mult
		shelf.bias_velocities[neuronID] = velocity
		shelf.bias[neuronID] += velocity
		shelf.bias_gradients[neuronID] = 0

		wt := shelf.weights[neuronID]
		wtg := shelf.weight_gradients[neuronID]

		for weightID, wtgv := range wtg {
			velocity := shelf.weight_velocities[neuronID][weightID]*momentum - wtgv*mult
			shelf.weight_velocities[neuronID][weightID] = velocity

			wt[weightID] = wt[weightID]*weight_decay + velocity
			wtg[weightID] = 0
		}
		// for weightID := range shelf.weights[neuronID] {
		// 	shelf.weights[neuronID][weightID] -= shelf.weights_gradiants[neuronID][weightID] * mult
		// 	shelf.weights_gradiants[neuronID][weightID] = 0
		// }
	}
}

func (shelf *DenseLayer) init_new_weights(xavierRange float64, r rand.Rand) {
	//Give each weights new random weights (Currentlu 0)

	for neuronID := range shelf.bias {
		shelf.bias[neuronID] = 0
		for weightID := range shelf.weights[neuronID] {
			shelf.weights[neuronID][weightID] = initWeightXavierUniform(xavierRange, r)
		}
	}
}

func (shelf *DenseLayer) load_weights(flat_weights []float64) {
	for i := 0; i < shelf.size; i++ {
		for i2 := 0; i2 < shelf.prev_layer_size; i2++ {
			shelf.weights[i][i2] = flat_weights[shelf.prev_layer_size*i+i2]
		}
	}

}
func (shelf *DenseLayer) load_biases(flat_bias []float64) {
	for i := 0; i < shelf.size; i++ {
		shelf.bias[i] = flat_bias[i]
	}
}
func (shelf *DenseLayer) get_weights() []float64 {
	flat_weights := make([]float64, shelf.prev_layer_size*shelf.size)

	for i := 0; i < shelf.size; i++ {
		for i2 := 0; i2 < shelf.prev_layer_size; i2++ {
			flat_weights[shelf.prev_layer_size*i+i2] = shelf.weights[i][i2]
		}
	}

	return flat_weights
}
func (shelf *DenseLayer) get_biases() []float64 {
	return shelf.bias
}

func (shelf *DenseLayer) print_weights() {
	for neuronID := range shelf.bias {
		fmt.Println(shelf.weights[neuronID])
	}
}

func (shelf *DenseLayer) get_size() []int {
	return []int{shelf.size}
}
func (shelf *DenseLayer) get_name() string {
	return shelf.layer_type
}

func (shelf *DenseLayer) get_act_interface() Activation {
	return shelf.act_interface
}

func (shelf *DenseLayer) get_init_vals() []int {
	return []int{}
}

func (shelf *DenseLayer) debug_print() {
	fmt.Println(shelf.bias[0])
}
