package main

import (
	"encoding/json"
	"net/http"
)

type EndpointHandler func(r *http.Request) (interface{}, error)

func endpoint(handler EndpointHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := handler(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if response == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if model, ok := response.(Model); ok && !model.IsValid() {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		bytes, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(bytes)
	}
}
