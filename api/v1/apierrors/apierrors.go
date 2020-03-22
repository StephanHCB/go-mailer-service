package apierrors

// Model for the generic error response.
//
// swagger:model errorDto
type ErrorDto struct {
	// The timestamp at which the error occurred
	Timestamp string `json:"timestamp"`
	// The request id associated with this request
	RequestId string `json:"requestid"`
	// The error code
	Message   string `json:"message"`
	// Additional details
	Details   []string `json:"details"`
}

// this seems necessary to reference a model

// The generic error response.
//
// swagger:response errorResponse
type ErrorResponse struct {
	// The details of the error
	//
	// in:body
	Body ErrorDto
}