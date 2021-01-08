package http

type standardResponse struct {
	Message string `json:"message"`
}

var (
	internalServerErrorResponse = standardResponse{"An unexpected error occurred"}
	resourceNotFoundResponse    = standardResponse{"The requested resource is not found"}
	badRequestResponse          = standardResponse{"Invalid payload supplied"}
)
