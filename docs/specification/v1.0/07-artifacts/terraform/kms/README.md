# terraform module: kms
Root CMK (annual rotation) + asymmetric signing key for approval tokens.
Per-tenant DEK aliases are created at tenant provisioning by tenant-svc via
IAM-scoped CreateAlias (policy in this module's iam.tf extension point).
Signing alg is provider-conditional: ECDSA-P256 (AWS) / Ed25519 (GCP, self-host);
token header carries alg+kid — see kms-layout.yaml.
