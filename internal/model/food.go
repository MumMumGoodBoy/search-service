package model

type FoodIndex struct {
	ID          string  `json:"id"`
	Restaurant  string  `json:"restaurant"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	ImageUrl    string  `json:"image_url"`
}

type FoodEvent struct {
	Event        string  `json:"event"`
	Id           string  `json:"id"`
	FoodName     string  `json:"foodName"`
	RestaurantId string  `json:"restaurantId"`
	Price        float32 `json:"price"`
	Description  string  `json:"description"`
	ImageUrl     string  `json:"image_url"`
}
