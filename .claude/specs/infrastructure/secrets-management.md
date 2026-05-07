# Infrastructure Spec: Gestão de Segredos e Sensíveis

## 1. Provedores de Segredos Nativos

* **AWS**: Uso do **AWS Secrets Manager**. Segredos de infraestrutura (ex: chaves de API de terceiros) devem ser armazenados aqui.
* **Azure**: Uso do **Azure Key Vault**. Armazenamento de segredos, chaves criptográficas e certificados.

## 2. Secrets Store CSI Driver

* **Abstração K8s**: Uso do **Secrets Store CSI Driver** em ambos os clusters (EKS/AKS).
* **Mecânica**: Os segredos são montados como volumes (tmpfs) nos pods. O Sidecar acessa esses arquivos via sistema de arquivos local, evitando a persistência de segredos em variáveis de ambiente.
* **Sincronização**: Habilitar `secretObjects` para que o driver sincronize os segredos do cofre nativo com `Kubernetes Secrets` apenas quando necessário para compatibilidade.
