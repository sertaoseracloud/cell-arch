# Infrastructure Spec: Topologia de Pods e Colocalização de Containers

## 1. Design de Sidecar Pattern

* **Pod Spec**: O manifesto deve definir dois containers: `main-app` (Go) e `sidecar-proxy` (Go/SDKs).
* **Comunicação Inter-Processo**: O `main-app` deve utilizar a variável de ambiente `DB_PROXY_URL=http://localhost:50051`.
* **Shared Resources**: Uso de `emptyDir` para volumes se houver necessidade de troca de arquivos de socket ou logs compartilhados.

## 2. Lifecycle e Health Checks

* **Dependência de Inicialização**: O `main-app` deve aguardar o `sidecar-proxy` estar pronto. Implementar `postStart` hooks ou lógica de retry no startup do App Go.
* **Probes**:
  * **Liveness**: O Sidecar deve reportar falha se perder conectividade com o endpoint de nuvem.
  * **Readiness**: O App Go só deve aceitar tráfego se o Sidecar estiver com o túnel de autenticação ativo.
