package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"sync"
)

type Message struct {
	Status string `json:"status"`
	Body string `json:"body"`
}

func perClientRateLimiter(next func(writer http.ResponseWriter, request *http.Request)) http.Handler {

	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}

	// Create a map to store the clients and a mutex to synchronize access to the map.
	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)

	// this goroutine will run in the background and will remove clients that have not been seen for more than 3 minutes.
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the client's IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		mu.Lock()
		// Check if the client exists in the map, if not create a new one and add it to the map.
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(2, 4),
			}
		}

		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()

			message := Message{
				Status: "Request failed",
				Body:   "Too many requests. Please try again later.",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		mu.Unlock()
		// Call the next handler in the chain if the request is allowed.
		next(w, r)
	})

}

// The endpoint handler that will be called when the client makes a request to the endpoint.
func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "success",
		Body:   "This is the response from the endpoint.",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	http.Handle("/ping", perClientRateLimiter(endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("There was an error listening on port :8080", err)
	}
}