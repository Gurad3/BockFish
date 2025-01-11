package neural

import "BockFish/Neural/data_handler"

func main() {
	//tmp()
	//cliHandler()

	net := CreateNewNetwork()
	train_mnist(net)
}

func train_mnist(net *Network) {
	MNIST_TrainDataLabel, MNIST_TrainData, MNIST_TestDataLabel, MNIST_TestData := data_handler.Load_mnist()
	td := trainer{
		TrainData:      MNIST_TrainData,
		TestData:       MNIST_TestData,
		TrainDataLabel: MNIST_TrainDataLabel,
		TestDataLabel:  MNIST_TestDataLabel,

		batch_size:        50,
		info_milestone:    60_000,
		save_at_milestone: true,

		threded: true,
	}

	net.train_network(td)
}

func CreateNewNetwork() *Network {
	net := new(Network)

	// Hyperparametrar
	net.learn_rate = .1
	net.regularization = 0.01
	net.momentum = 0

	// Filnamnet som nätvärket sparas till. (JSON)
	net.file_name = "MNIST_Example_Net"
	net.cost_interface = &MeanSquare{}

	net.input_shape = []int{1, 28, 28} // Djup, Bredd, Höjd.
	net.output_size = 10               // Antal neuroner i sista lagret.

	net.layers = []Layer{

		// Initialisering av ett convulutional layer
		&ConvLayer{
			act_interface: &RelU{},
			kernel_size:   3,
			depth:         2,
		},

		&ConvLayer{
			act_interface: &RelU{},
			kernel_size:   3,
			depth:         4,
		},

		// Initialisering av ett fully-connected layer
		// DenseLayer är ett alias till fully-connected layer.
		&DenseLayer{
			act_interface: &RelU{},
			size:          140,
		},
		&DenseLayer{
			act_interface: &RelU{},
			size:          40,
		},

		&DenseLayer{
			act_interface: &Sigmoid{},
			size:          10,
		},
	}

	net.init()             // Allokerar minne för nätvärket.
	net.init_new_weights() // Slumpar nya weights

	return net
}
