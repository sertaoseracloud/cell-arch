# Well-Architected: Excelência Operacional via Git

## 1. Versionamento de Infraestrutura

* **Immutable Releases**: Cada versão da infraestrutura deve ser tageada no GitHub, permitindo a reconstrução exata da Landing Zone de qualquer ponto no tempo.
* **Documentation as Code**: O README e as specs no diretório `.claude/` devem ser mantidos atualizados em cada PR, garantindo que a documentação técnica nunca esteja obsoleta.

## 2. Monitoramento de Pipeline

* **DORA Metrics**: Coleta de métricas de pipeline (Deployment Frequency, Lead Time for Changes, Change Failure Rate) para avaliar a saúde do fluxo de entrega.
