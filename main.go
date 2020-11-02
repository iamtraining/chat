package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/iamtraining/chat/chat"
)

func main() {
	addr := flag.String("addr", ":3000", "server address")
	flag.Parse()

	handler := mux.NewRouter()

	srv := http.Server{
		Addr:    *addr,
		Handler: handler,
	}

	chat := chat.NewChatServer(context.Background(), 100)
	handler.HandleFunc("/chat/{room}", chat.ServeHTTP)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	log.Println("serrver started", addr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	log.Println("server stopping", "why ", <-sig)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failure")
	}

	log.Println("goodbye")
}
