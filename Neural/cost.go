package neural

import (
	"math"
	"math/rand"
)

type Cost interface {
	call(out_values []float64, target_values []float64) float64

	ddx(out_val float64, target_val float64) float64
	get_name() string
}

func cost_name_to_interface(name string) Cost {
	switch name {
	case "meansquare":
		return &MeanSquare{}

	case "CrossEntropy":
		return &CrossEntropy{}

	}
	return &MeanSquare{}
}

type MeanSquare struct {
}

func (shelf *MeanSquare) call(out_values []float64, target_values []float64) float64 {
	sum := 0.0
	for i := range out_values {
		sum += (out_values[i] - target_values[i]) * (out_values[i] - target_values[i])
	}

	return sum
}

func (shelf *MeanSquare) ddx(out_val float64, target_val float64) float64 {
	return (out_val - target_val) * 2
}
func (shelf *MeanSquare) get_name() string {
	return "meansquare"
}

type CrossEntropy struct {
}

func (shelf *CrossEntropy) call(out_values []float64, target_values []float64) float64 {
	sum := 0.0
	for i := range out_values {
		sum -= target_values[i] * math.Log10(out_values[i])
	}
	return sum
}

func (shelf *CrossEntropy) ddx(out_val float64, target_val float64) float64 {
	return -target_val / (out_val * math.Log(10))
}

func (shelf *CrossEntropy) get_name() string {
	return "CrossEntropy"
}

func initWeightXavierUniform(xavierRange float64, r rand.Rand) float64 {
	return r.Float64()*2*xavierRange - xavierRange
}
