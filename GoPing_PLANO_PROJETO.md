# GoPing — Plataforma Open Source de Monitoramento de Sites e APIs

> **Status:** planejamento · **Linguagem principal:** Go · **Licença:** a definir · **Objetivo:** aprendizado profundo de backend e criação de uma ferramenta útil, com hospedagem econômica.

## 1. Visão do produto

O **GoPing** é uma plataforma de *uptime monitoring*: o usuário cadastra URLs HTTP/HTTPS, define a frequência de verificação e acompanha disponibilidade, tempo de resposta, falhas e histórico. Em versões futuras, recebe alertas e pode publicar uma página de status.

**Problema resolvido:** pequenos desenvolvedores, freelancers e equipes precisam saber quando seus sites e APIs deixam de responder, sem instalar uma infraestrutura complexa.

**Público inicial:** desenvolvedores independentes, donos de pequenos sites e mantenedores de APIs.

**Diferencial do projeto:** simplicidade, baixo custo, transparência do código e implementação didática em Go.

### Não objetivos do MVP

- Não construir uma alternativa completa ao Datadog, Grafana Cloud ou UptimeRobot.
- Não usar microsserviços, Kubernetes, Kafka ou múltiplos bancos no início.
- Não oferecer pagamentos, planos comerciais ou multi-tenancy na primeira versão.
- Não monitorar servidores via agentes instalados nas máquinas no MVP.

## 2. Objetivos de aprendizagem

1. Programar Go sem depender de código gerado por IA.
2. Compreender tipos, funções, structs, interfaces, pacotes e erros.
3. Dominar HTTP, JSON, SQL, testes, logs e APIs REST.
4. Aplicar goroutines, channels, context e limites de concorrência com segurança.
5. Entender arquitetura, segurança, observabilidade e deploy.
6. Publicar um projeto que possa ser demonstrado e explicado em entrevistas.

**Regra de estudo:** o estudante escreve o código; a IA pode explicar conceitos, propor exercícios, revisar e dar pistas. Só solicitar uma solução pronta após tentar resolver e conseguir explicar o resultado.

## 3. Escopo do produto

### MVP funcional (primeira versão web)

- Cadastrar, listar, editar e remover monitores HTTP/HTTPS.
- Definir URL e intervalo de verificação (inicialmente com mínimo configurável, por exemplo, 5 minutos).
- Realizar GET ou HEAD, com timeout configurado.
- Registrar horário, status HTTP, latência e eventual erro.
- Exibir estado atual: `UP`, `DOWN` ou `UNKNOWN`.
- Consultar histórico de verificações.
- Mostrar resumo simples por monitor.
- Executar com um único processo Go e persistência local.

### Depois do MVP

- Notificações por e-mail, Telegram e webhook.
- Política para evitar alertas falsos (ex.: confirmar falha após N verificações).
- Página pública de status.
- Autenticação e múltiplos usuários.
- PostgreSQL, métricas e deploy automatizado.
- Worker pool com concorrência controlada.
- API pública e documentação OpenAPI.

## 4. Roadmap de aprendizagem e entregas

**Importante:** cada etapa deve funcionar isoladamente antes de avançar. Não começar pela arquitetura final.

| Fase | Versão | O que construir | Conceitos de Go | Critério de conclusão |
|---|---|---|---|---|
| 0 | `v0.0` | Ambiente Go, Git, VS Code e primeiro programa | `package`, `main`, `fmt`, `go run`, `go test` | Executar um programa próprio no terminal |
| 1 | `v0.1` | CLI que verifica **uma** URL | variáveis, funções, `net/http`, erros, `defer` | Informar status HTTP ou erro sem travar |
| 2 | `v0.2` | Verificar várias URLs sequencialmente | slices, `for`, structs, organização em funções | Produzir relatório de cada URL |
| 3 | `v0.3` | Timeout, latência e JSON | `time`, `context`, `encoding/json` | Exibir duração e salvar resultado em JSON |
| 4 | `v0.4` | Testes automatizados | `testing`, `httptest`, testes de tabela | Testar sucesso, falha e timeout sem depender da internet |
| 5 | `v0.5` | API HTTP para monitores | `net/http`, rotas, handlers, JSON | Criar e consultar monitores via HTTP |
| 6 | `v0.6` | Persistência com SQLite | `database/sql`, SQL, migrations | Reiniciar app sem perder monitores/histórico |
| 7 | `v0.7` | Scheduler e worker pool | goroutines, channels, mutex, context | Monitorar várias URLs com concorrência limitada e sem vazamento |
| 8 | `v0.8` | Alertas e confiabilidade | interfaces, retries, logs | Alertar apenas após falhas confirmadas |
| 9 | `v0.9` | Docker, deploy e métricas | build, configuração, health check | Serviço público funcionando com orçamento controlado |
| 10 | `v1.0` | Segurança, documentação e UX | validação, testes de integração, profiling | MVP utilizável por terceiros e documentado |

