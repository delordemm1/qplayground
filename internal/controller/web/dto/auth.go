package dto

// SendOtpRequest is the DTO for the send OTP request.
// It includes validation tags that define the rules for each field.
type SendOtpRequest struct {
	// `json:"email"` is for JSON decoding.
	// `validate:"required,email"` tells the validator:
	// - `required`: This field must not be empty.
	// - `email`: This field must be a valid email address format.
	Email string `json:"email" validate:"required,email"`
}

// You can define more DTOs here, like for the actual OTP login.
type LoginWithOtpRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
}
