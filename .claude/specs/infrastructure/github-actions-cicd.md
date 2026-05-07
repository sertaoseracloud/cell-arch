# Infrastructure Spec: CI/CD com GitHub Actions

## 1. Workflow de Continuous Integration (CI)

* **Linter & Formatting**: Execução de `golangci-lint` e `terraform fmt -check`.
* **Automated Testing**: Execução de `go test ./...` utilizando `testcontainers-go`.
* **Security Scan**: Integração com `Trivy` para scan de vulnerabilidades em imagens Docker e `tfsec` para o Terraform.

## 2. Workflow de Continuous Deployment (CD)

* **Multicloud Deploy**: O GitHub Actions deve disparar o `terraform apply` simultaneamente para AWS e Azure após o merge na `main`.
* **OIDC Authentication**: Uso de **GitHub Actions OIDC** para autenticação na AWS e Azure, eliminando a necessidade de `secrets` estáticos (Access Keys) no repositório.
* **ArgoCD/Flux Integration**: (Opcional) O workflow pode atualizar os manifestos em um repositório de GitOps para sincronização nos clusters EKS/AKS.
