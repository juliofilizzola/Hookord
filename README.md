# Hookord - GitHub ↔ Discord & Slack Notification Bridge

<p align="center">
  <img src="asserts/hookord_github.png" alt="Hookord Banner" width="650"/>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord"/>
  <img src="https://img.shields.io/badge/Slack-4A154B?style=for-the-badge&logo=slack&logoColor=white" alt="Slack"/>
  <img src="https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white" alt="Redis"/>
  <img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"/>
</p>

**Hookord** é uma ponte de notificações de alta performance desenvolvida em Go, conectando webhooks do GitHub a múltiplos serviços de mensageria (**Discord** e **Slack**). O projeto foca em fornecer mensagens modernas, organizadas, ricas em informações e com atualização em tempo real de mensagens existentes.

---

## Sumário

- [Funcionalidades](#-funcionalidades)
- [Arquitetura](#-arquitetura)
- [Endpoints](#-endpoints)
- [Requisitos](#-requisitos)
- [Configuração de Ambiente](#-configuração-de-ambiente)
- [Como Executar](#-como-executar)
  - [Makefile](#atalhos-via-makefile)
  - [Docker & Docker Compose](#via-docker-compose)
  - [Localmente](#executando-localmente)
- [Configuração do Webhook no GitHub](#-configuração-do-webhook-no-github)
- [Automação e CI/CD](#-automação-e-cicd)
- [Status dos Eventos](#-status-dos-eventos-e-integrações)

---

##  Funcionalidades

* **Suporte Multiplataforma (Discord & Slack):** Envio simultâneo ou direcionado de eventos via `IntegrationManager`.
* **Atualização Dinâmica de Mensagens:** Edita mensagens/embeds existentes em vez de criar novas para o mesmo evento (ex: transição de PR de `draft` para `open`, `merged` ou `closed`).
* **Rastreamento de Code Reviews:** Contador de reviews totais e contagem de revisores únicos persistidos no Redis.
* **Classificação Semântica & Alertas Inteligentes:** 
  * Identificação automática de prefixos no título do PR (`feat`, `fix`, `hot`, `doc`, `chore`).
  * Esquema de cores dinâmico de acordo com o status e tipo do evento.
  * Menção automática de `@everyone` para PRs e Issues marcados como `hot`.
* **Roteamento por Canais (Discord):** Canais dedicados por categoria (`pull_requests`, `issues`, `workflows`, `repository`).
* **Observabilidade Integrada:**
  * Endpoint `/health` para sondagem de integridade (Liveness/Readiness).
  * Métricas Prometheus expostas em `/metrics`.
  * Logs estruturados com [zerolog](https://github.com/rs/zerolog) com suporte a níveis de log e ambientes.

---

## 🏛 Arquitetura

O projeto segue os princípios de **Clean Architecture** e **DDD**, desacoplando as regras de negócio das tecnologias externas:

```text
Hookord/
├── cmd/
│   └── hookord/               # Ponto de entrada (main.go)
├── internal/
│   ├── application/           # Serviços de aplicação e orquestração de webhooks
│   ├── common/                # Utilitários globais (formatação de cores, tipos de PR, rodapés)
│   ├── domain/                # Entidades, modelos de domínio e interfaces de repositório
│   ├── events/
│   │   └── review/            # Lógica de contagem e rastreamento de revisores
│   ├── infrastructure/        # Adaptadores externos e infraestrutura
│   │   ├── app/               # Ciclo de vida da aplicação e graceful shutdown
│   │   ├── config/            # Carregamento e validação de configurações de ambiente
│   │   ├── http/              # Servidor HTTP, roteamento e middlewares
│   │   ├── logger/            # Configuração do zerolog
│   │   └── redis/             # Repositório de mapeamento de mensagens em Redis
│   └── integrations/          # Adaptadores dos mensageiros
│       ├── discord/           # Cliente, builders de Embed e handlers do Discord
│       ├── slack/             # Cliente, builders de Blocos/Attachments e handlers do Slack
│       └── manager.go         # Orquestrador de integrações ativas
└── k8s/                       # Manifestos de Deployment e Service para Kubernetes
```

---

##  Endpoints

| Método | Rota | Descrição |
| :--- | :--- | :--- |
| `POST` | `/webhook` | Recebe e valida webhooks enviados pelo GitHub (com validação HMAC SHA-256) |
| `GET` | `/health` | Verificação de integridade do serviço (`200 OK`) |
| `GET` | `/metrics` | Métricas operacionais no formato Prometheus |

---

##  Requisitos

* **Go**: 1.22+
* **Redis**: Instância ativa para persistência do estado e mapeamento de mensagens
* **Docker & Docker Compose** *(opcional para ambiente conteinerizado)*

---

## ⚙ Configuração de Ambiente

Crie um arquivo `.env` na raiz do projeto com base no arquivo `.env.example`:

```bash
cp .env.example .env
```

### Variáveis Disponíveis

| Variável | Descrição | Obrigatória | Exemplo / Padrão |
| :--- | :--- | :---: | :--- |
| `APP_ENV` | Ambiente da aplicação (`development`, `production`) | Não | `development` |
| `PORT` | Porta onde o servidor HTTP escuta | Não | `8080` |
| `LOG_LEVEL` | Nível dos logs (`debug`, `info`, `warn`, `error`) | Não | `info` |
| `GITHUB_SECRET` | Secret configurado no webhook do GitHub para validação de assinatura | **Sim** | `seu_segredo_webhook` |
| `REDIS_URL` | URL de conexão da instância do Redis | **Sim** | `redis://localhost:6379` |
| `DISCORD_TOKEN` | Token do Bot do Discord | Condicional* | `seu_token_discord` |
| `DISCORD_CHANNEL_PULL_REQUESTS` | ID do canal do Discord para Pull Requests | Condicional* | `123456789012345678` |
| `DISCORD_CHANNEL_ISSUES` | ID do canal do Discord para Issues | Não | `123456789012345678` |
| `DISCORD_CHANNEL_WORKFLOWS` | ID do canal do Discord para Workflows | Não | `123456789012345678` |
| `DISCORD_CHANNEL_REPOSITORY` | ID do canal do Discord para eventos de repositório | Não | `123456789012345678` |
| `SLACK_TOKEN` | Token de Bot da API do Slack (`xoxb-...`) | Condicional* | `xoxb-seu-token-slack` |
| `SLACK_CHANNEL_ID` | ID do canal do Slack padrão para notificações | Condicional* | `C0123456789` |

> *\* As credenciais do Discord e Slack são necessárias de acordo com os adaptadores que você deseja manter habilitados.*

---

## 🚀 Como Executar

### Atalhos via Makefile

O projeto inclui um `Makefile` com comandos utilitários:

```bash
make help    # Exibe todos os comandos disponíveis
make build   # Compila o binário na raiz
make run     # Compila e executa o binário
make dev     # Inicia o ambiente de desenvolvimento com live-reload via Docker
make test    # Executa os testes unitários
make tidy    # Organiza e limpa as dependências (go mod tidy)
make clean   # Remove artefatos de compilação
```

### Via Docker Compose

Suba a aplicação junto com o Redis:

```bash
# Modo produção/padrão
docker-compose up -d app

# Modo desenvolvimento (com hot-reload)
docker-compose up app-dev
```

### Executando Localmente

1. Certifique-se de que o Redis está em execução (`localhost:6379`).
2. Execute a aplicação:

```bash
go run cmd/hookord/main.go
```

---

## 🔗 Configuração do Webhook no GitHub

1. No repositório ou organização no GitHub, vá em **Settings** -> **Webhooks** -> **Add webhook**.
2. Preencha os campos:
   * **Payload URL**: `https://seu-dominio.com/webhook`
   * **Content type**: `application/json`
   * **Secret**: O mesmo valor informado na variável `GITHUB_SECRET` do seu `.env`.
   * **SSL verification**: Habilitado (recomendado).
3. Em **Which events would you like to trigger this webhook?**:
   * Escolha **Let me select individual events**.
   * Marque:
     * `Pull requests`
     * `Issues`
4. Clique em **Add webhook**.

---

## 🛡 Automação e CI/CD

O repositório conta com workflows configurados no GitHub Actions:

1. **Validação de Código (`go-code-validation.yml`):**
   * Executado em PRs abertos ou atualizados para o branch `main`.
   * Executa `gofmt -s` (formatação de código).
   * Executa `go vet ./...` (análise estática).
   * Executa testes com detecção de race condition: `go test -v -race ./...`.

2. **Revisão Automática de PRs (`pr-request-changes-on-validation-failure.yml`):**
   * Avalia o resultado da validação.
   * Se a validação falhar, envia um **Request Changes** automático explicando os erros.
   * Se a validação passar com sucesso, adiciona um comentário de aprovação e validação automática.

---

## 📊 Status dos Eventos e Integrações

| Evento GitHub | Ações Suportadas | Discord | Slack |
| :--- | :--- | :---: | :---: |
| **Pull Requests** | `opened`, `reopened`, `closed`, `synchronize`, `ready_for_review` | ✅ Suportado | ✅ Suportado |
| **Issues** | `opened`, `reopened`, `closed` | ✅ Suportado | ⏳ Em breve |
| **Workflows** | Execuções e status de Actions | ⏳ Em breve | ⏳ Em breve |
| **Repository** | `push`, `release`, `tag` | ⏳ Em breve | ⏳ Em breve |

---

<p align="center">
  Desenvolvido com 💙 por <b>Julio Filizzola</b>
</p>
