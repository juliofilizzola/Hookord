package notification

import (
	"io"
	"log"
	"net/http"

	"github.com/juliofilizzola/Hookord/internal/config"
)

func StartNotificationService(cfg *Config) {
	http.HandleFunc("/github-webhook", Handler(cfg))
	log.Println("Servidor de webhook iniciado na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func Handler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Erro ao ler body:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		event, err := ParsePullRequestEvent(body)
		if err != nil {
			log.Println("Erro ao decodificar JSON:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if event.Action == "opened" || event.Action == "closed" || event.Action == "merged" {
			msg := "PR " + event.Action + ": " + event.PullRequest.Title + "\nAutor: " + event.PullRequest.User.Login + "\nRepo: " + event.Repository.FullName + "\nURL: " + event.PullRequest.HTMLURL
			err := SendDiscordNotification(cfg.DiscordWebhookURL, msg)
			if err != nil {
				log.Println("Erro ao enviar para Discord:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
