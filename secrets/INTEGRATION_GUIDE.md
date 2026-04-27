# Secrets Package - Cloud Integration Guide

This guide shows how to integrate the Golly secrets package with cloud providers for enterprise-grade credential management.

## Table of Contents

- [Overview](#overview)
- [AWS Secrets Manager Integration](#aws-secrets-manager-integration)
- [GCP Secret Manager Integration](#gcp-secret-manager-integration)
- [HashiCorp Vault Integration](#hashicorp-vault-integration)
- [Multi-Cloud Strategy](#multi-cloud-strategy)
- [Migration Guide](#migration-guide)
- [Production Checklist](#production-checklist)

---

## Overview

The secrets package provides a unified interface (`Store`) that works with any backend:

```go
// Same code works with any store
store := getCredentialStore() // Could be AWS, GCP, or Vault

// These operations are identical regardless of backend
credential, err := store.Get("api-key", ctx)
err = store.Write("api-key", credential, ctx)
err = store.Delete("api-key", ctx)
keys, err := store.List(ctx)
```

---

## AWS Secrets Manager Integration

### Setup

1. **Add Dependency**
   ```bash
   go get oss.nandlabs.io/golly-aws
   ```

2. **Configure AWS Credentials**
   - Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
   - AWS config file: `~/.aws/credentials`
   - IAM role (recommended for production)
   - ECS task role
   - EC2 instance profile

3. **Create IAM Policy**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "secretsmanager:CreateSecret",
           "secretsmanager:GetSecretValue",
           "secretsmanager:PutSecretValue",
           "secretsmanager:DeleteSecret",
           "secretsmanager:DescribeSecret",
           "secretsmanager:ListSecrets"
         ],
         "Resource": "arn:aws:secretsmanager:*:ACCOUNT-ID:secret:myapp/*"
       }
     ]
   }
   ```

### Basic Usage

```go
import (
    awssecrets "oss.nandlabs.io/golly-aws/secrets"
    "oss.nandlabs.io/golly/secrets"
)

// Create store
store, err := awssecrets.NewAWSSecretsStore(ctx, &awssecrets.AWSSecretsStoreConfig{
    Region: "us-east-1",
    TagFilter: map[string]string{
        "app":     "myapp",
        "environment": "production",
    },
    CacheTTL: 5 * time.Minute,
})
if err != nil {
    log.Fatal(err)
}

// Write credential
cred := &secrets.Credential{
    Value:       []byte("my-secret"),
    LastUpdated: time.Now(),
    Version:     "1.0",
}

err = store.Write("db-password", cred, ctx)

// Read credential
retrieved, err := store.Get("db-password", ctx)

// List all credentials with tag
keys, err := store.List(ctx)
```

### Advanced Features

**Automatic Secret Creation**:
```go
// Store.Write() automatically creates the secret if it doesn't exist
// Subsequent calls update the existing secret
store.Write("api-key", cred, ctx)  // Creates
store.Write("api-key", updated, ctx) // Updates
```

**Tagging Secrets**:
```go
config := &awssecrets.AWSSecretsStoreConfig{
    Region: "us-east-1",
    TagFilter: map[string]string{
        "app":     "myapp",
        "team":    "backend",
        "environment": "production",
        "version": "1.0",
    },
}

store, _ := awssecrets.NewAWSSecretsStore(ctx, config)
// Secrets created with these tags for organization
```

**Caching Strategy**:
```go
// For frequently accessed credentials
store, _ := awssecrets.NewAWSSecretsStore(ctx, &awssecrets.AWSSecretsStoreConfig{
    Region:   "us-east-1",
    CacheTTL: 5 * time.Minute, // Cache for 5 minutes
})

// Reduce API calls and costs
cred1, _ := store.Get("db-password", ctx) // Hits API
cred2, _ := store.Get("db-password", ctx) // Hits cache (same 5 min window)

// Clear cache when needed
store.ClearCache()
```

**Direct Client Access**:
```go
// For advanced operations not covered by Store interface
client := store.GetClient()
secret, err := client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
    SecretId: aws.String("db-password"),
})
```

### Cost Optimization

- **Regional Endpoints**: Use same region as application
- **VPC Endpoints**: Avoid NAT gateway costs
- **Caching**: Reduces API calls (charged per call)
- **Batch Operations**: Use ListSecrets with pagination

### Monitoring & Debugging

```go
// Track API calls
import "github.com/aws/aws-sdk-go-v2/aws/middleware"

// Enable debug logging
import "github.com/aws/aws-sdk-go-v2/aws"
// Set up logging in config

// Monitor credentials
keys, _ := store.List(ctx)
fmt.Printf("Storing %d credentials\n", len(keys))
```

---

## GCP Secret Manager Integration

### Setup

1. **Add Dependency**
   ```bash
   go get oss.nandlabs.io/golly-gcp
   ```

2. **Configure GCP Credentials**
   - Set `GOOGLE_APPLICATION_CREDENTIALS`: `/path/to/service-account.json`
   - Application Default Credentials (ADC): `gcloud auth application-default login`
   - Service account with appropriate permissions

3. **Enable APIs**
   ```bash
   gcloud services enable secretmanager.googleapis.com
   ```

4. **Create IAM Role**
   ```yaml
   title: "Golly Secrets Manager"
   description: "Access to Secret Manager for Golly"
   includedPermissions:
     - secretmanager.secrets.create
     - secretmanager.secrets.delete
     - secretmanager.secrets.get
     - secretmanager.secrets.list
     - secretmanager.secrets.update
     - secretmanager.versions.access
     - secretmanager.versions.add
   ```

### Basic Usage

```go
import (
    gcpsecrets "oss.nandlabs.io/golly-gcp/secrets"
    "oss.nandlabs.io/golly/secrets"
)

// Create store
store, err := gcpsecrets.NewGCPSecretStore(ctx, &gcpsecrets.GCPSecretStoreConfig{
    ProjectID: "my-gcp-project",
    Labels: map[string]string{
        "app":         "myapp",
        "environment": "production",
    },
    CacheTTL: 5 * time.Minute,
})
if err != nil {
    log.Fatal(err)
}

// Write credential (creates new version)
cred := &secrets.Credential{
    Value:       []byte("my-secret"),
    LastUpdated: time.Now(),
    Version:     "1.0",
}

err = store.Write("db-password", cred, ctx)

// Read credential (always gets latest version)
retrieved, err := store.Get("db-password", ctx)
```

### Advanced Features

**Automatic Version Management**:
```go
// GCP automatically maintains version history
store.Write("api-key", cred1, ctx) // Creates version 1
store.Write("api-key", cred2, ctx) // Creates version 2

// Accessing always gets latest, but versions are retained
// Enables rollback if needed
```

**Labeling Secrets**:
```go
config := &gcpsecrets.GCPSecretStoreConfig{
    ProjectID: "my-project",
    Labels: map[string]string{
        "app":         "myapp",
        "team":        "backend",
        "environment": "production",
        "data-classification": "confidential",
    },
}

store, _ := gcpsecrets.NewGCPSecretStore(ctx, config)
// Secrets created with these labels
// Use labels in GCP console for filtering
```

**Replication Strategy**:
```go
// GCP handles replication automatically
// Secrets are automatically replicated to supported locations
// Check in GCP console under Secret Manager → Replication
```

**Direct Client Access**:
```go
// For advanced operations
client := store.GetClient()

// Access specific version
version := "1"
secret, err := client.AccessSecretVersion(ctx, 
    &secretmanagerpb.AccessSecretVersionRequest{
        Name: fmt.Sprintf("projects/%s/secrets/api-key/versions/%s",
            projectID, version),
    })
```

### Cost Optimization

- **Automatic Replication**: No additional cost
- **Version Management**: Automatic, included in cost
- **Rotation**: Use Cloud Functions + Secret Manager
- **Caching**: Reduces API calls (charged per API call)

### Monitoring with Cloud Logging

```bash
# View access logs
gcloud logging read "resource.type=secretmanager.googleapis.com" \
  --limit=50 --format=json
```

---

## HashiCorp Vault Integration

### Setup

1. **Add Dependency**
   ```bash
   go get oss.nandlabs.io/golly-vault
   ```

2. **Install Vault**
   ```bash
   vault server -dev  # Development mode
   # Production: Deploy with HA storage backend
   ```

3. **Enable KV Engine**
   ```bash
   vault secrets enable -path=secret kv-v2
   ```

4. **Create Vault Policy**
   ```hcl
   # myapp-policy.hcl
   path "secret/data/myapp/*" {
     capabilities = ["create", "read", "update", "delete"]
   }
   
   path "secret/metadata/myapp/*" {
     capabilities = ["list"]
   }
   ```

   ```bash
   vault policy write myapp-policy myapp-policy.hcl
   ```

### Basic Usage

```go
import (
    vaultsecrets "oss.nandlabs.io/golly-vault/secrets"
    "oss.nandlabs.io/golly/secrets"
)

// Create store with token auth
store, err := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    Address:   "https://vault.example.com:8200",
    Token:     os.Getenv("VAULT_TOKEN"),
    Version:   "v2",
    BasePath:  "secret/data",
    CacheTTL:  5 * time.Minute,
})
if err != nil {
    log.Fatal(err)
}

