package models

import (
	"math"
)

type Vector struct {
	ID     int64
	Name   string
	Vector []float32
}

// CalculateDistance calculates the Euclidean distance between two vectors
func CalculateDistance(v1, v2 []float32) float32 {
	var sum float32
	for i := range v1 {
		diff := v1[i] - v2[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}
