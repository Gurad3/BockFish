package neural

import (
	"fmt"
	"math/rand"
)

type ConvLayer struct {
	act_interface Activation
	layerID       int
	out_size      []int
	input_shape   []int
	layer_type    string

	depth       int
	kernel_size int

	input_depth  int
	input_width  int
	input_height int

	output_width  int
	output_height int

	filters []Filter
}

type Filter struct {
	kernels [][][]float64
	bias    float64

	kernel_gradients [][][]float64
	bias_gradient    float64

	weight_velocities [][][]float64
	bias_velocitiy    float64
}

func (shelf *ConvLayer) init(layerID int, prev_layer_size []int) {
	//Init all layer arrays sizes, Sets bias AND WEIGHTS to 0
	shelf.layer_type = "ConvLayer"
	shelf.layerID = layerID

	shelf.input_depth = prev_layer_size[0]
	shelf.input_width = prev_layer_size[1]
	shelf.input_height = prev_layer_size[2]
	shelf.input_shape = prev_layer_size

	shelf.output_width = shelf.input_width - shelf.kernel_size + 1
	shelf.output_height = shelf.input_height - shelf.kernel_size + 1

	shelf.out_size = []int{shelf.depth, shelf.output_height, shelf.output_width}

	shelf.filters = make([]Filter, shelf.depth)

	for i := 0; i < shelf.depth; i++ {
		//kernels
		shelf.filters[i].kernels = make([][][]float64, shelf.input_depth)
		shelf.filters[i].kernel_gradients = make([][][]float64, shelf.input_depth)
		shelf.filters[i].weight_velocities = make([][][]float64, shelf.input_depth)
		shelf.filters[i].bias = 0
		shelf.filters[i].bias_gradient = 0
		shelf.filters[i].bias_velocitiy = 0
		for j := 0; j < shelf.input_depth; j++ {

			shelf.filters[i].kernels[j] = make([][]float64, shelf.kernel_size)
			shelf.filters[i].kernel_gradients[j] = make([][]float64, shelf.kernel_size)
			shelf.filters[i].weight_velocities[j] = make([][]float64, shelf.kernel_size)

			for k := 0; k < shelf.kernel_size; k++ {
				shelf.filters[i].kernels[j][k] = make([]float64, shelf.kernel_size)
				shelf.filters[i].kernel_gradients[j][k] = make([]float64, shelf.kernel_size)
				shelf.filters[i].weight_velocities[j][k] = make([]float64, shelf.kernel_size)
			}
		}

	}
}

func (shelf *ConvLayer) init_new_weights(xavierRange float64, r rand.Rand) {
	//Give each weights new random weights (Currently 0)

	for i := 0; i < shelf.depth; i++ {
		// shelf.filters[i].bias = initWeightXavierUniform(xavierRange, r)
		for j := 0; j < shelf.input_depth; j++ {
			for k := 0; k < shelf.kernel_size; k++ {
				for l := 0; l < shelf.kernel_size; l++ {
					shelf.filters[i].kernels[j][k][l] = initWeightXavierUniform(xavierRange, r)

					//shelf.filters[i].kernels[j][k][l] = tmp_manualKernel[i][k][l]
					//shelf.filters[i].kernels[j][k][l] = tmp_manualKernel[i][k][l] + initWeightXavierUniform(xavierRange, r)*0.05
				}
			}
		}
	}
}

func (shelf *ConvLayer) forward(mim *MiM) {

	matrix := *mim.data_flat

	for f, filter := range shelf.filters {
		outRangeStart := f * shelf.output_width * shelf.output_height
		outRangeEnd := (f + 1) * shelf.output_width * shelf.output_height
		filter.correlation(matrix, mim.layers_out[shelf.layerID][outRangeStart:outRangeEnd], mim.layers_out_non_activated[shelf.layerID][outRangeStart:outRangeEnd], shelf.act_interface, shelf.input_shape, shelf.out_size[1:])
	}

	mim.data_flat = &mim.layers_out[shelf.layerID]
}