// Write credential
cred := &secrets.Credential{
    Value:       []byte("my-secret"),
    LastUpdated: time.Now(),
    Version:     "1.0",
}

err = store.Write("myapp/db-password", cred, ctx)

// Read credential
retrieved, err := store.Get("myapp/db-password", ctx)
```

### Advanced Features

**Multiple Auth Methods**:

```go
// Token Auth (simple, for development)
store, _ := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    Token: "s.your-vault-token",
})

// Kubernetes Auth (production on K8s)
store, _ := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    Auth: &KubernetesAuth{
        ServiceAccountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
    },
})

// JWT Auth (JWT-based)
store, _ := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    Auth: &JWTAuth{
        Token: jwtToken,
    },
})

// AppRole Auth (automated)
store, _ := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    Auth: &AppRoleAuth{
        RoleID:   roleID,
        SecretID: secretID,
    },
})
```

**TLS Configuration**:
```go
store, _ := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    Address: "https://vault.example.com:8200",
    TLSConfig: &vaultsecrets.TLSConfig{
        CACert:     "/path/to/ca.crt",
        ClientCert: "/path/to/client.crt",
        ClientKey:  "/path/to/client.key",
    },
})
```

**Custom Path Mapping**:
```go
// Custom mapper for organizing credentials
type CustomPathMapper struct {
    org string
}

