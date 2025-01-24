package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"minimalytics/model"
	"net/http"
	"strings"

	// "github.com/sirupsen/logrus/hooks/writer"
)

type Message struct {
	Event string
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type StatRequest struct {
	Event string `json:"event"`
}

func writeResponse(w http.ResponseWriter, err error, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	var response Response
	var status string = "OK"

	if err != nil {
		status = "ERROR"
		log.Println(message)
		log.Println(err)
	}

	response = Response{
		Status:  status,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
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

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	log.Println(parts)
	log.Println(len(parts))

	if len(parts) < 4 {
		writeResponse(w, nil, "Request received", nil)
		return 
	}

	key := parts[3]
	config := model.GetConfig(key)

	value := config.Value
	writeResponse(w, nil, "Value", value)
}

func HandleEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var t Message
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	event := t.Event

	model.SubmitMinuteEvent(event)
	model.SubmitHourlyEvent(event)
	model.SubmitDailyEvent(event)

	io.WriteString(w, "OK")
}

func HandleStat(w http.ResponseWriter, r *http.Request) {
	var statRequest StatRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		writeResponse(w, err, "Unable to read body", nil)

		return
	}
	defer r.Body.Close()

	if len(string(body)) <= 2 {
		writeResponse(w, errors.New("Inavlid body size"), "No event provided in request", nil)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&statRequest)
	if err != nil {
		writeResponse(w, err, "Invalid Request Body", nil)
	}

	if r.URL.Path == "/api/stat/daily/" {
		writeResponse(w, nil, "Daily Stat", model.GetDailyStat(statRequest.Event))

	} else if r.URL.Path == "/api/stat/hourly/" {
		writeResponse(w, nil, "Hourly Stat", model.GetHourlyStat(statRequest.Event))

	} else if r.URL.Path == "/api/stat/minutes/" {
		writeResponse(w, nil, "Minute Stat", model.GetMinuteStat(statRequest.Event))

	} else {
		writeResponse(w, nil, "Not implemented", nil)

	}

}
