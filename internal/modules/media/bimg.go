package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/h2non/bimg"
)

type BimgProcessor struct{}

func NewBimgProcessor() ImageProcessor {
	return &BimgProcessor{}
}

func (p *BimgProcessor) Process(ctx context.Context, input io.Reader, options *ImageProcessOptions) (*ProcessedImage, error) {
	// Read input data
	buffer, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Create bimg image
	image := bimg.NewImage(buffer)

	// Get original size for validation
	size, err := image.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}

	slog.Debug("Processing image", "originalWidth", size.Width, "originalHeight", size.Height)

	// Build processing options
	bimgOptions := bimg.Options{
		Quality: options.Quality,
	}

	// Set format
	switch options.Format {
	case JPEG:
		bimgOptions.Type = bimg.JPEG
	case PNG:
		bimgOptions.Type = bimg.PNG
	case WEBP:
		bimgOptions.Type = bimg.WEBP
	default:
		bimgOptions.Type = bimg.JPEG
	}

	// Handle resizing/cropping
	if options.Width > 0 || options.Height > 0 {
		bimgOptions.Width = options.Width
		bimgOptions.Height = options.Height

		if options.Crop {
			bimgOptions.Crop = true
			bimgOptions.Gravity = bimg.GravityCentre
		}
	}

	// Process the image
	processedBuffer, err := image.Process(bimgOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Get processed image info
	processedImage := bimg.NewImage(processedBuffer)
	processedSize, err := processedImage.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get processed image size: %w", err)
	}

	// Determine content type
	contentType := "image/jpeg"
	switch options.Format {
	case PNG:
		contentType = "image/png"
	case WEBP:
		contentType = "image/webp"
	}

	result := &ProcessedImage{
		Data:        processedBuffer,
		ContentType: contentType,
		Width:       processedSize.Width,
		Height:      processedSize.Height,
		Size:        int64(len(processedBuffer)),
	}

	slog.Info("Image processed successfully",
		"originalSize", fmt.Sprintf("%dx%d", size.Width, size.Height),
		"processedSize", fmt.Sprintf("%dx%d", result.Width, result.Height),
		"format", contentType,
		"sizeBytes", result.Size,
	)

	return result, nil
}

func (p *BimgProcessor) GetImageInfo(ctx context.Context, input io.Reader) (*ImageInfo, error) {
	buffer, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	image := bimg.NewImage(buffer)
	size, err := image.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}

	metadata, err := image.Metadata()
	if err != nil {
		return nil, fmt.Errorf("failed to get image metadata: %w", err)
	}

	// Determine content type from format
	contentType := "image/jpeg"
	format := "JPEG"
	switch metadata.Type {
	case "png":
		contentType = "image/png"
		format = "PNG"
	case "webp":
		contentType = "image/webp"
		format = "WEBP"
	case "gif":
		contentType = "image/gif"
		format = "GIF"
	}

	return &ImageInfo{
		Width:       size.Width,
		Height:      size.Height,
		Format:      format,
		Size:        int64(len(buffer)),
		ContentType: contentType,
	}, nil
}