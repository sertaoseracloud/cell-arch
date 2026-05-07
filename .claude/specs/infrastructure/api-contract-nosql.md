# Technical Spec: Contrato de API e Esquema Unificado

## 1. Definição do Contrato (gRPC/Protobuf)

* O contrato deve definir operações atômicas: `PutItem`, `GetItem`, `DeleteItem`, `QueryByPartition`.
* Campos obrigatórios em cada payload: `partition_key`, `sort_key` (opcional), `data_payload` (JSON), `metadata`.

## 2. Paridade de Dados

* **Partition Key**: Deve ser mapeada para o HASH Key no DynamoDB e para o Partition Key Path no CosmosDB.
* **Timestamps**: Uso obrigatório de ISO 8601 em UTC para garantir que consultas baseadas em tempo funcionem de forma idêntica em ambos os provedores.
* **Consistency Level**: O contrato deve permitir que o App solicite `Strong` ou `Eventual` consistency, e o Sidecar deve traduzir isso para os parâmetros equivalentes de cada SDK.
