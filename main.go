package main

import (
	"fmt"
	"github.com/juliofilizzola/Hookord/internal/notification"
	"log"
)

func main() {
	StartNotificationService()
}

func StartNotificationService(cfg *notification.Config) {
	notification.StartNotificationService(cfg)
	fmt.Println("Serviço de notificação iniciado")
	log.Println("Serviço de notificação iniciado com sucesso")
}
