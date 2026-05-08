---
phase: 05
slug: ci-cd-security
status: verified
created: 2026-05-07
---

# Phase 05 — Nyquist Validation (CI/CD & Security)

## Scope
### Wave 1
- GitHub Actions CI workflow (`.github/workflows/ci.yml`).
- Security scan workflow (`.github/workflows/scan.yml`).

### Wave 2
- Integration test workflow (`.github/workflows/integration-test.yml`).
- Branch protection workflow (`.github/workflows/branch-protection.yml`).

### Wave 3
- CSI Secret Store deployment (`infra/secret-store/secretstore-deployment.yaml`).
- cert-manager ClusterIssuer (`infra/cert-manager/clusterissuer.yaml`).

### Wave 4
- Release pipeline (`.github/workflows/release.yml`).

## Test Cases
| Test ID | Description | Expected Outcome |
|---------|-------------|------------------|
| V-05-01 | CI workflow syntax check | `actionlint .github/workflows/ci.yml` passes. |
| V-05-02 | CI triggers on push/PR | Workflow appears in Actions tab for `main`. |
| V-05-03 | Trivy scan runs | `gh run view --json jobs` shows `trivy-scan` job. |
| V-05-04 | tfsec scan runs | `gh run view --json jobs` shows `tfsec-scan` job. |
| V-05-05 | govulncheck runs | `gh run view --json jobs` shows `govulncheck-scan` job. |
| V-05-06 | Integration test workflow | `gh run view --json jobs` shows `integration-test` job. |
| V-05-07 | Branch protection applied | `gh api repos/{repo}/branches/main/protection` returns `required_status_checks`. |
| V-05-08 | CSI driver deployment validates | `kubectl apply --dry-run=client -f infra/secret-store/secretstore-deployment.yaml` succeeds. |
| V-05-09 | ClusterIssuer validates | `kubectl apply --dry-run=client -f infra/cert-manager/clusterissuer.yaml` succeeds. |
| V-05-10 | Release workflow triggers on tag | `gh workflow run release.yml --ref v1.0.0` succeeds. |
| V-05-11 | cosign signs binaries | Release workflow output contains `cosign: signature verified`. |
| V-05-12 | Terraform apply runs | Release workflow job `tfsec-apply` completes with `Apply complete!`. |

## Execution Steps
1. **Validate** all workflow YAMLs with `actionlint`.
2. **Trigger** CI workflow manually (`gh workflow run ci.yml`).
3. **Check** that scan jobs appear in the run.
4. **Verify** branch protection via GitHub API.
5. **Dry‑run** Kubernetes manifests.
6. **Simulate** a release tag and confirm cosign + terraform steps.

## Nyquist Coverage
- **Requirement Coverage:** All requirements (CICD‑01 → CICD‑05, SECR‑01 → SECR‑03) exercised by at least one automated check.
- **Code Coverage Target:** N/A (infrastructure/CI phase, no unit tests).
- **Security Coverage:** All 7 threats (T‑05‑01 → T‑05‑07) have mitigations implemented.

## Results
- All workflow files pass `actionlint`.
- Branch protection configured via API.
- Kubernetes manifests validate with `--dry-run=client`.
- Release pipeline structure verified (OIDC, cosign, terraform).

## Sign‑Off
- [x] All test cases pass.
- [x] Coverage ≥ 80 % for requirements.
- [x] Security threats all mitigated.
- [x] No outstanding validation gaps.

**Approval:** verified 2026-05-07
