# API HTTP do GoPing

Base local: `http://localhost:8080`

Todas as respostas de sucesso com conteúdo usam o envelope `{"data": ...}`. Erros usam `{"error":{"code":"...","message":"..."}}`.

## Saúde

### `GET /healthz`

Confirma que o processo HTTP está respondendo.

```json
{
  "service": "goping",
  "status": "ok",
  "time": "2026-10-08T18:30:00Z"
}
```

## Monitores

### `GET /api/v1/monitors`

Lista os monitores na ordem mais recente primeiro.

### `POST /api/v1/monitors`

Cria um monitor no estado inicial `unknown`.

```json
{
  "name": "API pública",
  "url": "https://example.com/health",
  "interval_seconds": 300,
  "timeout_seconds": 10
}
```

Validações atuais:

- nome entre 1 e 80 caracteres;
- URL absoluta com esquema `http` ou `https` e sem credenciais;
- intervalo entre 300 e 86400 segundos;
- timeout entre 1 e 30 segundos;
- corpo máximo de 1 MiB, sem campos desconhecidos.

Retorna `201 Created` quando criado e `422 Unprocessable Entity` para valores inválidos.

### `DELETE /api/v1/monitors/{id}`

Exclui um monitor e registra a atividade. Retorna `204 No Content` ou `404 Not Found`.

## Atividades

### `GET /api/v1/activities`

Lista as atividades mais recentes.

### `DELETE /api/v1/activities`

Limpa o registro atual e retorna `204 No Content`.

## Exemplos com curl

```bash
curl http://localhost:8080/healthz

curl -X POST http://localhost:8080/api/v1/monitors \
  -H 'Content-Type: application/json' \
  -d '{"name":"Meu site","url":"https://example.com","interval_seconds":300,"timeout_seconds":10}'
```

## Limites atuais

A API ainda não possui autenticação nem rate limiting e não deve ser exposta para cadastro público. O servidor não executa as URLs cadastradas; checker, proteção completa contra SSRF e histórico pertencem aos próximos marcos.
