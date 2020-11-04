package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/iamtraining/chat/auth"
	"github.com/iamtraining/chat/chat"
	"github.com/iamtraining/chat/views"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("couldnt load .env")
	}
}

func main() {
	addr := flag.String("addr", ":3000", "server address")
	flag.Parse()

	fmt.Println(auth.Conf.ClientID)
	fmt.Println(auth.Conf.ClientSecret)

	handler := mux.NewRouter()

	srv := http.Server{
		Addr:    *addr,
		Handler: handler,
	}

	chat := chat.NewChatServer(context.Background(), 100)

	// chat
	handler.Handle("/chat/{room}", &views.Template{Filename: "chat.gohtml"})
	handler.HandleFunc("/join/{room}", chat.Join)

	// auth
	handler.Handle("/login", &views.Template{Filename: "login.gohtml"})
	handler.HandleFunc("/auth/google/login", auth.LoginHandler)
	handler.HandleFunc("/auth/google/callback", auth.CallbackHandler)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	log.Println("server started", *addr)

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
