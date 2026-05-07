# Technical Spec: Arquitetura Limpa e Independência de Nuvem

## 1. Estrutura de Camadas

* **Domain**: Contém entidades e interfaces de repositório. Proibido imports externos que não sejam da standard library ou bibliotecas de utilitários neutras.
* **Usecases**: Lógica de orquestração da aplicação. Consome interfaces definidas no domínio.
* **Infrastructure**: Implementação dos adaptadores. Aqui reside o cliente HTTP/gRPC que envia dados para o Sidecar.

## 2. Regra de Dependência

* As dependências devem apontar apenas para dentro (em direção ao Domain).
* O App Go deve ser compilável e testável sem a presença de credenciais da AWS ou Azure, utilizando apenas a URL do Sidecar injetada via variável de ambiente.
