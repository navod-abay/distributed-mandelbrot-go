package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/navod-abay/mandelbrotset-go/core"
	generate "github.com/navod-abay/mandelbrotset-go/core"
	"github.com/navod-abay/mandelbrotset-go/core/models"
	"github.com/navod-abay/mandelbrotset-go/core/writers"
)

type CheckRequestResponse struct {
	Num_iterations uint16 `json:"num_iteration"`
}

type ImageRequest struct {
	ImageDimension     *models.ImageDimensions `json:"image_dimensions"`
	Subdivision_levels int                     `json:"subdivision_levels"`
	ColorFlag          bool                    `json:"color_flag"`
	BmpWriteFlag       bool                    `json:"bmp_write_flag"`
}

func ImageRequestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Image Handler got called")
	var req ImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		slog.Debug("Error: %v", err)
		return
	}
	if req.ImageDimension == nil || ValidateImageRequest(*req.ImageDimension) != nil {
		http.Error(w, "No Image Dimensions Specified", http.StatusBadRequest)
		return
	}
	fmt.Printf("%v\n", req)
	imageDimensions := generate.CalculatePixelSize(*req.ImageDimension, req.Subdivision_levels)
	flags := models.Flags{
		OptimizationFlag:      true,
		ColorFlag:             req.ColorFlag,
		CsvWriteFlag:          false,
		BmpWriteFlag:          req.BmpWriteFlag,
		SaveSnapShotsFlag:     false,
		NoParellalizationFlag: false,
	}
	filename := "output.bmp"
	solverResult := generate.RouteToSolverOnFlags(flags, imageDimensions, req.Subdivision_levels)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=filename")
	var wg sync.WaitGroup
	writers.StreamBmpFile(w, *solverResult.ColorMatrix, imageDimensions, filename, core.Maximum_iteration_depth, &wg)
	wg.Wait()
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

func panicRecoverMiddlewear(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Error: %v\n Stack: %v", err, string(debug.Stack()))

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	tlsFlag := flag.Bool("tls", false, "Set to true if using TLS. certificates should be in the certs folder")
	flag.Parse()
	mux := http.NewServeMux()

	mux.Handle("GET /check-inclusion", http.HandlerFunc(CheckInclusionEndpoint))
	mux.HandleFunc("POST /image", ImageRequestHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      panicRecoverMiddlewear(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		var err error
		if *tlsFlag {
			fmt.Println("Running wit HTTPS")
			err = server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem")
		} else {
			fmt.Println("Running with just HTTP")
			err = server.ListenAndServe()
		}
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server closed unexpectedly")
		}
	}()

	queue := make(chan os.Signal, 1)
	signal.Notify(queue, syscall.SIGINT, syscall.SIGTERM)

	<-queue
	slog.Debug("Starting graceful shutdown")

}
