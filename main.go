package main

import (
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

	port := os.Getenv("PORT")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	app := fiber.New()

	client := meilisearch.New(os.Getenv("MEILISEARCH_URL"), meilisearch.WithAPIKey(os.Getenv("MEILISEARCH_MASTER_KEY")))
	log.Println("Connected to MeiliSearch")

	rabbitMQConn, err := amqp091.Dial(rabbitMQURL)
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
	log.Println("Connected to RabbitMQ")

	// searchService.InitIndexWithDocuments("restaurants", "data/restaurants.json")
	// searchService.InitIndexWithDocuments("foods", "data/foods.json")
	err = searchService.SetUpRestaurantConsumer()
	if err != nil {
		log.Fatal(err)
	}
	err = searchService.SetUpFoodConsumer()
	if err != nil {
		log.Fatal(err)
	}

	route.CreateSearchRoute(app, &searchService)

	log.Println("Search service is running on port", port)
	app.Listen(":" + port)
}
