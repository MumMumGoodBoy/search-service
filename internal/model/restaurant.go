package model

type RestaurantIndex struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	ImageUrl string `json:"imageUrl"`
}

func (i *RestaurantIndex) GetIndexName() string {
	return "restaurants"
}
