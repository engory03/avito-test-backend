package shareddto

// APIErrorDetail matches components.schemas.ErrorResponse / InternalErrorResponse (.error).
type APIErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIErrorResponse is the standard error envelope from api.yaml.
type APIErrorResponse struct {
	Error APIErrorDetail `json:"error"`
}
