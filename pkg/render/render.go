package render

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func BindJSON(r *http.Request, v any) error {
	const op = "render.BindJSON"

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return fmt.Errorf("%s: failed to decode request body: %w", op, err)
	}

	return nil
}

func JSON(w http.ResponseWriter, statusCode int, v any) error {
	const op = "render.JSON"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("%s: failed to encode v: %w", op, err)
	}

	return nil
}
