---
phase: 05
slug: ci-cd-security
created: 2026-05-07
---

# Phase 05 — Research (CI/CD & Security)

## Standard Stack
- **GitHub Actions** – primary CI/CD platform (no AWS CodePipeline/Azure DevOps).
- **GitHub OIDC** (`actions/id-token`) – federated authentication to AWS (IAM role) and Azure (federated credential).
- **Trivy** (`aquasecurity/trivy-action`) – container image vulnerability scanner.
- **tfsec** (`aquasecurity/tfsec-action`) – Terraform static analysis.
- **govulncheck** (`golang/govulncheck-action`) – Go dependency vulnerability scanner.
- **cosign** (`sigstore/cosign-installer`) – keyless signing of binaries/images with OIDC.
- **AWS Secrets Manager** / **Azure Key Vault** – runtime secrets for sidecar (via CSI Secret Store Driver + cert-manager, already in Phase 3).
- **testcontainers-go** – integration tests (LocalStack + CosmosDB emulator).

## Architecture Patterns
1. **OIDC Federation** – GitHub Actions assume IAM role (AWS) / federated credential (Azure) via `id-token` + `configure-aws-credentials` / `azure/login`.
2. **Pipeline Stages** – `build` → `test` → `scan` → `deploy` (block on scan findings).
3. **Branch Protection** – GitHub API (`octokit/request-action`) to enforce required PR reviews + status checks on `main`.
4. **Artifact Signing** – cosign sign images/binaries after build; verify on deployment.
5. **Secrets Injection** – CSI Secret Store Driver mounts secrets from AWS/Azure into sidecar pod at runtime.

## Don't Hand-Roll
- **OIDC Token Exchange** – use `actions/id-token` + cloud provider's official login action; never manually exchange tokens.
- **Container Scanning** – use Trivy action; don't parse image layers manually.
- **Terraform Scanning** – use tfsec action; don't write custom HCL linters.
- **Go Vulnerability Check** – use govulncheck; don't manually track CVE databases.
- **Binary Signing** – use cosign keyless OIDC; don't manage signing keys manually.

## Common Pitfalls
- **OIDC Audience Mismatch** – GitHub's default audience (`api://AzureADTokenExchange`) may not match Azure's expected audience; set `audience` explicitly.
- **Trivy Exit Code** – Trivy fails on HIGH/CRITICAL by default; ensure pipeline fails appropriately.
- **govulncheck False Positives** – may flag vulnerabilities in test dependencies; configure `govulncheck.yml` to exclude test packages.
- **cosign Key Management** – keyless OIDC requires correct `COSIGN_EXPERIMENTAL=1` env var in older versions.
- **Secrets Manager RBAC** – ensure the OIDC-assumed role has `secretsmanager:GetSecretValue` permission only (least privilege).

## Code Examples
```yaml
# GitHub OIDC to AWS
- name: Configure AWS Credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: ${{ secrets.AWS_IAM_ROLE_ARN }}
    role-session-name: GitHubActions
    aws-region: us-east-1

# Trivy scan
- name: Run Trivy
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'ghcr.io/${{ github.repository }}/cell-arch:${{ github.sha }}'
    format: 'table'
    exit-code: '1'
    severity: 'HIGH,CRITICAL'

# cosign sign
- name: Sign image
  run: cosign sign --oidc-provider=github-actions ghcr.io/${{ github.repository }}/cell-arch:${{ github.sha }}'
```

## Sources
- GitHub OIDC documentation (2026-05-07)
- Trivy GitHub Action README (2026-05-06)
- tfsec documentation (2026-05-06)
- govulncheck documentation (2026-05-05)
- cosign keyless signing guide (2026-05-07)
- AWS Secrets Manager CSI driver docs (2026-05-07)

---
*Research complete. This file will be consumed by `/gsd-plan-phase 5`.*