package media

import (
	"context"
	"io"
)

// ImageFormat represents supported image formats
type ImageFormat int

const (
	JPEG ImageFormat = iota
	PNG
	WEBP
)

// ImageProcessOptions contains options for image processing
type ImageProcessOptions struct {
	Width   int
	Height  int
	Quality int
	Format  ImageFormat
	Crop    bool
}

// ProcessedImage contains the processed image data and metadata
type ProcessedImage struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
	Size        int64
}

// ImageProcessor defines the interface for image processing operations
type ImageProcessor interface {
	Process(ctx context.Context, input io.Reader, options *ImageProcessOptions) (*ProcessedImage, error)
	GetImageInfo(ctx context.Context, input io.Reader) (*ImageInfo, error)
}

// ImageInfo contains metadata about an image
type ImageInfo struct {
	Width       int
	Height      int
	Format      string
	Size        int64
	ContentType string
}

// MediaService provides high-level media processing operations
type MediaService interface {
	ProcessImage(ctx context.Context, input io.Reader, options *ImageProcessOptions) (*ProcessedImage, error)
	ValidateImage(ctx context.Context, input io.Reader) (*ImageInfo, error)
}