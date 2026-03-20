package response

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pulse-backend/internal/models"
)

// JSONResponse writes a JSON body with the given HTTP status code.
func JSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload) //nolint:errcheck
}

// ErrResp creates an error response.
func ErrResp(code, message string, details interface{}, hint string) models.ErrorResponse {
	return models.ErrorResponse{
		Success: false,
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
			Hint:    hint,
		},
	}
}

// APIError logs error and returns error response.
func APIError(r *http.Request, status int, resp models.ErrorResponse) models.ErrorResponse {
	hint := ""
	if resp.Error.Hint != "" {
		hint = fmt.Sprintf(" | hint=%s", resp.Error.Hint)
	}
	log.Printf("[ERROR] %s %s | %d | code=%s | msg=%s%s",
		r.Method, r.URL.Path, status, resp.Error.Code, resp.Error.Message, hint)
	return resp
}
