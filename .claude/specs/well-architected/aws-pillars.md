# AWS Well-Architected Framework: Estratégias de Implementação

## 1. Confiabilidade (Reliability)

* **Backoff e Retries**: Uso obrigatório do SDK v2 para Go com política de `StandardRetryer` (Exponential Backoff e Jitter) para lidar com Throttling no DynamoDB.
* **Isolamento de Falhas**: Implementação de timeouts estritos em nível de contexto (`context.WithTimeout`) para evitar o esgotamento de recursos em caso de falha regional do serviço.

## 2. Otimização de Custos (Cost Optimization)

* **Provisionamento Inteligente**: Configuração de tabelas DynamoDB em modo `On-Demand` para a PoC; transição para `Provisioned` com `Auto Scaling` apenas para cargas de trabalho previsíveis.
* **Ciclo de Vida de Dados**: Implementação de TTL (Time to Live) no DynamoDB para limpeza automática de dados temporários de teste, reduzindo custos de armazenamento.

## 3. Eficiência de Performance

* **Seleção de Região**: Os recursos devem ser provisionados em regiões com menor latência inter-cloud (ex: us-east-1 vs Azure East US).
* **DAX Capability**: Arquitetura preparada para integração futura com Amazon DynamoDB Accelerator (DAX) sem alteração no contrato do Sidecar.
