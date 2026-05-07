# Hardness: Metas de Performance e Eficiência

## 1. Latência do Sidecar (Overhead)

* **Hop de Rede**: O salto adicional (App -> Sidecar -> Nuvem) não deve exceder 5ms de latência adicional (p95) em relação a uma chamada direta via SDK.
* **Serialização**: O tempo de marshalling/unmarshalling entre JSON/Protobuf no sidecar deve ser inferior a 1ms.

## 2. Recursos e Footprint

* **CPU/Memória**: O Sidecar (escrito em Go) deve manter um footprint estável de < 64MB RAM em carga máxima (500 TPS).
* **Connection Pooling**: Reutilização de conexões TCP deve ser superior a 95% para evitar exaustão de sockets.

## 3. Benchmarking

* **Throughput**: A abstração deve suportar 1000 leituras/segundo e 500 escritas/segundo mantendo o p99 abaixo de 30ms.
