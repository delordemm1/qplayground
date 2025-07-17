package domain

// RequestMetadata holds non-critical information about a request,
// often used for logging, auditing, or analytics.
type RequestMetadata struct {
	IPAddress string
	UserAgent string
}
