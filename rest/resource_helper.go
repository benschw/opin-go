package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Helpers
func PathString(req *http.Request, key string) (string, error) {
	return mux.Vars(req)[key], nil
}
func PathInt(req *http.Request, key string) (int, error) {
	str, err := PathString(req, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)

}

func Bind(req *http.Request, entity interface{}) error {
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(entity)
}

// Response Factories
func SetConflictResponse(res http.ResponseWriter) {
	JsonResponse(res)
	res.WriteHeader(http.StatusConflict)

}
func SetBadRequestResponse(res http.ResponseWriter) {
	JsonResponse(res)
	res.WriteHeader(http.StatusBadRequest)

}
func SetNotFoundResponse(res http.ResponseWriter) {
	JsonResponse(res)
	res.WriteHeader(http.StatusNotFound)
}

func SetInternalServerErrorResponse(res http.ResponseWriter, err error) {
	log.Print(err)
	JsonResponse(res)
	res.WriteHeader(http.StatusInternalServerError)
}

func SetCreatedResponse(res http.ResponseWriter, entity interface{}, location string) error {
	b, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	body := string(b[:])

	JsonResponse(res)
	res.Header().Set("Location", location)
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, body)
	return nil
}

func SetOKResponse(res http.ResponseWriter, entity interface{}) error {
	if entity != nil {
		b, err := json.Marshal(entity)
		if err != nil {
			return err
		}
		body := string(b[:])

		JsonResponse(res)
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, body)
	} else {
		res.WriteHeader(http.StatusOK)
	}
	return nil
}

func SetNoContentResponse(res http.ResponseWriter) error {
	JsonResponse(res)
	res.WriteHeader(http.StatusNoContent)
	return nil
}

func JsonResponse(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
}
