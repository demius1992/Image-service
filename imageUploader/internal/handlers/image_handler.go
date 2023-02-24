package handlers

import (
	"bytes"
	"github.com/demius1992/Image-service/imageUploader/internal/interfaces"
	"github.com/demius1992/Image-service/imageUploader/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
)

const MAX_UPLOAD_SIZE = 5 << 30

type IDs struct {
	ID1 string `json:"id1"`
	ID2 string `json:"id2"`
	ID3 string `json:"id3"`
}

// ImageHandle handles the image-related endpoints.
type ImageHandle struct {
	imageService interfaces.ImageHandler
}

// NewImageHandler creates a new ImageHandle instance.
func NewImageHandler(imageHandler interfaces.ImageHandler) *ImageHandle {
	return &ImageHandle{
		imageService: imageHandler,
	}
}

// UploadImage handles the image upload endpoint.
func (h *ImageHandle) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MAX_UPLOAD_SIZE)

	// Get the image file from the request
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		logrus.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get the image file from the request"})
		return
	}
	defer file.Close()

	buffer := make([]byte, header.Size)
	_, _ = file.Read(buffer)
	contentType := http.DetectContentType(buffer)

	// Upload the image to S3 and publish a message to Kafka
	image, err := h.imageService.UploadImage(file, contentType)
	if err != nil {
		logrus.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload the image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        image.ID,
		"createdAt": image.CreatedAt,
		"url":       image.URL,
	})
}

// GetImage handles the image retrieval endpoint.
func (h *ImageHandle) GetImage(c *gin.Context) {
	// Get the image ID from the request URL
	id := c.Param("id")

	fromBytes, err := uuid.FromBytes([]byte(id))
	if err != nil {
		logrus.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}
	// Retrieve the image from S3
	image, err := h.imageService.GetImage(fromBytes)
	if err != nil {
		logrus.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve the image"})
		return
	}

	// Return the image as the response
	c.DataFromReader(http.StatusOK, image.Size, image.ContentType, bytes.NewReader(image.Content), make(map[string]string))
}

// GetImageVariants handles the image variant retrieval endpoint.
func (h *ImageHandle) GetImageVariants(c *gin.Context) {
	// Get the image URLs from the request body
	var data IDs

	if err := c.BindJSON(&data); err != nil {
		logrus.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse image URLs from the request body"})
		return
	}

	ids := []string{data.ID1, data.ID2, data.ID3}

	// Retrieve the image variants from S3
	imageVariants, err := h.imageService.GetImageVariants(ids)
	if err != nil {
		logrus.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve the image variants"})
		return
	}

	// Map the image variants to a slice of ImageVariantResponse models
	imageVariantResponses := make([]models.Image, len(imageVariants))
	for i, imageVariant := range imageVariants {
		imageVariantResponses[i] = models.Image{
			Name: imageVariant.Name,
			URL:  imageVariant.URL,
		}
	}

	// Return the image variants as the response
	c.JSON(http.StatusOK, imageVariantResponses)
}