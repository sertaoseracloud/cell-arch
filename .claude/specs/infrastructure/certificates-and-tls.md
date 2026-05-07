# Infrastructure Spec: Gestão de Certificados e TLS

## 1. Cert-Manager e Emissores

* **Provisionamento**: Instalação do `cert-manager` via Terraform nos dois clusters.
* **AWS Integration**: Configuração de `ClusterIssuer` utilizando **AWS Private CA** ou **Let's Encrypt** via DNS01 Challenge com Route53.
* **Azure Integration**: Configuração de `ClusterIssuer` integrado ao **Azure Key Vault** para gerenciar o ciclo de vida dos certificados via DNS01 com Azure DNS.

## 2. TLS Inter-Pod e mTLS

* **Loopback Security**: A comunicação entre o App Go e o Sidecar (localhost) não exige TLS, mas a saída do Sidecar para os bancos de nuvem deve obrigatoriamente utilizar **TLS 1.2+**.
* **Ingress TLS**: O Ingress Controller (Nginx ou ALB/Application Gateway) deve gerenciar a terminação TLS utilizando certificados providenciados automaticamente pelo cert-manager.
