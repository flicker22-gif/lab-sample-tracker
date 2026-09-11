package main

import (
	"log"
	"os"

	"sample-tracker/internal/database"
	"sample-tracker/internal/router"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	r := router.New(db)
	addr := ":" + getenv("PORT", "8080")
	log.Printf("API 服务启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
