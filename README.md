# Hookord

Serviço de notificação de Pull Requests do GitHub para Discord.

## Como usar

1. Configure o webhook do Discord e defina a variável de ambiente `DISCORD_WEBHOOK_URL`.
2. Configure o webhook do GitHub apontando para `http://<seu-servidor>:8080/github-webhook`.
3. Rode o serviço:

```sh
go run cmd/hookord/main.go
```

## Estrutura

- `cmd/hookord/main.go`: ponto de entrada
- `internal/notification/`: lógica de integração
- `internal/config/`: configuração

