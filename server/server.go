package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/k911mipt/geolocation/store"
)

type Server struct {
	db DBClient
}

func New(db DBClient) *Server {
	return &Server{
		db: db,
	}
}

type DBClient interface {
	FetchGeoInfo(ctx context.Context, ip string) (store.IpGeoInfo, error)
}

func (s *Server) Run(ctx context.Context, port string) error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/geolocation/{ip}", getGeoInfoHandler(s.db)).Methods("GET")
	srv := &http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: router,
	}

	errChan := make(chan error, 1)

	go func() {
		<-ctx.Done()
		log.Print("Stopping server... ")
		shutdownCtx, stop := context.WithTimeout(ctx, 5*time.Second)
		defer stop()
		errChan <- srv.Shutdown(shutdownCtx)
		close(errChan)
		log.Println("Server stopped")
	}()

	log.Printf("Starting http server on :%s", port)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}

	select {
	case err := <-errChan:
		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("Error during stoping server: %w", err)
	default:
		return nil
	}
}

func getGeoInfoHandler(db DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := mux.Vars(r)["ip"]
		log.Printf("[INFO] Recieved geoinfo request for ip %s", ip)

		if net.ParseIP(ip) == nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid IP")
			return
		}

		ipGeoInfo, err := db.FetchGeoInfo(r.Context(), ip)
		if errors.Is(err, store.ErrNotFound) {
			writeErrorResponse(w, http.StatusNotFound, "No geoinfo found for the given IP")
			return
		}
		if err != nil {
			log.Print("FetchGeoInfo error:", err)
			writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		writeJSONResponse(w, http.StatusOK, ipGeoInfo)
	}
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, code int, resp store.IpGeoInfo) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Print("[ERROR] writing the response: ", err)
	}
}

func writeErrorResponse(w http.ResponseWriter, code int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(errorResponse{
		Code:    code,
		Message: errMsg,
	})
	if err != nil {
		log.Print("[ERROR] writing the response: ", err)
	}
}
