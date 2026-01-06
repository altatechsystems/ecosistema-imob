# Migration Tool: Brokers to Users

## Overview

This utility implements **PROMPT 10**: Separation of Brokers and Administrative Users.

### Problem

Currently, all users are stored in `/tenants/{tenantId}/brokers` collection, including:
- Real estate agents (with CRECI registration)
- Administrative users (without CRECI)

This creates UX and security issues:
- Administrators appear in broker listings
- CRECI validation is inconsistent
- Role-based permissions are unclear

### Solution

Separate users into two collections:
- `/tenants/{tenantId}/brokers` - Real estate agents with valid CRECI
- `/tenants/{tenantId}/users` - Administrative users without CRECI

## What This Tool Does

1. **Scans all tenants** in the database
2. **Examines each broker** in `/tenants/{tenantId}/brokers`
3. **Validates CRECI**:
   - Valid format: `XXXXX-F/UF` or `XXXXX-J/UF` (e.g., "12345-F/SP")
   - Invalid: empty, "-", "N/A", "PENDENTE", or malformed
4. **Migrates users without valid CRECI**:
   - Creates record in `/tenants/{tenantId}/users`
   - Maps roles: `broker_admin`/`admin` → `admin`, `broker`/`manager` → `manager`
   - Sets default permissions based on role
   - Deletes from `/brokers` collection
5. **Keeps brokers with valid CRECI** in `/brokers` collection

## Prerequisites

- Go 1.21+
- Firebase Admin SDK credentials
- Access to Firestore database

## Usage

### 1. Set up Firebase credentials

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/firebase-adminsdk.json"
```

Or place the file at: `./config/firebase-adminsdk.json`

### 2. Run the migration (DRY RUN first recommended)

```bash
cd backend/cmd/migrate-brokers-to-users
go run main.go
```

### 3. Review the output

The tool will show:
```
📋 Step 1: Finding all tenants...
✅ Found 3 tenants

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🏢 Tenant: Altatech Imóveis (bd71c02b-5fa5-43df-8b46-a1df2206f1ef)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ✅ KEEP: João Silva (joao@example.com) - CRECI: 12345-F/SP
   🔄 MIGRATE: Administrao (admin@example.com) - No valid CRECI (value: '-')
      ✅ Created user with ID: abc123
      ✅ Deleted from brokers collection

   📊 Tenant Summary: 7 total, 1 migrated to users, 6 kept as brokers
```

## CRECI Validation Rules

A CRECI is considered **valid** if:
- Has at least 8 characters
- Contains `-` and `/`
- Contains `-F/` (Pessoa Física) or `-J/` (Pessoa Jurídica)
- Has state code (e.g., "SP", "RJ", "MG")
- Example: `12345-F/SP`, `67890-J/RJ`

A CRECI is considered **invalid** if:
- Empty string
- Placeholder values: `-`, `N/A`, `PENDENTE`, `PENDING`
- Malformed (too short, missing separators)

## Role Mapping

| Broker Role | User Role | Permissions |
|------------|-----------|-------------|
| `broker_admin` | `admin` | Full permissions (properties, brokers, users, settings) |
| `admin` | `admin` | Full permissions |
| `broker` | `manager` | View/edit properties, view brokers/users, manage leads |
| `manager` | `manager` | View/edit properties, view brokers/users, manage leads |

## Safety Features

- **Non-destructive**: Creates user before deleting broker
- **Detailed logging**: Shows every action taken
- **Error handling**: Continues on errors, logs warnings
- **Rollback possible**: Can manually restore from Firestore backups if needed

## Post-Migration Steps

After running the migration:

1. **Verify in Firestore Console**
   - Check `/tenants/{tenantId}/users` collection exists
   - Verify migrated users have correct data
   - Confirm only valid CRECI brokers remain in `/brokers`

2. **Test Authentication**
   - Login with a migrated user (should work - Login now searches both collections)
   - Login with a broker (should still work)
   - Verify custom claims are set correctly

3. **Update Frontend**
   - Modify `/equipe` page to fetch from `/users` endpoint
   - Keep `/corretores` page fetching only from `/brokers`
   - Add user type badges to differentiate

4. **Make CRECI Mandatory**
   - Update broker creation validation to require CRECI
   - Update signup flow to ask "Are you a broker?"

## Troubleshooting

### Error: "Failed to create Firestore client"
- Check `GOOGLE_APPLICATION_CREDENTIALS` path
- Verify Firebase Admin SDK JSON is valid
- Ensure project ID and database ID are correct

### Error: "Failed to create user"
- Check Firestore security rules allow writes to `/users`
- Verify tenant ID is valid
- Check user data is complete (required fields)

### Warning: "Failed to delete broker"
- User was created successfully but broker remains
- Manually delete the broker from Firestore Console
- Or re-run the migration (it will skip existing users)

## Example Output

```
═══════════════════════════════════════════════════════════════
MIGRATION COMPLETE
═══════════════════════════════════════════════════════════════
📊 Total brokers processed: 23
🔄 Migrated to /users: 7
✅ Kept as /brokers: 16

Next steps:
1. Verify the migration by checking Firestore console
2. Test login with both brokers and users
3. Update frontend to use /users endpoint for administrative users
4. Make CRECI mandatory for new broker creations
```

## Rollback (if needed)

If you need to rollback the migration:

1. Export `/tenants/{tenantId}/users` to backup
2. For each user, create corresponding broker in `/brokers`
3. Delete `/tenants/{tenantId}/users` collection

Better approach: **Always backup Firestore before running migrations**

```bash
# Backup Firestore (requires gcloud CLI)
gcloud firestore export gs://your-backup-bucket/backup-$(date +%Y%m%d)
```

## Related Documentation

- [PROMPT 10](../../prompts/10_sistema_robusto_perfis_acesso.txt)
- [CHECKPOINT](../../CHECKPOINT_06_JAN_2026.md)
- [User Model](../../internal/models/user.go)
- [Broker Model](../../internal/models/broker.go)
