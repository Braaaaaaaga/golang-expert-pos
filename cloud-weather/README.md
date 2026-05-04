# Cloud Weather

Serviço em Go que recebe um CEP, busca a cidade via ViaCEP e retorna a temperatura atual em Celsius, Fahrenheit e Kelvin.

## URL no Cloud Run

```
https://<sua-url>.run.app
```

> Substitua pela URL gerada após o deploy no Cloud Run.

## Requisição

```bash
curl https://<sua-url>.run.app/01310100
```

Resposta (200):

```json
{"temp_C":25.0,"temp_F":77.0,"temp_K":298}
```

| Situação | HTTP | Mensagem |
|---|---|---|
| CEP inválido (≠ 8 dígitos ou não numérico) | 422 | `invalid zipcode` |
| CEP não encontrado | 404 | `can not find zipcode` |

## Como rodar localmente com Docker

```bash
docker build -t cloud-weather .
docker run -p 8080:8080 -e WEATHER_API_KEY=sua_chave cloud-weather
```

Depois:

```bash
curl http://localhost:8080/01310100
```

## Como rodar os testes

Nenhuma dependência externa necessária (APIs são mockadas):

```bash
go test ./... -v
```

## Deploy no Cloud Run

```bash
# Build e push para o Artifact Registry
gcloud builds submit --tag gcr.io/SEU_PROJETO/cloud-weather

# Deploy
gcloud run deploy cloud-weather \
  --image gcr.io/SEU_PROJETO/cloud-weather \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=sua_chave
```

A WEATHER_API_KEY é obtida gratuitamente em https://www.weatherapi.com/.
