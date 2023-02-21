package handlers

import (
	"log"
	"net/http"

	"github.com/demius1992/Image-service/ImageUploader/internal/interfaces"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	kafkaService interfaces.KafkaService
}

func NewMessageHandler(kafkaService interfaces.KafkaService) *MessageHandler {
	return &MessageHandler{
		kafkaService: kafkaService,
	}
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	// Get the messages from the Kafka service
	messages, err := h.kafkaService.GetMessages(c)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get messages",
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}
