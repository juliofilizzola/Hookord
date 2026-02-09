package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Estrutura para receber o payload do GitHub
// Simplificada para PRs

type PullRequestEvent struct {
	Action      string `json:"action"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// Função para enviar mensagem ao Discord
func sendDiscordNotification(webhookURL, message string) error {
	payload := map[string]string{"content": message}
	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", io.NopCloser(bytes.NewReader(jsonPayload)))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// Handler para receber webhooks do GitHub
func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var event PullRequestEvent
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Erro ao ler body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Println("Erro ao decodificar JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if event.Action == "opened" || event.Action == "closed" || event.Action == "merged" {
		msg := "PR " + event.Action + ": " + event.PullRequest.Title + "\nAutor: " + event.PullRequest.User.Login + "\nRepo: " + event.Repository.FullName + "\nURL: " + event.PullRequest.HTMLURL
		webhookURL := "COLOQUE_AQUI_O_WEBHOOK_DO_DISCORD"
		err := sendDiscordNotification(webhookURL, msg)
		if err != nil {
			log.Println("Erro ao enviar para Discord:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func StartNotificationService() {
	http.HandleFunc("/github-webhook", githubWebhookHandler)
	log.Println("Servidor de webhook iniciado na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
