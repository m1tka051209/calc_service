package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/m1tka051209/calc_service/internal/app"
	"github.com/m1tka051209/calc_service/internal/config"
	"github.com/m1tka051209/calc_service/internal/calc"
)

func main() {
	log.Println("Starting application...")

	cfg := config.New()
	calculator := calc.Calc
	application := app.New(cfg, calculator)

    var portNumber int
    flag.IntVar(&portNumber, "port", 8080, "Port number to listen on")
    flag.Parse()
    log.Printf("Starting server on port: %d\n", portNumber)
    
    // Create a context that can be cancelled
    ctx, cancel := context.WithCancel(context.Background())

    // Graceful Shutdown with signal handler
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func(){
        <-quit
        log.Println("Shutdown signal received")
        cancel()
    }()

    // Start a go routine to listen to console input
    go func() {
        reader := bufio.NewReader(os.Stdin)
        for {
            fmt.Print("Enter command: ")
            input, _ := reader.ReadString('\n')
            input = strings.TrimSpace(input)

            if input == "exit" {
                log.Println("Exit command received, shutting down...")
                cancel()
                return
            }
            if input == "" {
                continue
            }
        }
    }()

    // Start the server and check for context cancellation for shutdown
    err := application.RunServer(portNumber, ctx)
    if err != nil && err != http.ErrServerClosed {
        log.Fatalf("Failed to start server: %v", err)
    }
    log.Println("Server shut down gracefully")
}