func (m *CustomPathMapper) MapPath(key string) string {
    return fmt.Sprintf("secret/data/orgs/%s/credentials/%s", m.org, key)
}

store, _ := vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
    PathMapper: &CustomPathMapper{org: "acme"},
})

// Credentials stored at: secret/data/orgs/acme/credentials/api-key
```

### Vault-Specific Features

**Metadata Management**:
```bash
# View secret metadata
vault kv metadata get secret/myapp/api-key

# List versions
vault kv metadata get secret/myapp/api-key
```

**Manual Rotation**:
```bash
# Update secret
vault kv put secret/myapp/db-password password="new-password"

# Check version history
vault kv list secret/metadata/myapp/db-password
```

**Audit Logging**:
```bash
# Enable audit logging
vault audit enable file file_path=/vault/logs/audit.log

# View audit logs
tail -f /vault/logs/audit.log
```

---

## Multi-Cloud Strategy

### Provider Abstraction

```go
package config

import (
    "os"
    awssecrets "oss.nandlabs.io/golly-aws/secrets"
    gcpsecrets "oss.nandlabs.io/golly-gcp/secrets"
    vaultsecrets "oss.nandlabs.io/golly-vault/secrets"
    "oss.nandlabs.io/golly/secrets"
)

func GetCredentialStore(ctx context.Context) (secrets.Store, error) {
    provider := os.Getenv("SECRET_PROVIDER")
    
    switch provider {
    case "aws":
        return awssecrets.NewAWSSecretsStore(ctx, &awssecrets.AWSSecretsStoreConfig{
            Region: os.Getenv("AWS_REGION"),
        })
    
    case "gcp":
        return gcpsecrets.NewGCPSecretStore(ctx, &gcpsecrets.GCPSecretStoreConfig{
            ProjectID: os.Getenv("GCP_PROJECT_ID"),
        })
    
    case "vault":
        return vaultsecrets.NewVaultStore(&vaultsecrets.VaultStoreConfig{
            Address: os.Getenv("VAULT_ADDRESS"),
            Token:   os.Getenv("VAULT_TOKEN"),
        })
    
    default:
        return nil, fmt.Errorf("unknown provider: %s", provider)
    }
}
```

### Multi-Store Manager

```go
// Register multiple stores
manager := secrets.GetManager()

