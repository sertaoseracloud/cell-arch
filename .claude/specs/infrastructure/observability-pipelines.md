# Infrastructure Spec: Pipelines de Dados e Exportação

## 1. Coletor Central (OTel Collector)

* **Deployment**: Rodar como um `DaemonSet` ou um `Sidecar` adicional no Kubernetes para agregação local antes do envio para a nuvem.
* **Multi-Backend Export**: Configuração do coletor para enviar dados simultaneamente para:
  * **AWS**: Amazon Managed Service for Prometheus e AWS X-Ray.
  * **Azure**: Azure Monitor (Application Insights) e Log Analytics.

## 2. Dashboarding e Alerta

* **Grafana**: Provisionamento via Terraform de dashboards agnósticos que consolidam métricas de ambos os provedores.
* **SLIs/SLOs**: Definição de Service Level Indicators baseados na latência do Sidecar e taxa de erro dos bancos NoSQL.