func (shelf *ConvLayer) backprop(mim *MiM, prev_layer_act Activation) {

	for filterID := range shelf.filters {
		shelf.filters[filterID].compute_loss_kernel_gradient(mim, shelf.output_width, shelf.output_height, shelf.layerID, filterID, shelf.input_shape)
	}

	if shelf.layerID == 1 {
		return
	}

	prevOut := mim.layers_out[shelf.layerID-1]
	for i := 0; i < shelf.input_depth*shelf.input_height*shelf.input_width; i++ {
		prevOut[i] = 0
	}

	for filterID := range shelf.filters {
		shelf.filters[filterID].compute_output_gradient(mim, shelf.output_width, shelf.output_height, shelf.layerID, filterID, shelf.input_shape, prev_layer_act)
	}

	mim.data_flat = &mim.layers_out[shelf.layerID-1]
}
func (shelf *Filter) compute_loss_kernel_gradient(mim *MiM, O_W int, O_H int, layerID int, filterID int, inp_shape []int) {
	// Compute convolution between the input this layer recieved, and the matrix respresenting
	// the partial derivates of the Cost function with respect to this layers output.
	out_grade := *mim.data_flat
	filterOffset := filterID * O_H * O_W

	prev_layer_out := mim.layers_out[layerID-1]

	for c := 0; c < len(shelf.kernels); c++ {
		for i := 0; i < len(shelf.kernels[0]); i++ {
			for j := 0; j < len(shelf.kernels[0][0]); j++ {

				for k := 0; k < O_H; k++ {
					for l := 0; l < O_W; l++ {
						shelf.kernel_gradients[c][i][j] += prev_layer_out[c*inp_shape[1]*inp_shape[2]+(i+k)*inp_shape[1]+j+l] * out_grade[filterOffset+k*O_W+l]
					}
				}
			}
		}
	}

	for k := 0; k < O_W; k++ {
		for l := 0; l < O_H; l++ {
			//shelf.bias_gradient += (*mim.data_3d)[filterID][k][l]
			shelf.bias_gradient += out_grade[filterOffset+k*O_W+l]
			// fmt.Println("g", shelf.bias_gradient)
		}
	}
}

func (shelf *Filter) compute_output_gradient(mim *MiM, O_W int, O_H int, layerID int, filterID int, inp_shape []int, prev_act_out Activation) {
	// Kommer skickas vidare bakåt till nästa lager i backpropagation följden.
	// Outputen här blir matrisen som representerar partiella derivatorna för Cost med respekt till
	// lagrets output. Detta kommer då sedan användas i `compute_loss_kernel_gradient()`.
	// Använder även det föregående lagrets `compute_output_gradient`
	out_grade := *mim.data_flat
	filterOffset := filterID * O_H * O_W
	K_Y := len(shelf.kernels[0])
	K_X := len(shelf.kernels[0][0])
	for c := 0; c < inp_shape[0]; c++ {

		channelOffset := c * inp_shape[1] * inp_shape[2]

		for i := 0; i < inp_shape[1]; i++ {

			for j := 0; j < inp_shape[2]; j++ {

				for k := 0; k < K_Y; k++ {
					for l := 0; l < K_X; l++ {

						///TODO: Mer effetkiv statment, (Exit early)
						if !((j+l) < (K_X-1) || (i+k) < (K_Y-1) || (j+l) > (K_X-1)+O_W-1 || (i+k) > (K_Y-1)+O_H-1) {

							mim.layers_out[layerID-1][channelOffset+i*inp_shape[1]+j] += out_grade[filterOffset+(i+k-(K_Y-1))*O_W+j+l-(K_X-1)] * shelf.kernels[c][K_Y-1-k][K_X-1-l]
						}

						//flipped[i][j] = kernel[rows-1-i][cols-1-j]
					}
				}
			}
		}
	}

}

func (shelf *Filter) correlation(matrix []float64, target_activated []float64, target_non_activated []float64, activation Activation, input_shape []int, out_shape []int) {

	for i := 0; i < out_shape[0]; i++ {
		for j := 0; j < out_shape[1]; j++ {
			sum := shelf.bias
			for kernel_index := 0; kernel_index < len(shelf.kernels); kernel_index++ {
				kernel_offset := kernel_index * input_shape[1] * input_shape[2]
				for k := 0; k < len(shelf.kernels[kernel_index]); k++ {
					for l := 0; l < len(shelf.kernels[kernel_index][0]); l++ {
						sum += matrix[kernel_offset+(i+k)*input_shape[1]+(j+l)] * shelf.kernels[kernel_index][k][l]
					}
				}
			}
			target_activated[i*out_shape[0]+j] = activation.call(sum)
			target_non_activated[i*out_shape[0]+j] = sum

		}
	}
}

