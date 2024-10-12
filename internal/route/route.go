package route

import (
	"github.com/MumMumGoodBoy/search-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

func CreateSearchRoute(r fiber.Router, searchService *service.SearchService) {
	r.Get("/search/restaurants", searchService.SearchRestaurant)
	r.Get("/search/foods", searchService.SearchFood)
}
