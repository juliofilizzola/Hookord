package main

import (
	"github.com/juliofilizzola/Hookord/internal/config"
	"github.com/juliofilizzola/Hookord/internal/notification"
	"log"
)

func main() {
	cfg := config.Load()
	log.Println("Iniciando serviço...")
	notification.StartNotificationService(cfg)
}
