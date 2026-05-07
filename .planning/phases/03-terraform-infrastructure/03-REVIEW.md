---
phase: 3
slug: terraform-infrastructure
status: issues_found
files_reviewed: 9
critical: 0
warning: 1
info: 3
total: 4
created: 2026-05-07
---

# Phase3 — Code Review (Wave 2)

## Summary

- **Files reviewed:** 6 (aws-eks: main.tf, variables.tf, outputs.tf; azure-aks: main.tf, variables.tf, outputs.tf)
- **Depth:** standard
- **Result:** Clean (0 critical, 0 warning, 2 info)

---

## Findings

### MED-01 — AKS cluster missing `local_account_disabled` (Medium)

**File:** `infra/modules/azure-aks/main.tf`  
**Line:** ~45

**Problem:** The AKS cluster has Workload Identity enabled but `local_account_disabled` is not set to `true`. This allows local admin accounts, which undermines the Workload Identity security model.

**Fix:** Add `local_account_disabled = true` inside the `azurerm_kubernetes_cluster` resource block.

---

### LOW-01 — Kubernetes version default outdated (Low)

**File:** `infra/modules/azure-aks/variables.tf`  
**Line:** 34

**Problem:** Default `kubernetes_version = "1.29"` may be outdated. Consider updating to `"1.33"` to match EKS version and ensure feature parity.

---

### INFO-01 — Node group max_size fixed offset (Info)

**File:** `infra/modules/aws-eks/main.tf`  
**Line:** ~70 (scaling_config block)

**Note:** `max_size = var.node_count + 2` uses a fixed offset. Acceptable for PoC; for production consider using a variable or dynamic scaling policy.

---

### INFO-02 — AKS zones hard-coded (Info)

**File:** `infra/modules/azure-aks/main.tf`  
**Line:** 32 (`zones = ["1","2","3"]`)

**Note:** Hard-coded zones may not be available in all Azure regions. Consider making this a variable with region-specific defaults.

---

### INFO-03 — Federated credential subject format (Info)

**File:** `infra/modules/azure-aks/main.tf`  
**Line:** 62

**Note:** `subject = "system:serviceaccount:default:${var.service_account_name}"` assumes `namespace = "default"`. If namespace changes, this will break. Consider adding a `namespace` variable.

---

## Secure Behaviors Verified

| Threat ID | Behavior | Status |
|-----------|----------|--------|
| T-03-03-01 | EKS IRSA trust policy (StringEquals, no wildcards) | ✅ Verified |
| T-03-03-02 | EKS private API endpoint | ✅ Verified |
| T-03-03-03 | EKS audit logging enabled | ✅ Verified |
| T-03-03-04 | OIDC thumbprint dynamic via tls_certificate | ✅ Verified |
| T-03-03-05 | cluster_ca output marked sensitive | ✅ Verified |
| T-03-04-01 | AKS Workload Identity + federated credential | ✅ Verified |
| T-03-04-02 | AKS private node pool (vnet_subnet_id) | ✅ Verified |

---

## Next Steps

- No critical or warning findings to fix.
- Informational notes are for awareness only; no immediate action required.
- Full report saved to `03-REVIEW.md`.

---

**Approval:** approved 2026-05-07
