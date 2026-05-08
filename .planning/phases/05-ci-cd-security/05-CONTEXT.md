# Phase 5 – CI/CD & Security – Context & Decisions

## Goal
Build a fully automated, secure pipeline that builds, tests, scans, and deploys the PoC without static secrets.

## Locked Decisions (must be respected by downstream agents)

| Decision | Rationale | Result |
|---|---|---|
| **CI/CD Platform** | Simplicity and single pane of glass. | **GitHub Actions only** (no AWS CodePipeline or Azure DevOps). |
| **Secrets Management** | OIDC for CI; runtime needs cloud credentials. | **GitHub OIDC federation** for CI + **AWS Secrets Manager / Azure Key Vault** for sidecar runtime credentials. |
| **Scanning Strategy** | Need container, IaC, and dependency security. | **Trivy** (container images) + **tfsec** (Terraform) + **govulncheck** (Go dependencies). |
| **Branch Protection** | Prevent direct pushes to main. | **`main` protected** (require PR + status checks). `develop`/`feature/*` unprotected for flexibility. |
| **Integration Tests** | Fast, cheap, reproducible. | **testcontainers‑go** only (LocalStack + CosmosDB emulator). No real‑cloud ephemeral environments. |
| **OIDC Authentication** | No static credentials allowed. | GitHub OIDC provider configured for both AWS (IAM role) and Azure (federated credential). |
| **Artifact Signing** | Ensure binary integrity. | GitHub Actions signs released binaries with **cosign** (keyless OIDC). |
| **Runtime Secrets** | Sidecar needs cloud credentials at runtime. | CSI Secret Store Driver + cert‑manager (already in Phase 3) + cloud secret stores. |

## Open Questions (deferred)

- **Ephemeral PR environments** – not needed now; can be added later via Terraform workspaces.
- **Slack/Teams notifications** – out of scope for PoC.
- **Cost optimization** – GitHub Actions free tier is sufficient for PoC.

## Next Steps for Downstream Agents

1. **gsd‑phase‑researcher** – Investigate:
   - GitHub OIDC federation setup for AWS & Azure (IAM role trust policy, Azure federated credential).
   - Trivy + tfsec + govulncheck GitHub Actions examples.
   - testcontainers‑go patterns for LocalStack + CosmosDB.
   - cosign keyless signing with GitHub OIDC.

2. **gsd‑planner** – Produce concrete plan items:
   - GitHub Actions workflow: build → test → scan → deploy (OIDC auth).
   - Trivy container scan job + tfsec Terraform scan job.
   - govulncheck Go dependency scan job.
   - Branch protection settings via GitHub API.
   - Secrets Manager / Key Vault integration with CSI driver.
   - testcontainers‑go integration test job.
   - cosign signing job for releases.

3. **gsd‑security‑auditor** – Verify:
   - No static secrets in repo (grep for common patterns).
   - OIDC trust policies have no wildcards.
   - Trivy/tfsec/govulncheck scans actually run and block on findings.

These decisions are now locked; downstream agents should treat them as immutable constraints unless the user explicitly revises them.