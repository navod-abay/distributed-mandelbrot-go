package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	generate "github.com/navod-abay/mandelbrotset-go/core"
)

type CheckRequestResponse struct {
	Num_iterations uint16 `json:"num_iteration"`
}

func CheckInclusionEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handler got called\n")
	query := r.URL.Query()
	var x, y float64
	var err error
	x, err = strconv.ParseFloat(query.Get("X"), 64)
	y, err = strconv.ParseFloat(query.Get("Y"), 64)
	if err != nil {
		fmt.Printf("%s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("X: %v, Y: %v\n\n", x, y)
	z := complex(x, y)
	num_iteration := generate.CheckInclusion(z)
	w.Header().Add("Content-Type", "application/json")
	responseBody := CheckRequestResponse{num_iteration}
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/check-inclusion", CheckInclusionEndpoint)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server closed unexpectedly")
		}
	}()

	queue := make(chan os.Signal, 1)
	signal.Notify(queue, syscall.SIGINT, syscall.SIGTERM)

	<-queue
	slog.Debug("Starting graceful shutdown")

}
