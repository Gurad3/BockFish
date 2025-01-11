package neural

import (
	"math"
)

type Activation interface {
	call(float64) float64
	ddx(float64) float64
	get_name() string
}

func act_name_to_interface(name string) Activation {
	switch name {
	case "relu":
		return &RelU{}
	case "sigmoid":
		return &Sigmoid{}
	case "tanh":
		return &TanH{}
	case "SiLU":
		return &SiLU{}
	}
	return &RelU{}
}

type RelU struct {
}

func (shelf *RelU) get_name() string {
	return "relu"
}

func (shelf *RelU) call(val float64) float64 {
	return max(0, val)
}

func (shelf *RelU) ddx(val float64) float64 {
	if val > 0 {
		return 1
	} else {
		return 0
	}
}

type Sigmoid struct{}

func (shelf *Sigmoid) get_name() string {
	return "sigmoid"
}
func (shelf *Sigmoid) call(val float64) float64 {
	return 1 / (1 + math.Exp(-val))
}

func (shelf *Sigmoid) ddx(val float64) float64 {
	sig := shelf.call(val)
	return sig * (1 - sig)
}

type TanH struct{}

func (shelf *TanH) get_name() string {
	return "tanh"
}

func (shelf *TanH) call(val float64) float64 {
	posExp := math.Exp(val)
	negExp := math.Exp(-val)
	return (posExp - negExp) / (posExp + negExp)
}

func (shelf *TanH) ddx(val float64) float64 {

	return 1 - math.Pow(shelf.call(val), 2)
}

type SiLU struct{}

func (shelf *SiLU) get_name() string {
	return "SiLU"
}

func (shelf *SiLU) call(val float64) float64 {
	return val / (1 + math.Exp(-val))
}

func (shelf *SiLU) ddx(val float64) float64 {
	negExp := math.Exp(-val)
	return (1 + negExp + val*negExp) / math.Pow((1+negExp), 2)
}
