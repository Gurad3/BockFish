package neural

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type trainer struct {
	TrainData      [][]float64
	TrainDataLabel [][]float64
	TestData       [][]float64
	TestDataLabel  [][]float64

	TrainDataOrignal [][]float64

	train_batches       [][][]float64
	train_label_batches [][][]float64

	batch_size        int
	info_milestone    int
	save_at_milestone bool
	epochs            int

	threded bool
}

func (shelf *Network) train_network(trainData trainer) {

	trainData.TrainDataOrignal = make([][]float64, len(trainData.TrainData))
	fmt.Println(copy(trainData.TrainDataOrignal, trainData.TrainData))

	for i := 0; i < len(trainData.TrainData); i += trainData.batch_size {
		end := i + trainData.batch_size
		if end > len(trainData.TrainData) {
			end = len(trainData.TrainData)
		}
		trainData.train_batches = append(trainData.train_batches, trainData.TrainData[i:end])
		trainData.train_label_batches = append(trainData.train_label_batches, trainData.TrainDataLabel[i:end])
	}

	totalSamples := 0
	mim := new(MiM)
	mim.init(shelf)

	running := true

	if trainData.threded {
		mimArray := make([]MiM, trainData.batch_size)
		for i := 0; i < trainData.batch_size; i++ {
			mimArray[i].init(shelf)
		}
		var wg sync.WaitGroup

		for running {
			for batchID, batch := range trainData.train_batches {
				wg.Add(trainData.batch_size)
				for sampleID, sample := range batch {
					go shelf.ThreadRun(&mimArray[sampleID], sample, trainData.train_label_batches[batchID][sampleID], &wg)

					if totalSamples%trainData.info_milestone == 0 {
						shelf.Test(mim, trainData.TestData, trainData.TestDataLabel)

						if trainData.save_at_milestone {
							encode_to_json(shelf)
						}
					}
					totalSamples++
				}
				wg.Wait()
				shelf.apply_gradients(trainData.batch_size)

			}
			trainData.shuffle_batches()
		}
		wg.Wait()
	} else {
		for running {
			for batchID, batch := range trainData.train_batches {
				for sampleID, sample := range batch {
					shelf.forward(mim, sample)

					shelf.backprop(mim, trainData.train_label_batches[batchID][sampleID])
					if totalSamples%trainData.info_milestone == 0 {
						shelf.Test(mim, trainData.TestData, trainData.TestDataLabel)
						if trainData.save_at_milestone {
							encode_to_json(shelf)
						}
					}
					totalSamples++
				}
				shelf.apply_gradients(trainData.batch_size)
			}
			trainData.shuffle_batches()
		}
	}

}

func (shelf *Network) ThreadRun(mim *MiM, sample []float64, label []float64, wg *sync.WaitGroup) {
	defer wg.Done()
	shelf.forward(mim, sample)
	shelf.backprop(mim, label)

}

func (shelf *Network) Test(mim *MiM, TestData [][]float64, TestLabels [][]float64) float64 {
	totalCost := 0.0
	correct_choises := 0

	wrongMap := make([]int, 10)

	for sampleID, sample := range TestData {
		shelf.forward(mim, sample)
		totalCost += shelf.cost_interface.call(*mim.data_flat, TestLabels[sampleID])

		if isCorrect(*mim.data_flat, TestLabels[sampleID]) {
			correct_choises += 1
		} else {
			for id, v := range TestLabels[sampleID] {
				if v == 1 {
					wrongMap[id] += 1
					break
				}
			}
		}

	}
	fmt.Println("Percantage Correct on Test Data: ", float64(correct_choises)/float64(len(TestLabels)))
	fmt.Println("Cost: ", totalCost/float64(len(TestLabels)))
	fmt.Println("Wrong guess: ", wrongMap)
	//shelf.layers[0].debug_print()

	return totalCost / float64(len(TestLabels))
}

func isCorrect(values []float64, target_values []float64) bool {
	higT := 0
	for id, v := range target_values {
		if v == 1 {
			higT = id
			break
		}
	}
	highV := 0.0
	highVID := 0
	for id, v := range values {
		if v > highV {
			highVID = id
			highV = v
		}
	}
	return higT == highVID
}

// func (shelf *trainer) shuffle_batches() {
// 	rng := rand.New(rand.NewSource(time.Now().Unix()))
// 	for batch_index := range shelf.train_batches {
// 		rand_index := rng.Int() % len(shelf.train_batches)

//			shelf.train_batches[batch_index], shelf.train_batches[rand_index] = shelf.train_batches[rand_index], shelf.train_batches[batch_index]
//			shelf.train_label_batches[batch_index], shelf.train_label_batches[rand_index] = shelf.train_label_batches[rand_index], shelf.train_label_batches[batch_index]
//		}
//	}
func (shelf *trainer) shuffle_batches() {
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	for batch_index := range shelf.TrainData {
		rand_index := rng.Int() % len(shelf.TrainData)

		shelf.TrainData[batch_index], shelf.TrainData[rand_index] = shelf.TrainData[rand_index], shelf.TrainData[batch_index]
		shelf.TrainDataLabel[batch_index], shelf.TrainDataLabel[rand_index] = shelf.TrainDataLabel[rand_index], shelf.TrainDataLabel[batch_index]

		shelf.TrainDataOrignal[batch_index], shelf.TrainDataOrignal[rand_index] = shelf.TrainDataOrignal[rand_index], shelf.TrainDataOrignal[batch_index]
	}

	// for i := 0; i < int(float64(len(shelf.TrainData))*0.05); i++ {
	// 	rand_index := rng.Int() % len(shelf.TrainData)

	// 	copy(shelf.TrainData[rand_index], shelf.TrainDataOrignal[rand_index])

	// 	data_handler.NoiseInplace(&shelf.TrainData[rand_index])
	// }

}
