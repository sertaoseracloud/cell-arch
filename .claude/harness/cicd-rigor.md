# Harness: Rigor de Automação e Pipeline

## 1. Pipeline Failure Policy

* **Stop-the-line**: Qualquer falha em testes unitários ou violação de segurança (linter/scan) deve impedir o merge.
* **Artifact Integrity**: Imagens Docker devem ser assinadas (ex: utilizando `Cosign`) antes de serem enviadas para o ECR (AWS) ou ACR (Azure).

## 2. Deployment Safety

* **Blue/Green ou Canary**: O pipeline deve suportar estratégias de deploy progressivo para mitigar riscos durante a atualização do Sidecar ou do App Go.
* **Rollback Automatizado**: Em caso de falha nos `Health Checks` pós-deploy, o GitHub Actions deve disparar o rollback da versão anterior estável.
