# Labs Auction GoExpert

Sistema de leilões com fechamento automático via Goroutines.

## Como rodar

```bash
docker compose up --build
```

A API estará disponível em `http://localhost:8080`.

## Variáveis de ambiente

Configuradas em `cmd/auction/.env`:

| Variável | Descrição | Padrão |
|---|---|---|
| `AUCTION_DURATION` | Duração do leilão até fechar automaticamente (ex: `5m`, `30s`, `1h`) | `5m` |
| `AUCTION_INTERVAL` | Intervalo usado pelo batch de bids para checar expiração | `20s` |
| `BATCH_INSERT_INTERVAL` | Intervalo de inserção em lote de bids | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de bids | `4` |
| `MONGODB_URL` | URL de conexão com o MongoDB | — |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |

Exemplo para leilão com 30 segundos de duração:

```env
AUCTION_DURATION=30s
```

## Como rodar os testes

> Requer Docker em execução (os testes sobem um container MongoDB via Testcontainers).

```bash
go test ./internal/infra/database/auction/... -v -timeout 60s
```

## Funcionamento do fechamento automático

Ao criar um leilão, uma Goroutine é iniciada em background. Ela aguarda o tempo definido em `AUCTION_DURATION` e então atualiza o status do leilão para `Completed` diretamente no MongoDB.
