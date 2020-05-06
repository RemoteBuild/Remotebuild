package models

const (
	// NotFoundError error from server
	NotFoundError string = "Not found"
	// ActionNotAllowed error from server
	ActionNotAllowed string = "Action not allowed"
	// WrongLength error from server
	WrongLength string = "Wrong length"
	// ServerError error from server
	ServerError string = "Server Error"
	// WrongInputFormatError wrong user input
	WrongInputFormatError string = "Wrong inputFormat!"
	// InvalidTokenError token is not valid
	InvalidTokenError string = "Token not valid"
	// InvalidCallbackURL token is not valid
	InvalidCallbackURL string = "Callback url is invalid"
	// BatchSizeTooLarge batch is too large
	BatchSizeTooLarge string = "BatchSize soo large!"
	// WrongIntegerFormat integer is probably no integer
	WrongIntegerFormat string = "Number is string"
	// MultipleSourceNameErr err name already exists
	MultipleSourceNameErr string = "You can't have multiple sources with the same name"
	// UserIsInvalidErr err if user is invalid
	UserIsInvalidErr string = "user is invalid"
)

// ResponseStatus the status of response
type ResponseStatus uint8

const (
	//ResponseError if there was an error
	ResponseError ResponseStatus = 0
	//ResponseSuccess if the response is successful
	ResponseSuccess ResponseStatus = 1
)

const (
	// HeaderStatus headername for status in response
	HeaderStatus string = "X-Response-Status"

	// HeaderStatusMessage headername for status in response
	HeaderStatusMessage string = "X-Response-Message"

	// HeaderContentType contenttype of response
	HeaderContentType string = "Content-Type"

	// HeaderContentLength request content length
	HeaderContentLength string = "ContentLength"
)

// StringResponse response containing only one string
type StringResponse struct {
	String string `json:"content"`
}

// StringSliceResponse response containing only one string slice
type StringSliceResponse struct {
	Slice []string `json:"slice"`
}

// LoginResponse response for login
type LoginResponse struct {
	Token string `json:"token"`
}

// CountResponse response containing a count of changed items
type CountResponse struct {
	Count uint32 `json:"count"`
}

// IDsResponse response containing a list of ids
type IDsResponse struct {
	IDs []uint `json:"ids"`
}

// UserAttributeDataResponse response for userattribute data
type UserAttributeDataResponse struct {
	Namespace []Namespaceinfo `json:"nsData"`
}

// Namespaceinfo info for namespace
type Namespaceinfo struct {
	Name   string   `json:"ns"`
	Groups []string `json:"groups"`
}
