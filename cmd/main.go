package main

import (
    "dang/controller/httpserver"
    "dang/repository/kopuro"
    "dang/service"
    "os"
)

func main() {
    httpServerPort := os.Getenv("PORT")
    kopuroBaseURL := os.Getenv("KOPURO_BASE_URL")
    kopuroRepo := kopuro.NewRepository(kopuroBaseURL)
    decisionService := service.NewDecisionService(kopuroRepo)
    httpServer := httpserver.NewServer(httpServerPort, decisionService)
    httpServer.Start()
}