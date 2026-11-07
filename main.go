package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"bookmart/backend/auth"
	"bookmart/backend/handler"
	"bookmart/backend/mw"
	"bookmart/backend/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	loadDotEnv(".env")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL ต้องตั้งค่าใน Environment Variables")
	}
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("DATABASE_URL ไม่ถูกต้อง")
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("เปิดการเชื่อมต่อฐานข้อมูลไม่สำเร็จ")
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("เชื่อมต่อ PostgreSQL ไม่สำเร็จ")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	origins := splitOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(origins) == 0 {
		origins = []string{"http://localhost:3000"}
	}
	h := &handler.Handler{Store: store.New(pool), Tickets: auth.NewIssuer(secret)}
	var app http.Handler = http.TimeoutHandler(h.Routes(), 10*time.Second, `{"error":"คำขอใช้เวลานานเกินไป"}`)
	app = mw.Logging(mw.Recovery(mw.CORS(origins, app)))
	server := &http.Server{Addr: ":" + port, Handler: app, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() { log.Printf("BookMart API listening on :%s", port); errCh <- server.ListenAndServe() }()
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}
}

// loadDotEnv reads local development settings without overriding environment
// variables supplied by the deployment platform.
func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			if parsed, err := strconv.Unquote(value); err == nil {
				value = parsed
			}
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func splitOrigins(value string) []string {
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
