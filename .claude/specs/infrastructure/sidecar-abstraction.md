# Technical Spec: Lógica de Abstração Multicloud do Sidecar

## 1. Responsabilidade do Contêiner Sidecar

* **Protocol Proxy**: Atuar como servidor gRPC/HTTP local (127.0.0.1) para o App Go.
* **Identity Manager**: Gerenciar a renovação de tokens OIDC/Azure AD sem expor essa complexidade ao App.
* **Cloud Selector**: Determinar o destino do tráfego (AWS ou Azure) baseado na flag `CLOUD_TARGET`.

## 2. Tradução de Modelos

* O Sidecar recebe um payload neutro e o converte para:
  * **AWS**: `map[string]types.AttributeValue` para o DynamoDB.
  * **Azure**: Documento JSON compatível com a SQL API do CosmosDB.
* Implementação obrigatória de mapeamento de tipos complexos (Datas, Enums) para garantir consistência entre os bancos.
