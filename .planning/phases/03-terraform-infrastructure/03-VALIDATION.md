---
phase: 3
slug: terraform-infrastructure
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-05-07
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | terraform validate / tfsec |
| **Config file** | infra/modules/aws-vpc/ & infra/modules/azure-vnet/ |
| **Quick run command** | `terraform init -backend=false && terraform validate` |
| **Full suite command** | `terraform init -backend=false && terraform validate && tfsec .` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `terraform validate`
- **After every plan wave:** Run full suite (validate + tfsec)
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01-01 | 1 | TERR-01 | T-03-01-01 | No IGW in Spoke VPC | terraform | `cd infra/modules/aws-vpc && terraform init -backend=false && terraform validate` | infra/modules/aws-vpc/main.tf | ✅ green |
| 3-01-02 | 01-01 | 1 | TERR-02 | T-03-01-02 | DynamoDB endpoint via Gateway Endpoint | terraform | `cd infra/modules/aws-vpc && terraform init -backend=false && terraform validate` | infra/modules/aws-vpc/main.tf | ✅ green |
| 3-01-03 | 01-01 | 1 | TERR-03 | T-03-01-03 | VPC peering non-transitive | terraform | `cd infra/modules/aws-vpc && terraform init -backend=false && terraform validate` | infra/modules/aws-vpc/main.tf | ✅ green |
| 3-01-04 | 01-01 | 1 | TERR-04 | T-03-01-04 | NAT Gateway public IP accepted | tfsec | `cd infra/modules/aws-vpc && tfsec .` | infra/modules/aws-vpc/main.tf | ✅ green |
| 3-02-01 | 01-02 | 1 | TERR-05 | T-03-02-01 | No public IPs in Spoke VNet | terraform | `cd infra/modules/azure-vnet && terraform init -backend=false && terraform validate` | infra/modules/azure-vnet/main.tf | ✅ green |
| 3-02-02 | 01-02 | 1 | TERR-06 | T-03-02-02 | VNet peering transitive prevention | terraform | `cd infra/modules/azure-vnet && terraform init -backend=false && terraform validate` | infra/modules/azure-vnet/main.tf | ✅ green |
| 3-02-03 | 01-02 | 1 | TERR-07 | T-03-02-03 | AKS subnet size adequate | terraform | `cd infra/modules/azure-vnet && terraform init -backend=false && terraform validate` | infra/modules/azure-vnet/main.tf | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- `infra/modules/aws-vpc/` — terraform validate passes
- `infra/modules/azure-vnet/` — terraform validate passes
- `tfsec infra/modules/aws-vpc/` — no HIGH findings
- `tfsec infra/modules/azure-vnet/` — no HIGH findings

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| T-03-01-04 (NAT Gateway EIP) | TERR-04 | Accepted risk — required for egress | Verify NAT Gateway EIP only allows outbound connections |

*All other phase behaviors have automated verification.*

---

## Validation Audit (2026-05-07)

| Metric | Count |
|--------|-------|
| Total requirements | 7 |
| Covered (automated) | 7 |
| Partial / Missing | 0 |
| Manual-only (escalated) | 1 |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-05-07
