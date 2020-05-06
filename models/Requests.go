package models

// PingRequest ping request
type PingRequest struct {
	Payload string
}

// CredentialsRequest request containing credentials
type CredentialsRequest struct {
	MachineID string `json:"mid,omitempty"`
	Username  string `json:"username"`
	Password  string `json:"pass"`
}

// UserAttributesRequest request for getting
// namespaces and groups
type UserAttributesRequest struct {
	Mode uint `json:"m"`
}

//UploadType type of upload
type UploadType uint8

//Available upload types
const (
	FileUploadType UploadType = iota
	URLUploadType
)
