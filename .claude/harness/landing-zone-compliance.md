# Harness: Validação e Auditoria de Landing Zones

## 1. Isolamento de Rede

* **Reachability Test**: O Sidecar não deve conseguir alcançar a internet pública diretamente; todo tráfego deve passar pelos endpoints privados. O deploy deve falhar se o `terraform-compliance` detectar rotas diretas para `0.0.0.0/0`.
* **Egress Control**: Validação de que apenas os FQDNs dos serviços de nuvem necessários (DynamoDB/CosmosDB) estão liberados no Firewall.

## 2. Governança de Custos

* **Budget Alarms**: O Terraform deve provisionar alarmes de custo que notificam em 50%, 80% e 100% do orçamento da PoC em ambas as nuvens.
* **Tagging Enforcement**: Qualquer recurso criado sem as tags `project` e `environment` deve ser automaticamente marcado para deleção por scripts de governança.

## 3. Segurança de Identidade

* **MFA Enforcement**: Auditoria de que nenhum usuário administrativo acessa a Landing Zone sem Multi-Factor Authentication ativo.
