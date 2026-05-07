# Technical Spec: Padrões de Implementação em Golang

## 1. Estrutura de Projeto (Layout)

* **Standard Layout**: Uso do padrão `/cmd`, `/internal`, `/pkg`.
  * `/cmd/app`: Entrypoint da aplicação principal.
  * `/cmd/sidecar`: Entrypoint do proxy multicloud.
  * `/internal/repository`: Interfaces e implementações de persistência.
* **Dependências**: Uso estrito de Go Modules (`go.mod`). Proibido o uso de `init()` functions que escondam efeitos colaterais de conexão.

## 2. Concorrência e Contexto

* **Context Propagation**: Todas as funções de IO devem aceitar `context.Context` como primeiro parâmetro.
* **Graceful Shutdown**: Implementação obrigatória de captura de sinais (`SIGTERM`, `SIGINT`) para fechar conexões com o sidecar e bancos de dados de forma limpa.

## 3. Communication Layer (App <-> Sidecar)

* **Client HTTP/gRPC**: O App Go deve usar um `http.Client` customizado com `MaxIdleConnsPerHost` otimizado para reaproveitamento de sockets no loopback.
* **Serialization**: Preferência por `encoding/json` com `easyjson` ou `Protobuf` para minimizar o overhead de CPU durante a abstração.
