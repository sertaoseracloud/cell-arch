# Technical Spec: TDD e Testes Idiomáticos em Go

## 1. Unit Testing

* **Table-Driven Tests**: Uso obrigatório de testes baseados em tabelas para validar casos de borda na lógica de domínio.
* **Interfaces & Mocks**: Uso de `gomock` ou `moq` para gerar simuladores das interfaces de infraestrutura.
  * Ex: `type Database interface { Save(ctx, data) error }`.

## 2. Integration Testing (Testcontainers-go)

* **Container Lifecycle**: Os testes de integração devem iniciar containers Docker temporários via `testcontainers-go`.
* **Parallelism**: Uso de `t.Parallel()` para garantir que a suíte de testes seja rápida, isolando os recursos de container por namespace de teste.

## 3. Assertions

* **Standard Library**: Preferência por verificações manuais `if got != want`. Uso de `stretchr/testify` permitido apenas para comparações complexas de structs/slices.
