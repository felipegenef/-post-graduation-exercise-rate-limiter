# Rate Limiter API - Golang (English version below)

![Test Coverage](https://codecov.io/gh/felipegenef/post-graduation-exercise-rate-limiter/branch/main/graph/badge.svg)  
![Test Status](https://github.com/felipegenef/post-graduation-exercise-rate-limiter/actions/workflows/go.yaml/badge.svg)

## Descrição

Este projeto é uma API desenvolvida como parte de um exercício de pós-graduação em Golang. O objetivo é desenvolver um **rate limiter** configurável que controle o número máximo de requisições por segundo com base em **endereços IP** ou em **tokens de acesso**. A lógica de limitação está desacoplada do middleware, e o sistema utiliza **Redis** como mecanismo de persistência, com suporte a troca futura para outras estratégias.

### Requisitos do exercício

- O rate limiter deve ser implementado como um **middleware** injetável em servidores HTTP.
- A lógica de limitação deve ser **separada do middleware**, com uma interface ("strategy") que permita substituição futura do backend (por exemplo, Redis por outro).
- Deve limitar requisições com base em:
  - **IP**: quantidade máxima de requisições por segundo por IP.
  - **Token de Acesso**: definido via header "API_KEY". Quando presente, o limite por token **sobrepõe** o limite por IP.
- Deve ser possível configurar:
  - O **número máximo de requisições por segundo** por IP e por token.
  - O **tempo de bloqueio** após ultrapassar o limite (em segundos).
- As configurações devem ser feitas via **variáveis de ambiente** (ou `.env`).
- Quando o limite for excedido:
  - Código HTTP: "429"
  - Mensagem: "you have reached the maximum number of requests or actions allowed within a certain time frame"
- Toda a lógica de bloqueio e contagem deve ser armazenada no **Redis**.
- O sistema deve responder na porta **8080**.
- Exemplo:
  - Se o IP `192.168.1.1` exceder o limite de 5 req/s, a 6ª requisição é bloqueada.
  - Se o token `abc123` tiver limite de 10 req/s e fizer 11, a 11ª é bloqueada.
  - Após ultrapassar o limite, o IP/token só pode fazer novas requisições após o tempo total de bloqueio (ex: 5 minutos).

## Funcionalidades

- Middleware com controle por IP ou token
- Tokens passados via header: `"API_KEY: <TOKEN>"`
- Redis como armazenamento dos contadores e bloqueios
- Interface abstrata para substituição do mecanismo de persistência
- Tempo de bloqueio e limites configuráveis por variáveis
- Testes automatizados cobrindo concorrência, bloqueios e desbloqueios
- Resposta clara para limites excedidos
- Docker-friendly (uso recomendado de docker-compose para Redis)

## Requisitos

- Go 1.23.3 ou superior
- Redis rodando (pode usar Docker)
- Variáveis de ambiente:
  - "RATE_LIMIT_IP": limite por segundo por IP (ex: 5)
  - "RATE_LIMIT_TOKEN": limite por segundo por token (ex: 10)
  - "BLOCK_DURATION_SECONDS": duração do bloqueio em segundos (ex: 300)
  - "REDIS_ADDR": endereço do Redis (ex: "localhost:6379")
  - "REDIS_PASSWORD": senha do Redis (pode ser vazia)

## Como Testar

### Suba o container redis localmente

```
 docker compose up redis
```

### Execute os testes:

```
 go test ./test/... -v
```

## Como Executar a API

Com as variáveis de ambiente corretamente configuradas:

```
docker compose up -d
```

A aplicação estará disponível em:

```
http://localhost:8080/
```

Teste com token:

```
curl -H "API_KEY: abc123" http://localhost:8080/
```

---

# Rate Limiter API - Golang (Versão em Português acima)

![Test Coverage](https://codecov.io/gh/felipegenef/post-graduation-exercise-rate-limiter/branch/main/graph/badge.svg)  
![Test Status](https://github.com/felipegenef/post-graduation-exercise-rate-limiter/actions/workflows/go.yaml/badge.svg)


## Description

This project is an API developed as part of a postgraduate exercise in Golang. The goal is to create a configurable **rate limiter** that controls the maximum number of requests per second based on **IP addresses** or **access tokens**. The rate limiting logic is decoupled from the middleware, and the system uses **Redis** as a persistence mechanism, allowing future swapping with other strategies.

### Exercise Requirements

- The rate limiter must be implemented as an **injectable middleware** for HTTP servers.
- The rate limiting logic should be **separated from the middleware**, with an interface ("strategy") that allows future backend replacement (e.g., Redis to another).
- It should limit requests based on:
  - **IP**: maximum number of requests per second per IP.
  - **Access Token**: defined via the "API_KEY" header. When present, the token limit **overrides** the IP limit.
- It should be configurable:
  - The **maximum number of requests per second** per IP and per token.
  - The **block duration** after exceeding the limit (in seconds).
- Configurations must be done via **environment variables** (or `.env`).
- When the limit is exceeded:
  - HTTP status code: 429
  - Message: "you have reached the maximum number of requests or actions allowed within a certain time frame"
- All blocking and counting logic must be stored in **Redis**.
- The system should listen on port **8080**.
- Example:
  - If IP 192.168.1.1 exceeds the limit of 5 req/s, the 6th request is blocked.
  - If token abc123 has a limit of 10 req/s and makes 11 requests, the 11th is blocked.
  - After exceeding the limit, the IP/token can only make new requests after the full block duration (e.g., 5 minutes).

## Features

- Middleware controlling access by IP or token
- Tokens passed via header: "API_KEY: <TOKEN>"
- Redis used for storing counters and blocks
- Abstract interface for replacing persistence mechanism
- Block time and limits configurable via environment variables
- Automated tests covering concurrency, blocking, and unblocking
- Clear response when limits are exceeded
- Docker-friendly (recommended use of docker-compose for Redis)

## Requirements

- Go 1.23.3 or higher
- Redis running (Docker can be used)
- Environment variables:
  - RATE_LIMIT_IP: requests per second limit per IP (e.g., 5)
  - RATE_LIMIT_TOKEN: requests per second limit per token (e.g., 10)
  - BLOCK_DURATION_SECONDS: block duration in seconds (e.g., 300)
  - REDIS_ADDR: Redis address (e.g., "localhost:6379")
  - REDIS_PASSWORD: Redis password (can be empty)

## How to Test

### Start Redis container locally

```
docker compose up redis
```

### Run tests:

```
go test ./test/... -v
```

## How to Run the API

With environment variables properly set:

```
docker compose up -d
```

The application will be available at:

```
http://localhost:8080/
```

Test with token:

```
curl -H "API_KEY: abc123" http://localhost:8080/
```