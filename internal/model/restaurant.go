package model

type RestaurantIndex struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	ImageUrl string `json:"imageUrl"`
}

type RestaurantEvent struct {
	Event          string `json:"event"`
	Id             string `json:"id"`
	RestaurantName string `json:"restaurantName"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
}

