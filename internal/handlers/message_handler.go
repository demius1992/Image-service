package handlers

import (
	"log"
	"net/http"

	"github.com/demius1992/Image-service/internal/services"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	kafkaService services.KafkaService
}

func NewMessageHandler(kafkaService services.KafkaService) *MessageHandler {
	return &MessageHandler{
		kafkaService: kafkaService,
	}
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	// Get the messages from the Kafka service
	messages, err := h.kafkaService.GetMessages()
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get messages",
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}
