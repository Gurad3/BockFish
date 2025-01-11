package neural

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Network struct {
	layers []Layer

	learn_rate       float64
	learn_rate_decay float64
	regularization   float64
	momentum         float64

	input_shape []int
	input_size  int
	output_size int

	cost_interface Cost
	file_name      string
}

func (shelf *Network) init() {

	prev_layer_size := shelf.input_shape
	shelf.input_size = 1
	for _, v := range shelf.input_shape {
		shelf.input_size *= v
	}

	for layerID, layer := range shelf.layers {
		layer.init(layerID+1, prev_layer_size)
		prev_layer_size = layer.get_size()
	}
}

func (shelf *Network) init_new_weights() {
	xavierRange := math.Sqrt(6 / float64(shelf.input_size+shelf.output_size))
	r := rand.New(rand.NewSource(time.Now().Unix()))
	//r := rand.New(rand.NewSource(2))
	for _, layer := range shelf.layers {
		layer.init_new_weights(xavierRange, *r)
	}
}

func (shelf *Network) print_weights() {
	for _, layer := range shelf.layers {
		layer.print_weights()
	}
}

var g int = 0

func (shelf *Network) forward(mim *MiM, data []float64) {
	mim.data_flat = &data
	mim.layers_out[0] = data
	mim.layers_out_non_activated[0] = data

	for _, layer := range shelf.layers {
		layer.forward(mim)
	}
}

func (shelf *Network) backprop(mim *MiM, labels []float64) {
	mim.data_flat = shelf.get_output_ddx(mim, labels)

	for layerID := len(shelf.layers) - 1; layerID > 0; layerID-- {
		shelf.layers[layerID].backprop(mim, shelf.layers[layerID-1].get_act_interface())
	}
	shelf.layers[0].backprop(mim, shelf.layers[0].get_act_interface())
}

func (shelf *Network) get_output_ddx(mim *MiM, labels []float64) *[]float64 {
	gradiants := make([]float64, len(labels))

	for outID, output := range *mim.data_flat {
		gradiants[outID] = shelf.layers[len(shelf.layers)-1].get_act_interface().ddx(mim.layers_out_non_activated[len(shelf.layers)][outID])
		gradiants[outID] *= shelf.cost_interface.ddx(output, labels[outID])
	}

	return &gradiants
}

func (shelf *Network) apply_gradients(batch_size int) {
	for _, layer := range shelf.layers {
		layer.apply_gradients(shelf.learn_rate, batch_size, shelf.regularization, shelf.momentum)
	}
}

func cliHandler() {
	reader := bufio.NewReader(os.Stdin)
	var net *Network
	var mim *MiM
	// Loop to read commands from the GUI
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(0)
			continue
		}
		command = strings.TrimSpace(command)
		// Parse the command
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "load":
			net = load_from_json(parts[1])
			mim = new(MiM)
			mim.init(net)

			fmt.Println("loadok")

		case "eval":
			jsonData := strings.Join(parts[1:], " ")
			var data []float64

			// Parse the JSON data
			err := json.Unmarshal([]byte(jsonData), &data)
			if err != nil {

			}
			net.forward(mim, data)
			//response, _ := json.Marshal(*mim.data_flat)
			fmt.Println(*mim.data_flat)

		case "hello":
			fmt.Println("hellook" + parts[1])
		case "isready":
			fmt.Println("readyok")
		case "quit":
			os.Exit(0)
		case "exit":
			os.Exit(0)
		}
	}

}
