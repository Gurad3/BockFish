package neural

import (
	"math/rand"
)

type Layer interface {
	forward(*MiM)
	backprop(mim *MiM, prev_layer_act Activation)
	get_act_interface() Activation
	apply_gradients(learn_rate float64, batch_size int, regularization float64, momentum float64)

	init(layerID int, prev_layer_size []int)
	init_new_weights(xavierRange float64, r rand.Rand)
	load_weights(flat_weights []float64)
	load_biases(flat_biases []float64)
	get_weights() []float64
	get_biases() []float64
	get_init_vals() []int

	get_size() []int
	get_name() string

	print_weights()
	debug_print()
}
