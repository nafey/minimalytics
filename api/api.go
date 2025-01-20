package api

import (
	"bytes"
	"log"
	"io"
	"net/http"
	"encoding/json"
	"minimalytics/model"
)

type Message struct {
	Event string
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type StatRequest struct {
	Event string `json:"event"`
}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func HandleEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var t Message
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	event := t.Event

	model.GetMinuteEvent(event)
	model.GetHourlyEvent(event)
	model.GetDailyEvent(event)

	io.WriteString(w, "OK")
}


func HandleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var statRequest StatRequest
	status := "OK"
	message := "Request received"

	var data any = nil

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Unable to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if string(body) == "" {
		var response Response
		response = Response{
			Status:  "OK",
			Message: "Request received",
			Data:    nil,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}

		return
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&statRequest)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error decoding request", http.StatusBadRequest)
	}

	if r.URL.Path == "/api/stat/daily/" {
		message = "Daily stat"
		data = model.GetDailyStat(statRequest.Event)

	} else if r.URL.Path == "/api/stat/hourly/" {
		message = "Hourly stat"
		data = model.GetHourlyStat(statRequest.Event)

	} else if r.URL.Path == "/api/stat/minutes/" {
		message = "Minute stat"
		data = model.GetMinuteStat(statRequest.Event)

	} else {
		message = "Not implemented"

	}

	var response Response = Response{
		Status: status,
		Message: message,
		Data: data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}