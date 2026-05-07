# Technical Spec: Stack de Observabilidade e Telemetria

## 1. Padronização com OpenTelemetry (OTel)

* **Instrumentation**: O App Go deve ser instrumentado utilizando a SDK oficial do OpenTelemetry.
* **Context Propagation**: O Sidecar deve obrigatoriamente propagar o `trace_id` e `span_id` recebidos do App Go para as chamadas de SDK da AWS/Azure.
* **Protocolo**: Uso de OTLP (OpenTelemetry Protocol) via gRPC para exportação de dados para um coletor central (OTel Collector).

## 2. Pilares de Telemetria

* **Traces**: Rastreamento distribuído ponta-a-ponta, desde a entrada da requisição no App até a resposta do DynamoDB/CosmosDB.
* **Metrics**: Coleta de métricas de Golden Signals (Latência, Tráfego, Erros, Saturação) do runtime do Go e do consumo de recursos dos containers.
* **Logs**: Logs estruturados em JSON correlacionados com `trace_id`.
