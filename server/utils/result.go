package utils

type ErrorResponse struct {
	Msg     string   `json:"message,omitempty"`
	Details []string `json:"details,omitempty"`
}

type Response[T any] struct {
	OK    bool           `json:"ok"`
	Data  T              `json:"data,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

// NewResponse creates a new generic Response instance.
func NewOkResponse[T any](data T) Response[T] {
	return Response[T]{
		OK:   true,
		Data: data,
	}
}

func NewErrorResponse(message string) Response[any] {
	return Response[any]{
		OK: false,
		Error: &ErrorResponse{
			Msg:     message,
			Details: nil,
		},
	}
}

func NewErrorResponseWithDetails(message string, details []string) Response[any] {
	return Response[any]{
		OK: false,
		Error: &ErrorResponse{
			Msg:     message,
			Details: details,
		},
	}
}