**Nota:** o roadmap é orientativo; as fases 5–9 podem ser reorganizadas após o domínio dos fundamentos.

## 5. Primeiro exercício: GoPing CLI (`v0.1`)

**Desafio:** criar `main.go` que faça uma requisição HTTP a uma URL fixa e imprima:

- URL consultada;
- código HTTP, se houver resposta;
- mensagem de erro clara, se a conexão falhar;
- duração da requisição;
- timeout para não aguardar indefinidamente.

**Exemplo de saída (ilustrativo):**

```text
GoPing v0.1
URL: https://example.com
HTTP: 200
Tempo: 182 ms
Resultado: UP
```

**Não copiar uma implementação pronta.** Começar estudando `http.Client`, `http.NewRequestWithContext`, `time.Now`, `time.Since` e `defer resp.Body.Close()` na documentação. Um HTTP 404 significa que o servidor respondeu, mas não que o endpoint está saudável; definir a regra de saúde explicitamente.

**Exercícios extras:** testar URL inválida, domínio inexistente, resposta 500, timeout e status 301/302. Entender o comportamento padrão de redirecionamentos do cliente HTTP.

## 6. Arquitetura planejada

### Inicial: monólito simples

```text
                  Cliente / Dashboard
                           |
                           v
                     API HTTP (Go)
                           |
                +----------+----------+
                |                     |
          Cadastro de URLs       Consulta de histórico
                |                     |
                +----------+----------+
                           |
                     Banco SQLite
                           ^
                           |
                       Scheduler
                           |
                      Fila interna
                           |
                    Worker Pool Go
                           |
                   HTTP/HTTPS checks
```

- **Um único processo** para API, scheduler e workers.
- **SQLite** inicialmente; PostgreSQL quando houver necessidade real de múltiplas instâncias ou escala.
- **Biblioteca padrão Go** (`net/http`) antes de introduzir frameworks.
- **Configuração via variáveis de ambiente**, sem segredos no repositório.
- Separação entre lógica de domínio, persistência e HTTP apenas quando a complexidade justificar.

### Atenção: scheduler e persistência

A fila em memória **não é durável**: reinícios podem interromper verificações. O banco registra monitores e a próxima execução; o scheduler recupera tarefas pendentes ao reiniciar. Evitar executar a mesma verificação simultaneamente para o mesmo monitor.

## 7. Estrutura de pastas (alvo, não criar tudo agora)

```text
goping/
├── README.md
├── LICENSE                       # escolher antes de publicar
├── go.mod
├── cmd/
│   ├── goping-cli/
│   │   └── main.go
│   └── goping-server/
│       └── main.go
├── internal/
│   ├── monitor/                  # regras de domínio
│   ├── checker/                  # verificação HTTP
│   ├── scheduler/                # agendamento
│   ├── worker/                   # concorrência controlada
│   ├── storage/                  # SQLite/PostgreSQL
│   ├── api/                      # handlers HTTP
│   └── alert/                    # notificações futuras
├── migrations/
├── web/                          # interface futura, opcional
├── docs/
│   ├── architecture.md
│   └── decisions/
├── .github/workflows/            # CI futura
├── Dockerfile                    # em fase posterior
└── .gitignore
```

**Na fase `v0.1`, basta:** `go.mod`, `main.go`, `README.md` e `.gitignore`. Evitar criar diretórios vazios e abstrações antes da necessidade.

## 8. Modelo de dados inicial

### Monitor

