# Hardness: Resiliência Cruzada e Paridade de Erros

## 1. Simulação de Falhas (Chaos Engineering)

* **Cloud Outage**: O Sidecar deve demonstrar comportamento de `Circuit Breaker` se o serviço de nuvem alvo retornar 5xx por mais de 500ms.
* **Network Partition**: Simular latência de 1s entre o Kubernetes e o CosmosDB/DynamoDB e validar se o timeout configurado no Sidecar é respeitado.

## 2. Tradução Unificada de Erros

* **Mapeamento de Exceções**:
  * AWS `ProvisionedThroughputExceededException` -> Erro 429 Unificado.
  * Azure `RequestRateTooLarge` -> Erro 429 Unificado.
  * Ambos -> Devem retornar um `HTTP 429` para o App Go, permitindo tratamento idêntico de backoff.

## 3. Failover de Identidade

* **Identity Resilience**: Validar se a renovação de tokens (OIDC/Workload Identity) ocorre sem interromper as chamadas ao banco.
