package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/m1tka051209/calc_service/internal/calc"
	"github.com/m1tka051209/calc_service/internal/config"
)

type Application struct {
	config *config.Config
	calc   func(expression string) (float64, error)
}

func New(config *config.Config, calc func(expression string) (float64, error)) *Application {
	return &Application{
		config: config,
		calc:   calc,
	}
}

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result float64 `json:"result"`
}

type Error struct {
	Error string `json:"error"`
}

func (a *Application) RunServer(port int, ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", a.CalcHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	log.Printf("Starting server on :%d\n", port)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server")

	// Timeout if server doesn't shutdown gracefully
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed:%+s", err)
	}
	log.Printf("Server stopped on :%d\n", port)

	return nil
}

func (a *Application) CalcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := a.calc(req.Expression)
	if err != nil {
		if errors.Is(err, calc.ErrInvalidExpression) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Error{Error: "Expression is not valid"})
		} else if errors.Is(err, calc.ErrDivideByZero) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{Error: "Division by zero"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{Error: "Internal Server Error"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Result: result})
}