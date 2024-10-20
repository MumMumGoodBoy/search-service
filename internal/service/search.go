package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/MumMumGoodBoy/search-service/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rabbitmq/amqp091-go"
)

type SearchService struct {
	Client          meilisearch.ServiceManager
	RabbitMQChannel *amqp091.Channel
}

func (s *SearchService) InitIndexWithDocuments(indexName, filePath string) error {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var documents []map[string]interface{}
	if err := json.Unmarshal(byteValue, &documents); err != nil {
		return fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	_, err = s.Client.Index(indexName).AddDocuments(documents)
	if err != nil {
		return fmt.Errorf("failed to add documents to MeiliSearch index %s: %w", indexName, err)
	}

	return nil
}

func (s *SearchService) SearchRestaurant(c *fiber.Ctx) error {
	search := c.Query("search")
	offset := c.QueryInt("offset")
	limit := c.QueryInt("limit")

	searchResult, err := s.Client.Index("restaurants").Search(search, &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(200).JSON(searchResult)
}

func (s *SearchService) SearchFood(c *fiber.Ctx) error {
	search := c.Query("search")
	offset := c.QueryInt("offset")
	limit := c.QueryInt("limit")

	searchResult, err := s.Client.Index("foods").Search(search, &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(200).JSON(searchResult)
}

func (s *SearchService) SetUpRestaurantConsumer() error {
	q, err := s.RabbitMQChannel.QueueDeclare(
		"restaurant_search_queue", // Empty name creates a random queue
		false,                     // Not durable
		false,                     // Auto-delete when not used
		true,                      // Exclusive (only this connection can use it)
		false,                     // No-wait
		nil,                       // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = s.RabbitMQChannel.QueueBind(
		q.Name,             // Queue name
		"restaurant.*",     // Routing key
		"restaurant_topic", // Exchange name
		false,              // No-wait
		nil,                // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	messages, err := s.RabbitMQChannel.Consume(
		q.Name, // Queue name
		"",     // Consumer tag
		true,   // Auto-Ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Arguments
	)

	if err != nil {
		return fmt.Errorf("failed to consume msg: %w", err)
	}

	go func() {
		for d := range messages {
			var restaurant model.RestaurantEvent
			err := json.Unmarshal(d.Body, &restaurant)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
			}
			event := strings.Split(string(restaurant.Event), ".")[1]
			switch event {
			case "create":
				s.insertRestaurantocument(restaurant)
			case "update":
				s.updateRestaurantDocument(restaurant)
			case "delete":
				s.deleteRestaurantDocument(restaurant)
			default:
				log.Printf("Error unsupported event: %v", event)
			}

		}
	}()
	return nil
}

func (s *SearchService) SetUpFoodConsumer() error {
	q, err := s.RabbitMQChannel.QueueDeclare(
		"food_search_queue", // Empty name creates a random queue
		false,               // Not durable
		false,               // Auto-delete when not used
		true,                // Exclusive (only this connection can use it)
		false,               // No-wait
		nil,                 // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = s.RabbitMQChannel.QueueBind(
		q.Name,       // Queue name
		"food.*",     // Routing key
		"food_topic", // Exchange name
		false,        // No-wait
		nil,          // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	messages, err := s.RabbitMQChannel.Consume(
		q.Name, // Queue name
		"",     // Consumer tag
		true,   // Auto-Ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Arguments
	)

	if err != nil {
		return fmt.Errorf("failed to consume msg: %w", err)
	}

	go func() {
		for d := range messages {
			var food model.FoodEvent
			err := json.Unmarshal(d.Body, &food)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
			}
			event := strings.Split(string(food.Event), ".")[1]
			switch event {
			case "create":
				s.insertFoodDocument(food)
			case "update":
				s.updateFoodDocument(food)
			case "delete":
				s.deleteFoodDocument(food)
			default:
				log.Printf("Error unsupported event: %v", event)
			}

		}
	}()
	return nil
}

func (s *SearchService) insertFoodDocument(newFood model.FoodEvent) {
	documents := model.FoodIndex{
		ID:          newFood.Id,
		Name:        newFood.FoodName,
		Restaurant:  newFood.RestaurantId,
		Description: newFood.Description,
		Price:       newFood.Price,
		ImageUrl:    newFood.ImageUrl,
	}
	s.Client.Index("foods").AddDocuments(documents)
}

func (s *SearchService) updateFoodDocument(newFood model.FoodEvent) {
	addIfNotEmpty := func(m map[string]interface{}, key string, value interface{}) {
		if value != nil {
			m[key] = value
		}
	}
	documents := map[string]interface{}{}

	addIfNotEmpty(documents, "id", newFood.Id)
	addIfNotEmpty(documents, "name", newFood.FoodName)
	addIfNotEmpty(documents, "restaurant", newFood.RestaurantId)
	addIfNotEmpty(documents, "description", newFood.Description)
	addIfNotEmpty(documents, "price", newFood.Price)
	addIfNotEmpty(documents, "imageUrl", newFood.ImageUrl)

	s.Client.Index("foods").UpdateDocuments([]map[string]interface{}{documents})
}

func (s *SearchService) deleteFoodDocument(food model.FoodEvent) {
	s.Client.Index("foods").DeleteDocument(food.Id)
}

func (s *SearchService) insertRestaurantocument(newRestaurant model.RestaurantEvent) {
	documents := model.RestaurantIndex{
		ID:      newRestaurant.Id,
		Name:    newRestaurant.RestaurantName,
		Address: newRestaurant.Address,
		Phone:   newRestaurant.Phone,
	}
	s.Client.Index("restaurants").AddDocuments(documents)
}

func (s *SearchService) updateRestaurantDocument(newRestaurant model.RestaurantEvent) {
	addIfNotEmpty := func(m map[string]interface{}, key string, value interface{}) {
		if value != nil {
			m[key] = value
		}
	}
	documents := map[string]interface{}{}

	addIfNotEmpty(documents, "id", newRestaurant.Id)
	addIfNotEmpty(documents, "name", newRestaurant.RestaurantName)
	addIfNotEmpty(documents, "address", newRestaurant.Address)
	addIfNotEmpty(documents, "phone", newRestaurant.Phone)

	s.Client.Index("restaurants").UpdateDocuments([]map[string]interface{}{documents})
}

func (s *SearchService) deleteRestaurantDocument(restaurant model.RestaurantEvent) {
	s.Client.Index("restaurants").DeleteDocument(restaurant.Id)
}