func (shelf *ConvLayer) apply_gradients(learn_rate float64, batch_size int, regularization float64, momentum float64) {
	mult := learn_rate / float64(batch_size)
	weight_decay := (1 - regularization*mult)

	for i := 0; i < shelf.depth; i++ {
		for j := 0; j < shelf.input_depth; j++ {
			for k := 0; k < shelf.kernel_size; k++ {
				for l := 0; l < shelf.kernel_size; l++ {

					//shelf.filters[i].kernels[j][k][l] -= shelf.filters[i].kernel_gradients[j][k][l] * mult

					velocity := shelf.filters[i].weight_velocities[j][k][l]*momentum - shelf.filters[i].kernel_gradients[j][k][l]*mult
					shelf.filters[i].weight_velocities[j][k][l] = velocity
					shelf.filters[i].kernels[j][k][l] = shelf.filters[i].kernels[j][k][l]*weight_decay + velocity
					shelf.filters[i].kernel_gradients[j][k][l] = 0
				}
			}

		}
		velocity := shelf.filters[i].bias_velocitiy*momentum - shelf.filters[i].bias_gradient*mult
		shelf.filters[i].bias_velocitiy = velocity
		shelf.filters[i].bias += velocity
		shelf.filters[i].bias_gradient = 0
	}

}

func (shelf *ConvLayer) load_weights(flat_weights []float64) {
	for i := 0; i < shelf.depth; i++ {
		for j := 0; j < shelf.input_depth; j++ {
			for k := 0; k < shelf.kernel_size; k++ {
				for l := 0; l < shelf.kernel_size; l++ {
					shelf.filters[i].kernels[j][k][l] = flat_weights[i*shelf.input_depth*shelf.kernel_size*shelf.kernel_size+j*shelf.kernel_size*shelf.kernel_size+k*shelf.kernel_size+l]
				}
			}
		}
	}
}
func (shelf *ConvLayer) get_weights() []float64 {
	weights := make([]float64, shelf.depth*shelf.input_depth*shelf.kernel_size*shelf.kernel_size)

	for i := 0; i < shelf.depth; i++ {
		for j := 0; j < shelf.input_depth; j++ {
			for k := 0; k < shelf.kernel_size; k++ {
				for l := 0; l < shelf.kernel_size; l++ {
					weights[i*shelf.input_depth*shelf.kernel_size*shelf.kernel_size+j*shelf.kernel_size*shelf.kernel_size+k*shelf.kernel_size+l] = shelf.filters[i].kernels[j][k][l]
				}
			}
		}
	}

	return weights
}

func (shelf *ConvLayer) load_biases(flat_biases []float64) {
	for i := 0; i < shelf.depth; i++ {
		shelf.filters[i].bias = flat_biases[i]
	}
}
func (shelf *ConvLayer) get_biases() []float64 {
	biases := make([]float64, shelf.depth)
	for filterId, filter := range shelf.filters {
		biases[filterId] = filter.bias
	}
	return biases
}
func (shelf *ConvLayer) print_weights() {

}

func (shelf *ConvLayer) get_act_interface() Activation {
	return shelf.act_interface
}
func (shelf *ConvLayer) get_size() []int {
	return shelf.out_size
}
func (shelf *ConvLayer) get_name() string {
	return shelf.layer_type
}

func (shelf *ConvLayer) get_init_vals() []int {
	return []int{shelf.kernel_size, shelf.depth}
}

func (shelf *ConvLayer) debug_print() {
	fmt.Println("---")
	for i := 0; i < shelf.depth; i++ {
		for j := 0; j < shelf.input_depth; j++ {
			for k := 0; k < shelf.kernel_size; k++ {
				for l := 0; l < shelf.kernel_size; l++ {
					fmt.Println(shelf.filters[i].kernels[j][k][l])
				}
			}
		}
	}
	fmt.Println("------")
}
