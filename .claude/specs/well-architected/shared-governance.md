# Governança Compartilhada Multicloud

## 1. Observabilidade Unificada

* **Trace ID**: Propagação obrigatória do `X-Ray Trace ID` (AWS) e `Correlation ID` (Azure) através do App Go e Sidecar para rastreamento de requisições inter-cloud.
* **Logging Estruturado**: Todos os logs devem estar em formato JSON, contendo os campos `cloud_provider`, `region`, `service_name` e `request_id`.

## 2. Gestão de Mudanças

* **Infrastructure as Code (IaC)**: Todas as alterações de Well-Architected (ex: mudança de nível de consistência no CosmosDB) devem ser aprovadas via PR no Terraform antes da execução.
* **Policy Enforcement**: Uso de `Azure Policy` e `AWS Config` para impedir a criação de recursos que não atendam aos requisitos de criptografia e rede definidos.