| Campo | Tipo lógico | Significado |
|---|---|---|
| `id` | identificador | ID do monitor |
| `name` | texto | Nome amigável |
| `url` | texto | Endereço HTTP/HTTPS |
| `interval_seconds` | inteiro | Frequência de verificação |
| `timeout_seconds` | inteiro | Limite por requisição |
| `enabled` | booleano | Ativo ou pausado |
| `next_check_at` | data/hora | Próxima execução |
| `created_at` | data/hora | Cadastro |

### CheckResult

| Campo | Tipo lógico | Significado |
|---|---|---|
| `id` | identificador | ID do resultado |
| `monitor_id` | identificador | Monitor verificado |
| `checked_at` | data/hora | Instante da tentativa |
| `status_code` | inteiro opcional | Código HTTP recebido |
| `latency_ms` | inteiro | Duração |
| `is_up` | booleano | Resultado segundo regra de saúde |
| `error_message` | texto opcional | Erro de rede/timeout |

**Definição inicial de UP:** status HTTP de 200 a 399. Permitir personalização futura. Resposta 4xx/5xx é falha de saúde, mas não necessariamente indisponibilidade da rede.

**Cálculo de uptime:** percentual de verificações bem-sucedidas em uma janela de tempo, claramente rotulado como estimativa baseada em amostras; não confundir com disponibilidade contínua garantida.

## 9. API REST proposta (fase `v0.5`)

| Método | Rota | Ação |
|---|---|---|
| `GET` | `/healthz` | Saúde do próprio GoPing |
| `POST` | `/api/v1/monitors` | Criar monitor |
| `GET` | `/api/v1/monitors` | Listar monitores |
| `GET` | `/api/v1/monitors/{id}` | Detalhar monitor |
| `PATCH` | `/api/v1/monitors/{id}` | Editar monitor |
| `DELETE` | `/api/v1/monitors/{id}` | Excluir monitor |
| `GET` | `/api/v1/monitors/{id}/checks` | Histórico paginado |
| `GET` | `/api/v1/monitors/{id}/summary` | Resumo de disponibilidade |

Exemplo de cadastro:

```json
{
  "name": "Meu site",
  "url": "https://example.com",
  "interval_seconds": 300,
  "timeout_seconds": 10
}
```

Resposta ilustrativa:

```json
{
  "id": "mon_001",
  "name": "Meu site",
  "url": "https://example.com",
  "enabled": true,
  "status": "UNKNOWN"
}
```

O status começa como `UNKNOWN` até a primeira verificação. Definir códigos de resposta, paginação, erros padronizados e validações durante a implementação.

## 10. Segurança — requisito obrigatório antes de disponibilizar a API publicamente

O GoPing acessa URLs fornecidas por usuários. Isso cria risco de **SSRF (Server-Side Request Forgery)**.

- Validar esquema `http`/`https`, tamanho da URL, portas e credenciais embutidas.
- Bloquear endereços loopback, privados, link-local, multicast e endpoints de metadados de nuvem.
- Resolver DNS e validar os IPs **também no momento de conectar**, para evitar DNS rebinding.
- Validar cada redirecionamento e limitar sua quantidade.
- Restringir portas e métodos HTTP; usar limites de tempo, tamanho de resposta e concorrência.
- Não armazenar tokens/senhas de usuários na URL; ocultar informações sensíveis em logs.
- Autenticar rotas administrativas e aplicar rate limiting antes de liberar cadastro público.
- Evitar que usuários criem monitores capazes de atingir recursos internos da hospedagem.

**Até essas proteções estarem implementadas e testadas, usar o sistema apenas localmente ou com URLs confiáveis cadastradas pelo próprio administrador.**

## 11. Estratégia de testes

- **Unitários:** regras de saúde, cálculo de latência, validação, agendamento e deduplicação.
- **HTTP sem internet:** `httptest.NewServer` para 200, 404, 500, redirects e timeout.
- **Concorrência:** `go test -race ./...` e limites de worker pool.
- **Integração:** operações SQLite e migrations em ambiente temporário.
- **Recuperação:** reiniciar o processo com verificações pendentes e confirmar retomada.
- **Segurança:** testes de bloqueio SSRF, IP privado, redirecionamento e resolução DNS.
- **Qualidade:** `gofmt`, `go vet`, `go test ./...` e CI quando chegar a fase adequada.

## 12. Hospedagem econômica

### Etapa 1 — custo zero

