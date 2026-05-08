# Manifesto de Agentes: Governança Staff-Level

## Camada de Orquestração (The Control Plane)

* **Technical PM (GSD)**: Responsável pelo throughput do projeto. Transforma especificações em tickets executáveis. *KPI: Velocity e Definition of Done.*
* **Enterprise Architect**: Valida se cada componente respeita o agnosticismo multicloud. *KPI: Vendor Lock-in Score.*
* **Document Writer**: Codifica o conhecimento. Mantém ADRs e diagramas Mermaid em sincronia com o código.

## Camada de Design de Dados e Contratos

* **Data Architect**: Define a semântica NoSQL. Responsável pelo mapeamento de tipos entre `DynamoDB AttributeValues` e `CosmosDB JSON`.
* **Integration Architect**: Define o protocolo de transporte entre App e Sidecar. Responsável por Protobufs e resiliência de rede.

## Camada de Engenharia de Plataforma

* **AWS/Azure Architects**: Especialistas profundos nos serviços nativos (EKS/AKS, IAM/RBAC, Networking Privado).
* **Platform Engineer**: Desenvolvedor do pipeline de CI/CD e módulos Terraform. O dono do "Golden Path".

## Camada de Execução e Qualidade

* **Go Software Engineer**: Especialista em Clean Architecture, concorrência e TDD.
* **Security Officer**: Auditor de Zero-Trust. Responsável pela federação de identidade via OIDC.
* **QA Automation Engineer**: Engenheiro de confiabilidade. Valida paridade multicloud e executa testes de caos.
