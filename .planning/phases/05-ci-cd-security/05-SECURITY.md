---
phase: 05
slug: ci-cd-security
status: planned
threats_open: 7
asvs_level: 1
created: 2026-05-07
---

# Phase 05 — Security (Planned)

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| GitHub Actions → AWS/Azure | OIDC token exchange for cloud access | Cloud credentials (temporary) |
| CI Runner → Container Registry | Pushing images | Container images |
| Secrets Manager/Key Vault → Sidecar | Runtime credential injection | Cloud secrets |
| Trivy/Tfsec/Govulncheck → CI | Security scan results | Vulnerability reports |
| Cosign → Registry | Binary/image signing | Signed artifacts |

---

## Threat Register (Planned)

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-05-01 | Spoofing | GitHub OIDC | plan | Configure federated identity with audience validation | planned |
| T-05-02 | Tampering | CI pipeline | plan | Branch protection on `main`; require PR + status checks | planned |
| T-05-03 | Information Disclosure | Secrets in CI | plan | No static secrets; use OIDC + Secrets Manager/Key Vault | planned |
| T-05-04 | Information Disclosure | Trivy scans | plan | Fail pipeline on HIGH/CRITICAL findings | planned |
| T-05-05 | Denial of Service | CI resource exhaustion | plan | Limit concurrent jobs; timeout on long-running steps | planned |
| T-05-06 | Elevation of Privilege | cosign signing | plan | Keyless signing with OIDC; verify signatures in deployment | planned |
| T-05-07 | Tampering | Terraform state | plan | Remote state with encryption; restrict access via RBAC | planned |

---

## Accepted Risks Log

_To be determined during execution._

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| _pending execution_ | 7 | 0 | 7 | _pending_ |

---

## Sign-Off

- [ ] All threats have a disposition
- [ ] Accepted risks documented
- [ ] `threats_open: 0` confirmed
- [ ] `status: verified` set

**Approval:** _pending execution_
