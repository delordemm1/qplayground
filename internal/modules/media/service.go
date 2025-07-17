package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type mediaService struct {
	processor ImageProcessor
}

func NewMediaService(processor ImageProcessor) MediaService {
	return &mediaService{
		processor: processor,
	}
}

func (s *mediaService) ProcessImage(ctx context.Context, input io.Reader, options *ImageProcessOptions) (*ProcessedImage, error) {
	// Set default options if not provided
	if options == nil {
		options = &ImageProcessOptions{
			Quality: 85,
			Format:  JPEG,
		}
	}

	// Set default quality if not specified
	if options.Quality == 0 {
		options.Quality = 85
	}

	result, err := s.processor.Process(ctx, input, options)
	if err != nil {
		slog.Error("Failed to process image", "error", err)
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	slog.Info("Image processed successfully", 
		"width", result.Width, 
		"height", result.Height, 
		"size", result.Size,
		"contentType", result.ContentType,
	)

	return result, nil
}

func (s *mediaService) ValidateImage(ctx context.Context, input io.Reader) (*ImageInfo, error) {
	info, err := s.processor.GetImageInfo(ctx, input)
	if err != nil {
		slog.Error("Failed to validate image", "error", err)
		return nil, fmt.Errorf("failed to validate image: %w", err)
	}

	// Basic validation
	if info.Width == 0 || info.Height == 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", info.Width, info.Height)
	}

	// Check for reasonable size limits (e.g., max 50MB)
	maxSize := int64(50 * 1024 * 1024) // 50MB
	if info.Size > maxSize {
		return nil, fmt.Errorf("image too large: %d bytes (max %d bytes)", info.Size, maxSize)
	}

	slog.Info("Image validated successfully", 
		"width", info.Width, 
		"height", info.Height, 
		"format", info.Format,
		"size", info.Size,
	)

	return info, nil
}