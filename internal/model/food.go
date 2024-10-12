package model

type FoodIndex struct {
	ID          string  `json:"id"`
	Restaurant  string  `json:"restaurant"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	ImageUrl    string  `json:"imageUrl"`
}

func (i *FoodIndex) GetIndexName() string {
	return "foods"
}
