package responder

import (
	"encoding/json"
	"log"
	"net/http"
)

type Responder interface {
	ErrorBadRequest(w http.ResponseWriter, err error)
	ErrorInternal(w http.ResponseWriter, err error)
	ErrorNotFound(w http.ResponseWriter)
	OutputJSON(w http.ResponseWriter, responseData interface{})
}

type ApiResponce struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Respond struct{}

func NewResponder() Responder {
	return &Respond{}
}

func (r *Respond) ErrorBadRequest(w http.ResponseWriter, err error) {
	ApiResponce := ApiResponce{
		Code:    http.StatusBadRequest,
		Type:    "Bad request",
		Message: err.Error(),
	}
	w.WriteHeader(ApiResponce.Code)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ApiResponce)
	if err != nil {
		log.Println("encode json error: ", err)
	}
}

func (r *Respond) ErrorInternal(w http.ResponseWriter, err error) {
	ApiResponce := ApiResponce{
		Code:    http.StatusInternalServerError,
		Type:    "Internal error",
		Message: err.Error(),
	}
	w.WriteHeader(ApiResponce.Code)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ApiResponce)
	if err != nil {
		log.Println("encode json error: ", err)
	}
}

func (r *Respond) ErrorNotFound(w http.ResponseWriter) {
	ApiResponce := ApiResponce{
		Code:    http.StatusNotFound,
		Type:    "Not found",
		Message: "Object not found",
	}
	w.WriteHeader(ApiResponce.Code)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(ApiResponce)
	if err != nil {
		log.Println("encode json error: ", err)
	}
}

func (r *Respond) OutputJSON(w http.ResponseWriter, responseData interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(responseData)
	if err != nil {
		log.Println("encode json error: ", err)
	}
}