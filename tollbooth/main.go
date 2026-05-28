package main

import (
	"encoding/json"
	"log"
	"net/http"

	tollbooth "github.com/didip/tollbooth/v6"
)

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "success",
		Body:   "This is the response from the endpoint.",
	}
	err := json.NewEncoder(writer).Encode(message)
	if err != nil {
		return
	}
}

func main() {
	message := Message{
		Status: "Request failed",
		Body:   "The API is at capacity, try again later.",
	}
	// Marshal the message struct to JSON format.
	jsonMessage, _ := json.Marshal(message)

	// Create new limiter with a maximum of 1 request per second and no expiration time.
	tlbthLimiter := tollbooth.NewLimiter(1, nil)
	// Set the message content type to application/json and the message to returned when the request is not allowed.
	tlbthLimiter.SetMessageContentType("application/json")
	// Set the message to be returned when the request is not allowed to the JSON message created above.
	tlbthLimiter.SetMessage(string(jsonMessage))

	http.Handle("/ping", tollbooth.LimitFuncHandler(tlbthLimiter, endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error starting server:8080", err)
	}
}
