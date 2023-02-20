package handlers

import (
	"net/http"

	"github.com/demius1992/Image-service/internal/models"
	"github.com/demius1992/Image-service/internal/services"
	"github.com/gin-gonic/gin"
)

// ImageHandler handles the image-related endpoints.
type ImageHandler struct {
	imageService services.ImageService
}

// NewImageHandler creates a new ImageHandler instance.
func NewImageHandler(imageService services.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

// UploadImage handles the image upload endpoint.
func (h *ImageHandler) UploadImage(c *gin.Context) {
	// Get the image file from the request
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get the image file from the request"})
		return
	}

	// Open the file
	imageFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open the image file"})
		return
	}
	defer imageFile.Close()

	// Get the content type of the image
	contentType := file.Header.Get("Content-Type")

	// Upload the image to S3 and publish a message to Kafka
	image, err := h.imageService.UploadImage(imageFile, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload the image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        image.ID,
		"createdAt": image.CreatedAt,
	})
}

// GetImage handles the image retrieval endpoint.
func (h *ImageHandler) GetImage(c *gin.Context) {
	// Get the image ID from the request URL
	id := c.Param("id")

	// Retrieve the image from S3
	image, err := h.imageService.GetImage(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve the image"})
		return
	}

	// Return the image as the response
	c.DataFromReader(http.StatusOK, image.ContentType, image.Size, image.Content)
}

// GetImageVariants handles the image variant retrieval endpoint.
func (h *ImageHandler) GetImageVariants(c *gin.Context) {
	// Get the image ID from the request URL
	id := c.Param("id")

	// Retrieve the image variants from S3
	imageVariants, err := h.imageService.GetImageVariants(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve the image variants"})
		return
	}

	// Map the image variants to a slice of ImageVariantResponse models
	imageVariantResponses := make([]models.Image, len(imageVariants))
	for i, imageVariant := range imageVariants {
		imageVariantResponses[i] = models.Image{
			Title:    imageVariant.Name,
			ImageURL: imageVariant.URL,
		}
	}

	// Return the image variants as the response
	c.JSON(http.StatusOK, imageVariantResponses)
}