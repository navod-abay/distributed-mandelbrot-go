package main

import (
	"errors"

	"github.com/navod-abay/mandelbrotset-go/core/models"
)

const epsilon = 1e-9

func ValidateImageRequest(imageDimensions models.ImageDimensions) error {
	if (imageDimensions.X_high - imageDimensions.X_low) < epsilon {
		return errors.New("Specified image too small")
	}
	if (imageDimensions.Y_high - imageDimensions.Y_low) < epsilon {
		return errors.New("Specified image too small")
	}
	return nil
}
