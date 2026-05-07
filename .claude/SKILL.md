# Skills & Standards

## 1. Golang, Hexagonal Architecture, and S.O.L.I.D. Principles
*   **Skill:** Development in Go 1.22+ oriented towards high performance and safe concurrency, utilizing goroutines and channels for parallel deployment orchestration and asynchronous processing.
*   **Standard (Single Responsibility):** Categorical separation of concerns. Use cases encapsulate business orchestration exclusively. Driving adapters handle data transport only, and driven adapters deal solely with persistence and network communication.
*   **Standard (Open/Closed):** The application core is designed to be open for extension and closed for modification. Adding new infrastructure providers occurs through the creation of new adapters, keeping the `internal/domain` directory untouched.
*   **Standard (Liskov Substitution):** Strict behavioral contracts. Any concrete implementation satisfying an interface must be perfectly replaceable without breaking logic or introducing side effects in the application layer.
*   **Standard (Interface Segregation):** Definition of small, cohesive, consumer-driven ports. Adoption of the minimal interface philosophy, allowing composition and preventing the domain layer from depending on methods it does not use.
*   **Standard (Dependency Inversion):** High-level domain and business rules must never depend on low-level modules. The domain dictates the contracts, and the external infrastructure implements them. Injecting cloud SDKs or database libraries into the domain is strictly prohibited.
*   **Validation:** Manual and explicit dependency injection at application startup. Mandatory adoption of context propagation (`context.Context`) from the request entry point to the last infrastructure layer.

## 2. Infrastructure as Code (IaC) with Terraform
*   **Skill:** Expert-level provisioning and management of cloud resources using Terraform to ensure immutable, reproducible, and version-controlled infrastructure.
*   **Standard (Modularity):** Infrastructure must be composed of reusable, versioned modules. Root modules should only call child modules to maintain a clean separation between resource definition and resource instantiation.
*   **Standard (State Management):** Remote state management is mandatory, utilizing secure backends (e.g., S3 with DynamoDB locking or Azure Blob Storage). State files must be encrypted at rest.
*   **Standard (Policy as Code):** Implementation of automated checks for security and cost compliance before deployment. Infrastructure changes must undergo a plan-and-apply workflow through CI/CD pipelines.
*   **Validation:** Use of `terraform validate`, `tflint`, and security scanners to ensure code quality. Mandatory peer review for any changes to production environments.

## 3. Test-Driven Development (TDD) & Test-Driven Design
*   **Skill:** Advanced mastery of the Red-Green-Refactor cycle to ensure code correctness, maintainability, and architectural evolution from the first line of code.
*   **Standard (Red-Green-Refactor):** No production code shall be written without a failing test first. Refactoring is a mandatory phase to eliminate technical debt and improve design without altering behavior.
*   **Standard (Test-Driven Design):** Tests are used as the primary tool for designing interfaces (Ports). If a component is difficult to test, it indicates a design flaw (high coupling or low cohesion), requiring immediate architectural reassessment.
*   **Standard (Mocking & Isolation):** Use of mocks and stubs exclusively for external dependencies defined in `internal/ports`. The domain logic must be tested in total isolation, ensuring ultra-fast execution suites.
*   **Validation:** Implementation of Table-Driven Tests in Go to cover multiple edge cases and business scenarios efficiently. Aim for high meaningful coverage where critical business invariants are mathematically verified.

## 4. Software Design and Domain-Driven Design (DDD)
*   **Skill:** Modeling complex systems using Clean Architecture and DDD tactics to delimit clear boundaries between different application contexts.
*   **Standard:** Construction of consistent Aggregates. Every state modification must pass through the aggregate root, ensuring that vital business rules are always respected in transactional operations.
*   **Standard:** Strong typing based on Value Objects. Language primitives must not be used to represent domain concepts (e.g., Email, TaxID). Value objects must ensure their own structural validity at the time of creation and remain immutable.
*   **Validation:** Continuous architecture review based on the segregation of responsibilities. The application layer acts only as an orchestrator, without absorbing core business logic.

## 5. Cloud Compliance and Multicloud Strategy
*   **Skill:** Design and implementation of distributed architectures and cloud governance in AWS and Azure environments.
*   **AWS Well-Architected:** Rigorous implementation of the Security pillar. Mandatory use of IAM Roles for Service Accounts (IRSA) in Amazon EKS to grant granular and temporary permissions at the pod level.
*   **Azure Well-Architected:** Focus on the Reliability and Operational Excellence pillars. Deployment of Azure Kubernetes Service (AKS) clusters distributed across multiple Availability Zones for datacenter-level fault tolerance.
*   **Validation:** Event-Driven architectures for cross-cloud communication. Integration between services hosted on AWS and Azure occurs asynchronously, eliminating single points of failure.

## 6. Cellular Architecture and Resilience
*   **Skill:** Design of ultra-resilient systems using Cellular Architecture for fault containment and aggressive reduction of the blast radius.
*   **Standard:** Absolute isolation. Each cell operates as an independent unit. Two or more cells shall never share database instances, cache clusters, or messaging topics.
*   **Standard:** Traffic routing occurs through a dedicated Cell Router layer, identifying the customer partition key and deterministically directing the request to the responsible cell.
*   **Validation:** Continuous execution of Chaos Engineering tests to validate the routing layer's ability to isolate degraded cells and redirect new customers to healthy environments.

