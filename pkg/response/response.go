package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	Total    int64 `json:"total"`
	LastPage int   `json:"last_page"`
}

func New(success bool, data interface{}, err string, meta *Meta) *Response {
	return &Response{
		Success: success,
		Data:    data,
		Error:   err,
		Meta:    meta,
	}
}

func Success(data interface{}, meta *Meta) *Response {
	return New(true, data, "", meta)
}

func Error(err string) *Response {
	return New(false, nil, err, nil)
}

func JSON(w http.ResponseWriter, status int, response *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func SuccessJSON(w http.ResponseWriter, data interface{}, meta *Meta) {
	JSON(w, http.StatusOK, Success(data, meta))
}
		
func ErrorJSON(w http.ResponseWriter, status int, err string) {
	JSON(w, status, Error(err))
}
