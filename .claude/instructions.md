### `.claude/instructions.md`

Este documento estabelece as diretrizes obrigatórias e inegociáveis para o desenvolvimento da PoC Multicloud (AWS/Azure). Qualquer código, manifesto de infraestrutura ou automação gerado deve estar em conformidade estrita com estas regras.

## 1. Princípios de Engenharia e Golang

* **Idiomatic Go**: Aplique os padrões do `Effective Go`. Utilize binários estáticos, tipagem forte e tratamento de erro explícito com `fmt.Errorf("context: %w", err)`.
* **Clean Architecture**: Mantenha separação rigorosa entre as camadas. O domínio em `/internal/domain` deve ser agnóstico e sem imports de SDKs externos.
* **Sidecar Pattern (Agnosticismo)**: A aplicação principal (`/cmd/app`) nunca deve conhecer o provedor de nuvem. Toda a lógica de tradução para DynamoDB ou CosmosDB reside no `/cmd/sidecar`, comunicando-se via `localhost:50051`.
* **Injeção de Dependência**: Proibido o uso de `init()` ou globais para conexões. Use construtores para facilitar o TDD.

## 2. Ciclo de TDD e Qualidade de Software

* **TDD Obrigatório**: Inicie qualquer funcionalidade pelo teste de contrato ou unitário. Siga o fluxo Red-Green-Refactor conforme `.claude/specs/technical/tdd-lifecycle-go.md`.
* **Testcontainers-go**: Testes de integração devem validar a paridade entre LocalStack (AWS) e Cosmos Emulator (Azure).
* **Hardness Compliance**: O código só é válido se passar nos critérios de cobertura (thresholds) e ]nos cenários de falha definidos em `.claude/hardness/`.

## 3. Infraestrutura e Landing Zones

* **Terraform Simétrico**: Utilize módulos reutilizáveis e padronizados. O estado deve ser protegido por State Locks (S3/DynamoDB e Blob/Lease).
* **Identidade Zero-Trust**: Proibido o uso de segredos estáticos (Access Keys). Utilize **IRSA** na AWS e **Workload Identity** na Azure para acesso a recursos.
* **Landing Zone Rigor**: Clusters Kubernetes e bancos de dados devem operar em sub-redes privadas, utilizando **Private Links/Endpoints** para comunicação, sem exposição à internet pública.

## 4. Gestão de Segredos, Certificados e Observabilidade

* **Secrets & Certs**: Utilize **Secrets Store CSI Driver** para montagem de volumes de segredos e **cert-manager** para automação de TLS via DNS01 Challenge.
* **Observabilidade OTel**: Implemente instrumentação nativa com **OpenTelemetry** para Traces, Metrics e Logs correlacionados via `trace_id`.
* **Trace Propagation**: O Sidecar deve obrigatoriamente propagar o contexto de rastreamento do App Go para os serviços de nuvem.

## 5. Fluxo de Trabalho (GitHub & Git Flow)

### Commitlint e GitFlow
- Instale `commitlint` e `husky` como dev dependencies (`npm i -D @commitlint/{config-conventional,cli} husky`).
- O hook `commit-msg` garantirá que todas as mensagens sigam o padrão Conventional Commits, adequado ao fluxo Gitflow (feature/, hotfix/, release/).
- O autor dos commits deve ser **sertaoseracloud <engcfraposo@gmail.com>**.
- Antes de cada merge, execute `git flow release start <versão>` e siga o ciclo de release.



* **Git Flow**: Nenhuma alteração é permitida diretamente em `main` ou `develop`. Utilize branches `feature/`, `hotfix/` ou `release/`.
* **GitHub Actions CI/CD**: O pipeline deve realizar scans de segurança (`Trivy`/`tfsec`) e testes de integração antes de qualquer deploy.
* **OIDC no Pipeline**: O GitHub Actions deve autenticar-se na AWS e Azure via OIDC, eliminando segredos estáticos no repositório.

## 6. Well-Architected Framework (WAF)

* **Confiabilidade**: Implemente obrigatoriamente `Exponential Backoff`, `Circuit Breaker` e `Timeout` em todas as chamadas de rede.
* **Excelência Operacional**: Todo recurso deve possuir tags de identificação e logs estruturados em JSON para auditoria.

## 7. Protocolo de Validação de Resposta

Antes de fornecer qualquer solução, valide internamente:

1. A lógica de negócio está isolada de SDKs de nuvem?
2. A autenticação sugerida utiliza identidade federada (OIDC/Workload)?
3. O código Terraform segue a estrutura de módulos e Landing Zone?
4. Existe um plano de testes baseado em TDD e Testcontainers?
5. A comunicação App/Sidecar está documentada e segura?

---

> **Atenção**: Se houver conflito entre a facilidade de implementação e o rigor das especificações em `.claude/specs/` ou `.claude/hardness/`, o rigor deve prevalecer. Esta PoC é um projeto de alta densidade técnica e deve refletir práticas de nível Staff Engineer.
