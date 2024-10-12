package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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
