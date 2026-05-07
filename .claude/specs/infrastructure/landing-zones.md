# Infrastructure Spec: Design de Landing Zones Multicloud

## 1. Landing Zone AWS (Control Tower & Organizations)

* **Estrutura de OUs**: Divisão entre `Core` (Shared Services, Logging, Security) e `Workloads` (PoC-Dev, PoC-Prod).
* **Networking (Hub & Spoke)**:
  * **Hub VPC**: Centraliza Inspeção de Tráfego e NAT Gateways.
  * **Spoke VPC**: Hospeda o cluster EKS. Comunicação com DynamoDB via **Gateway Endpoints** para evitar tráfego via Internet.
* **Guardrails**: Implementação de SCPs (Service Control Policies) que impedem a desativação do CloudTrail e a criação de recursos fora das regiões permitidas.

## 2. Landing Zone Azure (Cloud Adoption Framework - CAF)

* **Management Groups**: Hierarquia baseada em `Org -> Platform -> Workloads`.
* **Connectivity (VNet Hub-Spoke)**:
  * **Hub VNet**: Contém o Azure Firewall e conexões de VPN/ExpressRoute.
  * **Spoke VNet**: Hospeda o AKS. Uso de **Azure Private Link** para o CosmosDB, garantindo que o banco não tenha IP público.
* **Azure Policy**: Aplicação de Blueprints que garantem que todos os discos sejam criptografados com CMK (Customer Managed Keys).

## 3. Paridade e Identidade Centralizada

* **SSO/IdP**: Integração do AWS IAM Identity Center com o Azure AD (Microsoft Entra ID) para garantir que o desenvolvedor tenha o mesmo nível de acesso em ambos os consoles.
* **Log Aggregation**: Sincronização de logs do CloudWatch e Azure Monitor para um SIEM central ou um bucket S3/Blob dedicado para auditoria.
