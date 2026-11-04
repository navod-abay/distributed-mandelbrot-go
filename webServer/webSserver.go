package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	generate "github.com/navod-abay/mandelbrotset-go/core"
	"github.com/navod-abay/mandelbrotset-go/core/models"
)

type CheckRequestResponse struct {
	Num_iterations uint16 `json:"num_iteration"`
}

type ImageRequest struct {
	ImageDimensaion       models.ImageDimensions
	Subdivision_levels    int  `json:"subdivision_levels"`
	OptimizationFlag      bool `json:"optimization_flag"`
	ColorFlag             bool `json:"color_flag"`
	CsvWriteFlag          bool `json:"csv_write_flag"`
	BmpWriteFlag          bool `json:"bmp_write_flag"`
	SaveSnapShotsFlag     bool `json:"save_snapshot_flag"`
	NoParellalizationFlag bool `json:"no_parellalization_flag"`
}

func ImageRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req ImageRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid Request", 400)
		return
	}
	fmt.Printf("%v\n", req)
	imageDimensions := generate.CalculatePixelSize(req.ImageDimensaion, req.Subdivision_levels)
	flags := models.Flags{
		OptimizationFlag:      req.OptimizationFlag,
		ColorFlag:             req.ColorFlag,
		CsvWriteFlag:          req.CsvWriteFlag,
		BmpWriteFlag:          req.BmpWriteFlag,
		SaveSnapShotsFlag:     req.SaveSnapShotsFlag,
		NoParellalizationFlag: req.NoParellalizationFlag,
	}
	generate.RouteToSolverOnFlage(flags, imageDimensions, req.Subdivision_levels)

}

func CheckInclusionEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handler got called: \n %v\n	", *r)
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

func printRequestBodyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server error", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Requset Body:\n %v\n", string(body))
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("GET /check-inclusion", printRequestBodyHandler(http.HandlerFunc(CheckInclusionEndpoint)))

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
