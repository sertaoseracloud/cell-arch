# Harness: Metas de Cobertura e Contratos de Teste

## 1. Cobertura de Código (Code Coverage)

* **Domínio e Casos de Uso**: Mínimo de 100% de cobertura. A lógica de negócio deve ser testada isoladamente de SDKs.
* **Adaptadores (Sidecar)**: Mínimo de 80% de cobertura, focando em lógica de tradução de erros e mapeamento de campos NoSQL.

## 2. Test-Driven Development (TDD) Rigor

* **Mocking**: Proibido o uso de chamadas reais de rede nos testes unitários. Interfaces de repositório devem ser mockadas via `uber-go/mock`.
* **Integration Tests**: Obrigatório o uso de `testcontainers-go`. Cada teste de integração deve rodar contra uma instância de LocalStack (DynamoDB) e CosmosDB Emulator.

## 3. Testes de Contrato

* **Paridade de Resposta**: O payload retornado pelo Sidecar-AWS deve ser estruturalmente idêntico ao Sidecar-Azure.
* **Esquema de Validação**: Uso de JSON Schema ou Protobuf para validar que o contrato entre App e Sidecar não sofreu drift.
