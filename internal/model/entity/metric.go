package dto

type MetricEntity struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
}
