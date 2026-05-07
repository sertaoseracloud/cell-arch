# Technical Spec: Ciclo de Desenvolvimento TDD em Go

## 1. Fluxo Red-Green-Refactor

* **Fase Red**: Escrita de testes unitários em `domain_test.go` definindo o comportamento esperado da persistência. O teste deve falhar por falta de implementação da interface.
* **Fase Green**: Implementação mínima do adaptador que aponta para o localhost do Sidecar. Uso de `testcontainers-go` para validar a integração real com emuladores (LocalStack/Cosmos Emulator).
* **Fase Refactor**: Otimização do código mantendo a passagem dos testes. Proibido acoplamento de lógica de negócio com tipos específicos dos SDKs de nuvem.

## 2. Estratégia de Mocking

* **Interface-Driven**: Toda interação com o banco deve ser feita via interfaces (ex: `type Repository interface`).
* **Geradores**: Uso de `mockgen` para criar stubs de interfaces de domínio, permitindo testes de lógica sem dependência de rede.
