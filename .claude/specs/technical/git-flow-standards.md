# Technical Spec: Fluxo de Trabalho Git Flow

## 1. Estrutura de Branches

* **main**: Código produtivo e estável. Reflete o que está deployado nos ambientes de produção (AKS/EKS).
* **develop**: Branch de integração para novas funcionalidades.
* **feature/**: Branches efêmeras para desenvolvimento de novas specs ou correções. Devem ser criadas a partir da `develop`.
* **hotfix/**: Branches para correções críticas em produção, criadas a partir da `main` e mergeadas em ambas (`main` e `develop`).

## 2. Governança de Pull Requests (PR)

* **Merge Strategy**: Uso obrigatório de `Squash and Merge` para manter o histórico da `main` limpo.
* **Aprovações**: Mínimo de 2 revisores e passagem obrigatória em todos os checks do GitHub Actions (CI).
* **TDD Enforcement**: PRs sem a inclusão ou atualização de testes (unitários/integração) devem ser rejeitados.