Desenvolver e testar localmente no Linux, com SQLite e dados de teste.

### Etapa 2 — demonstração pública

Avaliar uma pequena VPS ou instância com baixo custo, rodando **um processo Go + SQLite em volume persistente**. O preço depende do provedor e da região; conferir tarifas atuais antes da contratação. Usar backups periódicos e HTTPS.

### Alternativa serverless

Cloud Run pode ser econômico para requisições ocasionais, mas **um scheduler em memória não é confiável em serviços que podem dormir ou escalar a zero**. Nesse caso, usar agendador externo e execução sob demanda, calculando também o custo do scheduler e da persistência.

### Medidas para reduzir custos

- Intervalo mínimo de 5 minutos no plano inicial.
- Timeout de 5–10 segundos por verificação.
- Worker pool com concorrência limitada.
- Retenção curta do histórico detalhado (ex.: 30 dias), com agregados posteriores.
- Não guardar corpo completo das respostas HTTP.
- Limites por usuário e quotas de verificação.
- Métricas simples antes de contratar serviços externos.
- Monitorar gastos, tráfego de saída, armazenamento e alertas de orçamento.

**Observação:** mesmo que o GoPing seja pequeno, cada verificação gera tráfego de rede; escalar o número de monitores sem limites pode elevar custos e causar abuso.

## 13. Interface e experiência do usuário

Primeiro usar terminal e respostas JSON. Posteriormente, criar uma interface simples com:

- Lista de monitores e estados UP/DOWN/UNKNOWN.
- Última verificação e tempo de resposta.
- Histórico de falhas e gráfico de latência.
- Cadastro/edição de monitores.
- Página pública de status (fase futura).

**Escolha recomendada:** HTML templates do Go + CSS simples no início, evitando obrigatoriedade de React/Next.js. O frontend pode evoluir separadamente se houver necessidade.

## 14. Boas práticas de engenharia

- Commits pequenos, descritivos e frequentes.
- Cada versão com objetivo e critérios de aceite claros.
- Documentar decisões arquiteturais relevantes em `docs/decisions/`.
- Manter dependências externas no mínimo necessário.
- Não adicionar bibliotecas antes de entender a solução com a biblioteca padrão.
- Configurações por ambiente e exemplos em `.env.example` (sem segredos).
- Usar dados sintéticos e URLs públicas autorizadas nos exemplos.
- Registrar limitações e bugs conhecidos no README.

## 15. Checklist da primeira semana

- [ ] Instalar Go e confirmar com `go version`.
- [ ] Configurar VS Code e extensão oficial Go.
- [ ] Criar diretório `goping` e inicializar `go mod init`.
- [ ] Escrever um `main.go` que imprima uma mensagem.
- [ ] Estudar funções, variáveis e tratamento de erros.
- [ ] Fazer uma requisição HTTP com timeout e imprimir o status.
- [ ] Testar um site válido e um domínio inexistente.
- [ ] Medir a duração da requisição.
- [ ] Escrever um README curto explicando como executar.
- [ ] Fazer o primeiro commit Git.

**Definição de pronto da semana:** conseguir explicar cada linha do `main.go` e reconstruir a solução sem copiar código pronto.

## 16. Evolução para portfólio

Quando estiver implementado, o projeto poderá demonstrar:

- Go e biblioteca padrão;
- API REST, JSON, validação e persistência SQL;
- programação concorrente com goroutines e channels;
- testes automatizados e tratamento de erros;
- segurança de serviços que fazem requisições externas;
- Docker, CI/CD, deploy e observabilidade;
- capacidade de evoluir um produto real em etapas.

**Importante:** não anunciar funcionalidades como concluídas no GitHub, LinkedIn ou currículo antes de implementá-las e testá-las.

## 17. Próximas decisões

1. Nome definitivo: **GoPing** (provisório; verificar disponibilidade de nome e domínio antes de comercializar).
2. Repositório GitHub: público ou privado durante a fase inicial.
3. Licença open source: escolher antes de aceitar contribuições.
4. Primeiro lançamento: apenas CLI, sem login e sem dashboard.
5. Hospedagem: decidir após validar o MVP local e estimar o volume de verificações.

---

**Princípio do projeto:** começar pequeno, aprender profundamente, medir antes de otimizar e evoluir somente o que já funciona.
