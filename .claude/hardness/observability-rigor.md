# Hardness: Rigor de Telemetria e Diagnóstico

## 1. Validação de Rastreabilidade

* **Trace Gap Analysis**: O deploy deve ser considerado falho se houver "buracos" no trace (ex: o App Go inicia o trace, mas o Sidecar não anexa sua operação).
* **Sampling Policy**: Em ambiente de PoC, o `sampling_rate` deve ser de 100% para validar todos os cenários de falha.

## 2. Alerting Hardness

* **False Positive Test**: Simular falhas controladas (Chaos Engineering) e validar se os alertas são disparados nos canais corretos (Slack/PagerDuty) em menos de 2 minutos.
* **Log Retention**: Garantir que logs de erro críticos tenham retenção mínima de 30 dias para análise pós-morte (Post-mortem).
