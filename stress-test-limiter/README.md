# Stress Test CLI

Ferramenta de linha de comando para testes de carga em serviços web.

## Como usar

### Compilar e executar:
```bash
go build -o stress-test .
./stress-test --url=http://google.com --requests=100 --concurrency=10
```

### Ou executar direto:
```bash
go run main.go --url=http://google.com --requests=100 --concurrency=10
```

## Parâmetros

- `--url`: URL do serviço para testar
- `--requests`: Número total de requisições  
- `--concurrency`: Quantas requisições simultâneas

## Exemplo prático

```bash
./stress-test --url=http://httpbin.org/status/200 --requests=50 --concurrency=5
```

**Resultado:**
```
🚀 Initiating Load Test
Target URL: http://httpbin.org/status/200
Total Requests: 50
Concurrent Workers: 5
----------------------------------------

📊 === LOAD TEST RESULTS ===
⏱️  Execution Time: 2.5s
📝 Total Requests: 50
✅ Successful (200 OK): 50
📈 Success Rate: 100.00%

📋 Status Code Breakdown:
   ✅ HTTP 200: 50 (100.0%)
```

## Docker

```bash
# Construir imagem
docker build -t stress-test .

# Executar teste
docker run stress-test --url=http://google.com --requests=1000 --concurrency=10
```

## Testando diferentes cenários

```bash
# Teste rápido
./stress-test --url=http://httpbin.org/status/200 --requests=20 --concurrency=5

# Teste com erro 404
./stress-test --url=http://httpbin.org/status/404 --requests=10 --concurrency=2

# Teste com delay
./stress-test --url=http://httpbin.org/delay/2 --requests=10 --concurrency=3
```