awsStore, _ := awssecrets.NewAWSSecretsStore(ctx, awsConfig)
gcpStore, _ := gcpsecrets.NewGCPSecretStore(ctx, gcpConfig)

manager.Register("aws", awsStore)
manager.Register("gcp", gcpStore)

// Select at runtime
store := manager.Get(os.Getenv("ACTIVE_STORE"))
cred, _ := store.Get("api-key", ctx)
```

---

## Migration Guide

### From Local Store to AWS

```go
// Step 1: Read from local store
localStore, _ := secrets.NewLocalStore("/path/old.json", masterKey)
keys, _ := localStore.List(ctx)

// Step 2: Write to AWS
awsStore, _ := awssecrets.NewAWSSecretsStore(ctx, config)

for _, key := range keys {
    cred, _ := localStore.Get(key, ctx)
    awsStore.Write(key, cred, ctx)
    log.Printf("Migrated: %s", key)
}

// Step 3: Verify
awsKeys, _ := awsStore.List(ctx)
if len(awsKeys) == len(keys) {
    log.Println("Migration successful")
}
```

### From AWS to Vault

```go
awsStore, _ := awssecrets.NewAWSSecretsStore(ctx, config)
vaultStore, _ := vaultsecrets.NewVaultStore(vaultConfig)

keys, _ := awsStore.List(ctx)

for _, key := range keys {
    cred, _ := awsStore.Get(key, ctx)
    vaultStore.Write(key, cred, ctx)
}
```

---

## Production Checklist

### Security

- [ ] Master/KMS keys never in code or logs
- [ ] IAM policies follow principle of least privilege
- [ ] TLS/mTLS enabled for all connections
- [ ] Credentials rotated regularly
- [ ] Expiration dates enforced
- [ ] Audit logging enabled and monitored
- [ ] Encryption at rest enabled on all stores
- [ ] Encryption in transit (TLS) enforced

### Operations

- [ ] High availability configured
- [ ] Backup/disaster recovery tested
- [ ] Monitoring and alerting set up
- [ ] Error handling implemented
- [ ] Logging and audit trails configured
- [ ] Cache TTL appropriate for security/performance
- [ ] Rate limiting considered
- [ ] Credential rotation policy defined

### Compliance

- [ ] Access controls documented
- [ ] Audit logs retained per policy
- [ ] Encryption algorithm approved
- [ ] Data residency requirements met
- [ ] Compliance frameworks applied (SOC2, HIPAA, PCI-DSS, etc.)
- [ ] Documentation updated
- [ ] Team trained on secrets management

### Performance

- [ ] Caching strategy tuned
- [ ] Regional endpoints used
- [ ] Connection pooling enabled
- [ ] Load testing completed
- [ ] Scaling plan documented
- [ ] Cost optimization reviewed

---

For package-specific details, see:
- [AWS Secrets Store](../golly-aws/secrets/README.md)
- [GCP Secrets Store](../golly-gcp/secrets/README.md)
- [Vault Store](../golly-vault/secrets/README.md)
- [Architecture](ARCHITECTURE.md)
