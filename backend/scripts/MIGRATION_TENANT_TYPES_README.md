# Tenant Type Migration Script

This script migrates existing tenants to the new schema that includes `tenant_type` and subscription fields.

## What it does

The script updates all existing tenant documents in Firestore with the following fields:

- `tenant_type`: "pf" (Pessoa Física) or "pj" (Pessoa Jurídica)
- `business_type`: Inferred from existing data or set to defaults
- `subscription_plan`: Set to "full" for all existing tenants
- `subscription_status`: Set to "active"
- `subscription_started_at`: Current timestamp

## Prerequisites

### For JavaScript version:
```bash
cd backend/scripts
npm install firebase-admin
```

### For Go version:
```bash
cd backend/scripts
go mod init scripts 2>/dev/null || true
go get cloud.google.com/go/firestore
go get firebase.google.com/go/v4
```

### Firebase credentials:
Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to point to your Firebase service account JSON file:

```bash
# Linux/Mac
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/serviceAccountKey.json"

# Windows (PowerShell)
$env:GOOGLE_APPLICATION_CREDENTIALS="C:\path\to\serviceAccountKey.json"

# Windows (CMD)
set GOOGLE_APPLICATION_CREDENTIALS=C:\path\to\serviceAccountKey.json
```

## Running the migration

### ⚠️ IMPORTANT: Test in staging first!

1. **Backup your Firestore data** before running in production
2. Run the script in your **staging environment** first
3. Validate the results
4. Only then run in production

### JavaScript version (recommended):
```bash
cd backend/scripts
node migrate_tenant_types.js
```

### Go version:
```bash
cd backend/scripts
go run migrate_tenant_types.go
```

## Migration logic

The script uses the following logic to infer values:

### tenant_type:
- If `document_type == "cpf"` → `tenant_type = "pf"`
- Otherwise → `tenant_type = "pj"`

### business_type:
- If `tenant_type == "pf"` → `business_type = "corretor_autonomo"`
- If `tenant_type == "pj"` and no existing `business_type` → `business_type = "imobiliaria"` (safest default)
- Otherwise → keeps existing `business_type`

### Subscription:
- All tenants: `subscription_plan = "full"`, `subscription_status = "active"`

## Expected output

```
🚀 Starting tenant type migration...
⚠️  This will update all existing tenants with tenant_type and subscription fields

📊 Found 15 tenants to migrate

[1/15] Processing tenant: abc123...
   ✅ Migrated successfully
      - Tenant Type: pj
      - Business Type: imobiliaria
      - Subscription: full (active)

[2/15] Processing tenant: def456...
   ⏭️  Already migrated, skipping...

...

============================================================
📊 MIGRATION SUMMARY
============================================================
Total tenants: 15
✅ Successfully migrated: 12
⏭️  Skipped (already migrated): 3
❌ Errors: 0
============================================================

🎉 Migration completed successfully!
```

## Rollback

If something goes wrong, you can restore from your Firestore backup or manually remove the added fields:

```javascript
// Remove migration fields (run in Firestore console or script)
const batch = db.batch();
tenantsSnapshot.docs.forEach(doc => {
  batch.update(doc.ref, {
    tenant_type: admin.firestore.FieldValue.delete(),
    subscription_plan: admin.firestore.FieldValue.delete(),
    subscription_status: admin.firestore.FieldValue.delete(),
    subscription_started_at: admin.firestore.FieldValue.delete(),
  });
});
await batch.commit();
```

## Post-migration verification

After running the migration, verify:

1. All tenants have `tenant_type` field
2. All tenants have subscription fields
3. PF tenants have `business_type = "corretor_autonomo"`
4. No errors in application logs
5. Test signup flow with new multi-step form

## Related files

- Backend models: `backend/internal/models/tenant.go`
- Signup handler: `backend/internal/handlers/auth_handler.go`
- Frontend form: `frontend-admin/components/auth/signup-form.tsx`
- Validation schema: `frontend-admin/lib/validations.ts`
