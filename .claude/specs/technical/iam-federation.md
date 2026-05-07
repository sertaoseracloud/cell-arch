# Infrastructure Spec: Identidade Federada e Segurança Zero Trust

## 1. AWS: IAM Roles for Service Accounts (IRSA)

* **Mecânica**: Criação de `OIDC Provider` associado ao cluster EKS. O Terraform deve criar uma `iam_role` com `assume_role_policy` restrita à ServiceAccount e Namespace específicos da aplicação.
* **Princípio de Menor Privilégio**: A role deve permitir apenas as ações `GetItem`, `PutItem` e `UpdateItem` no ARN da tabela específica.

## 2. Azure: Workload Identity

* **Mecânica**: Uso de `User-Assigned Managed Identity`. O Terraform deve estabelecer o `federated_identity_credential` entre o emissor de tokens do AKS e a identidade da Azure.
* **RBAC**: Atribuição da role `Cosmos DB Built-in Data Contributor` à Managed Identity no escopo do container do CosmosDB.
