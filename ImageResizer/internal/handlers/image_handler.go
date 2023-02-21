package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImageHandler is a struct that contains a reference to the ImageService.
type ImageHandler struct {
	ImageService ImageService
}

// NewImageHandler is a constructor function for ImageHandler.
func NewImageHandler(imageService ImageService) *ImageHandler {
	return &ImageHandler{
		ImageService: imageService,
	}
}

// UploadImage is an HTTP handler function that accepts an image file,
// uploads it to S3, and returns the ID of the uploaded image.
func (h *ImageHandler) UploadImage(c *gin.Context) {
	// Get the uploaded image file from the form data.
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image file"})
		return
	}

	// Upload the image to S3.
	imageID, err := h.ImageService.UploadImage(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
		return
	}

	// Return the ID of the uploaded image.
	c.JSON(http.StatusOK, gin.H{"image_id": imageID})
}

// GetImage is an HTTP handler function that accepts an image ID,
// retrieves the image from S3, and returns the image data.
func (h *ImageHandler) GetImage(c *gin.Context) {
	// Get the image ID from the URL path parameter.
	imageID := c.Param("image_id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image ID"})
		return
	}

	// Retrieve the image from S3.
	image, err := h.ImageService.GetImage(imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	// Return the image data in the HTTP response.
	c.Data(http.StatusOK, image.ContentType, image.Data)
}
