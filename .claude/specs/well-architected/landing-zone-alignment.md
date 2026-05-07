# Well-Architected: Alinhamento de Landing Zones

## 1. Segurança (Security)

* **Princípio do Menor Privilégio**: As Landing Zones devem implementar o isolamento total entre o plano de controle do Kubernetes e o plano de dados dos bancos NoSQL.
* **Data Residency**: Garantir que os dados nunca saiam das regiões geográficas especificadas via políticas de conformidade.

## 2. Excelência Operacional

* **IaC Reprodutível**: A Landing Zone deve ser inteiramente provisionada via Terraform/Terragrunt. Nenhuma alteração de rede ou política de segurança deve ser feita via clique.
* **Self-Healing Infrastructure**: Configuração de logs que disparam funções (Lambda/Azure Functions) para remediar automaticamente configurações fora de conformidade (drift remediation).
