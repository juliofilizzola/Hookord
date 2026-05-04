# Hookord 🚀

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)
[![Docker Image](https://img.shields.io/badge/Docker-20MB-2496ED?logo=docker)](https://hub.docker.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Hookord** é um webhook forwarder de alta performance em Go que recebe notificações do GitHub e as envia para Discord com embeds bonitos e funcionais. Projetado com arquitetura de **providers** extensível (fácil adicionar Slack, Teams, etc.).

## ✨ Funcionalidades

- **Recebe GitHub Webhooks** (push, PRs, issues, releases, ping) com validação HMAC
- **Embeds Discord profissionais** com cores dinâmicas, author, footer, fields inline
- **Arquitetura extensível**: providers de input/output desacoplados
- **Structured logging** com Zerolog (JSON, pronto pra ELK/Loki)
- **Config YAML + ENV vars** (12-factor app)
- **Docker multi-stage** (<20MB) + Healthchecks
- **K8s ready** com non-root user e readiness probes

## 📁 Estrutura do Projeto

```text
Hookord/
├── cmd/
│   └── hookord/
│       └── main.go
├── configs/
│   ├── config.example.yaml
│   └── config.yaml
├── deployments/
│   └── docker/
│       ├── docker-compose.yml
│       └── Dockerfile
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── structs.go
│   ├── core/
│   │   ├── dispatcher.go
│   │   ├── event.go
│   │   └── ports.go
│   ├── httpserver/
│   │   ├── github_handler.go
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   └── router.go
│   ├── log/
│   │   ├── logger.go
│   │   └── structs.go
│   └── providers/
│       ├── interface.go
│       ├── discord/
│       │   ├── client.go
│       │   ├── embeds.go
│       │   └── stucts.go
│       └── github/
│           ├── mappers.go
│           ├── parser.go
│           └── stucts.go
├── mock/
│   └── discord.html
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Quickstart

### 1. Clone e configure
```bash
git clone https://github.com/seuuser/hookord
cd hookord
cp configs/config.example.yaml configs/config.yaml
# edite com seu GITHUB_WEBHOOK_SECRET e DISCORD_WEBHOOK_URL
```

### 2. Docker Compose (recomendado)
```bash
docker-compose up --build
# testa:
curl http://localhost:8080/healthz
```

### 3. Teste webhook
```bash
curl -X POST http://localhost:8080/webhooks/github \
  -H "X-Hub-Signature-256: sha256=$(echo -n '{\"ref\":\"refs/heads/main\"}' | openssl sha256 -hmac 'SEU_SECRET' | cut -d' ' -f2)" \
  -H "X-GitHub-Event: push" \
  -d '{"ref":"refs/heads/main","repository":{"full_name":"meuorg/meu-repo"},"pusher":{"name":"test"}}'
```

## 🔧 Configuração

`configs/config.yaml`

```yaml
app:
  name: "hookord"
  version: "v1.0.0"

http:
  port: "8080"

github:
  webhook_secret: "sha256=SEU_GITHUB_WEBHOOK_SECRET"
  allowed_repos:
    - "sua-org/seu-repo"

discord:
  webhook_url: "https://discord.com/api/webhooks/ID/TOKEN"
```

OU variáveis de ambiente:

```bash
GHNOTIFY_GITHUB_WEBHOOK_SECRET=sha256=secret
GHNOTIFY_DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
```

## 🚀 Deploy Produção

### Docker
```bash
docker build -f deployments/docker/Dockerfile -t hookord:latest .
docker run -p 8080:8080 \
  -e GHNOTIFY_GITHUB_WEBHOOK_SECRET=sha256=prod \
  -e GHNOTIFY_DISCORD_WEBHOOK_URL=https://... \
  hookord:latest
```

### Kubernetes (exemplo mínimo)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hookord
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: hookord
        image: hookord:latest
        ports:
        - containerPort: 8080
        env:
        - name: GHNOTIFY_GITHUB_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: hookord-secrets
              key: webhook-secret
---
apiVersion: v1
kind: Service
metadata:
  name: hookord
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: hookord
```

## 📊 Observability

Logs estruturados JSON (Grafana Loki, ELK):

```json
{"level":"info","time":1707890000,"service":"hookord","request_id":"abc123","path":"/webhooks/github","msg":"webhook processed"}
```

Endpoints:
- `GET /healthz` → healthcheck
- `GET /metrics` → Prometheus (futuro)

## 🧪 Testando Local
- Discord Mock: docker-compose up inclui mock em http://localhost:8081
- GitHub Payloads: exemplos em test/payloads/ (futuro)
- Unit tests: go test ./...

## 🔮 Roadmap
- GitHub → Discord (v1.0)
- Slack/Teams providers
- Rate limiting + retry queue (Redis)
- Web UI dashboard
- Multiple GitHub repos/orgs
- Prometheus metrics
- Horizontal scaling (sticky sessions)

## 🙌 Contribuições
- Fork o projeto
- Crie feature branch (`git checkout -b feature/slack-provider`)
- Commit suas mudanças (`git commit -m 'Add Slack provider'`)
- Push pro branch (`git push origin feature/slack-provider`)
- Abra Pull Request

## 📄 Licença
[MIT License](LICENSE)

## 👥 Autores
Julio Filizzola - DevOps Engineer

⭐ Star pra ajudar a comunidade!
🐛 Issues: abra um issue

<div align="center"> <img src="https://img.shields.io/badge/built%20with-%E2%9D%A4%EF%B8%8F%20by%20Julio Filizzola-FF6B6B" alt="built with love"> </div>
