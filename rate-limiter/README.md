# Rate Limiter Go

Rate limiter que controla requisições por IP e token de acesso com suporte a Redis.

## Funcionalidades

- Rate limiting por IP
- Rate limiting por Token via header `API_KEY`  
- Token sobrescreve limitações de IP
- Storage strategy (Redis + fallback memória)
- Middleware Gin
- Bloqueio temporal configurável
- Docker + Docker Compose

## Execução

```bash
docker-compose up -d
```

## Configuração

```env
REDIS_HOST=localhost
REDIS_PORT=6379
RATE_LIMIT_PER_IP=10
RATE_LIMIT_BLOCK_TIME_IP=300
API_TOKENS=abc123:100:600,def456:50:300,xyz789:200:900
SERVER_PORT=8080
```

## Uso

```bash
curl http://localhost:8080/test
curl -H "API_KEY: abc123" http://localhost:8080/test
```

## Testes

```bash
go test ./tests/
./stress-test.sh
```

## Estrutura

```
├── main.go                     # Ponto de entrada da aplicação
├── internal/
│   ├── config/                 # Configurações da aplicação
│   │   └── config.go
│   ├── storage/                # Strategy pattern para storage
│   │   ├── interface.go        # Interface Storage
│   │   ├── redis.go            # Implementação Redis
│   │   └── memory.go           # Implementação In-Memory
│   ├── limiter/                # Lógica core do rate limiter
│   │   └── limiter.go
│   └── middleware/             # Middleware Gin
│       └── ratelimit.go
├── tests/
├── docker-compose.yml
└── Dockerfile
```

Servidor responde na porta 8080. Rate limit excedido retorna HTTP 429 com mensagem: "you have reached the maximum number of requests or actions allowed within a certain time frame".