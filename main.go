package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MumMumGoodBoy/search-service/internal/route"
	"github.com/MumMumGoodBoy/search-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rabbitmq/amqp091-go"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	app := fiber.New()

	client := meilisearch.New(os.Getenv("MEILISEARCH_URL"), meilisearch.WithAPIKey(os.Getenv("MEILISEARCH_MASTER_KEY")))

	rabbitMQConn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMQConn.Close()

	rabbitMQChannel, err := rabbitMQConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMQChannel.Close()


	searchService := service.SearchService{
		Client:          client,
		RabbitMQChannel: rabbitMQChannel,
	}

	// searchService.InitIndexWithDocuments("restaurants", "data/restaurants.json")
	// searchService.InitIndexWithDocuments("foods", "data/foods.json")
	searchService.SetUpRestaurantConsumer()
	searchService.SetUpFoodConsumer()
	route.CreateSearchRoute(app, &searchService)

	fmt.Println("Server is running on port 8080")
	app.Listen(":8080")
}