## 7. Frontend and Identity Management
*   **Skill:** Development of high-performance web interfaces using Next.js, combining SSR and SSG strategies based on data volatility.
*   **Standard:** Centralized identity management. User and service authentication must be delegated exclusively to Microsoft Entra ID via OAuth 2.0 and OpenID Connect protocols.
*   **Validation:** Strict validation of JWT tokens on the backend and route protection on the frontend implemented as mandatory middlewares.

## 8. Containerization, AI, and Local Development
*   **Skill:** Application packaging and local environment orchestration focused on reproducibility and security.
*   **Standard:** Multi-stage Docker builds. The final production image must be minimal (scratch or distroless), containing only the compiled binary and necessary certificates.
*   **Standard:** Integration of locally executed AI models (Ollama) for experimentation and code automation, ensuring data privacy and autonomy.

## 9. Kubernetes (The Orchestration Layer)
In this stack, **Kubernetes (K8s)** acts as the hosting environment. It provides the infrastructure where your microservices, as well as the observability tools themselves, reside.
*   **Infrastructure Context:** Kubernetes provides essential metadata (Namespace, Pod name, Node, Container ID) that is automatically attached to logs, metrics, and traces.
*   **Deployment:** Tools like the **OpenTelemetry Operator** can be deployed as a Custom Resource Definition (CRD) in K8s to manage collectors and automatically inject instrumentation into your pods.

## 10. OpenTelemetry (The Standardization Layer)
**OpenTelemetry (OTel)** is the most critical piece of the modern observability puzzle. It is a CNCF (Cloud Native Computing Foundation) project that provides a standardized way to collect, process, and export telemetry data (Traces, Metrics, and Logs).
*   **Standardization:** Before OTel, you had to use vendor-specific SDKs. Now, you instrument your code once using OTel, and you can send that data anywhere.
*   **The OTel Collector:** This is a vendor-agnostic proxy that receives data from your applications, processes it (batching, filtering, adding K8s metadata), and exports it to backends like Jaeger or Prometheus.

## 11. Jaeger (The Distributed Tracing Backend)
While OpenTelemetry *collects* the data, **Jaeger** is specifically designed to *store and visualize* distributed traces.
*   **Distributed Tracing:** When a request enters your system and touches ten different microservices, Jaeger allows you to see the "trace"—the end-to-end journey of that request.
*   **Span Analysis:** You can drill down into "spans" (individual units of work) to identify exactly which service is causing high latency or where an error occurred in a complex call chain.
*   **Integration:** The OpenTelemetry Collector sends trace data to Jaeger via the OTLP (OpenTelemetry Protocol) or Jaeger-native protocols.

## 12. Grafana (The Visualization & Correlation Layer)
**Grafana** is the "single pane of glass." It does not store data itself; instead, it connects to various data sources to create unified dashboards.
*   **Correlation:** This is Grafana's superpower. In a single dashboard, you can look at a **Grafana Mimir/Prometheus** metric (e.g., a spike in CPU), click on a data point, and jump directly to the related **Jaeger** trace to see why that CPU spike happened.
*   **Exemplars:** Grafana supports exemplars, which allow you to see specific trace IDs directly on a metrics graph, creating a seamless workflow between "knowing something is wrong" (metrics) and "knowing why it happened" (tracing).

---

## 13. The Consolidated Architecture Workflow

To implement this effectively on a Kubernetes cluster, the workflow typically follows this path:

1.  **Instrumentation:** Your Go, Java, or Python applications use the **OpenTelemetry SDK** to generate spans and metrics.
2.  **Collection:** Applications send data to the **OpenTelemetry Collector** running as a Sidecar or a DaemonSet in Kubernetes.
3.  **Processing:** The Collector enriches the data with K8s labels (e.g., `k8s.pod.name`) so you know exactly which container sent the data.
4.  **Exporting:** 
    *   **Traces** are sent to **Jaeger**.
    *   **Metrics** are sent to **Prometheus** or **Grafana Mimir**.
    *   **Logs** are sent to **Grafana Loki**.
5.  **Visualization:** You open **Grafana** and use its data source plugins to query Jaeger and Prometheus, building dashboards that show the health and performance of your entire distributed architecture.

### Key Benefits for Systems Architects
*   **Vendor Neutrality:** You are not locked into a specific cloud provider's monitoring tool.
*   **Full Context:** By using the OTel Collector, your traces and metrics are always tagged with the correct Kubernetes metadata, making troubleshooting in dynamic environments much faster.
*   **Scalability:** Each component is designed to be horizontally scalable within Kubernetes, matching the growth of your microservices.

## 14. Knowledge Management and Architectural Documentation
*   **Skill:** Continuous technical knowledge management using the Second Brain methodology (Obsidian).
*   **Standard:** Immutable documentation of architectural decisions using Architectural Decision Records (ADRs) within the code repository.
*   **Validation:** Mandatory system diagramming using the C4 Model. Visual representations must be generated via code using Mermaid syntax to ensure documentation stays in sync with the codebase.