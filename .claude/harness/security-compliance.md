# Harness: Auditoria de Segredos e Segurança Criptográfica

## 1. Rotação e Expiração

* **Automatic Rotation**: A PoC deve demonstrar que o Sidecar consegue recarregar segredos (ex: renovação de token) sem a necessidade de reinicializar o Pod (uso de `inotify` ou TTL baixo).
* **Certificate Expiry**: Alertas devem ser disparados se um certificado tiver menos de 30 dias de validade.

## 2. Varredura de Vulnerabilidades

* **Secret Leakage**: Uso de ferramentas como `trufflehog` ou `git-secrets` no pipeline para garantir que nenhum segredo do Terraform ou código Go foi commitado.
* **Encryption at Rest**: O Terraform deve validar que as chaves de criptografia (KMS na AWS e Customer Managed Keys na Azure) estão ativas nos bancos NoSQL.
