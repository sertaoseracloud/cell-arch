# Microsoft Azure Well-Architected Framework: Diretrizes Operacionais

## 1. Segurança (Security)

* **RBAC e Identity**: Uso de `Azure AD Workload Identity` para acesso ao CosmosDB. Proibição de chaves compartilhadas (Primary Keys) no Sidecar.
* **Network Security**: Implementação de `Service Endpoints` ou `Private Links` para garantir que o tráfego do AKS para o CosmosDB não transite pela internet pública.

## 2. Excelência Operacional (Operational Excellence)

* **Monitoramento Nativo**: Exportação de métricas via `OpenTelemetry` para o `Azure Monitor` e `Application Insights`.
* **Deployment Confiável**: Uso de `Health Probes` (Liveness/Readiness) que validam não apenas o container, mas a conectividade do Sidecar com o endpoint do CosmosDB.

## 3. Confiabilidade e Escopo

* **Consistência de Dados**: Configuração do CosmosDB com nível de consistência `Session` ou `Bounded Staleness` para equilibrar latência e integridade de dados.
* **Zonal Redundancy**: Configuração de `Availability Zones` no AKS e redundância de zona no CosmosDB para mitigar falhas de datacenter único.
