#!/bin/bash

# Script to create tenant_master document in Firestore DEV
# This creates a fixed ID tenant document instead of using UUIDs

echo "🔧 Creating tenant_master in Firestore DEV..."
echo ""
echo "Steps to execute manually in Firebase Console:"
echo ""
echo "1. Go to: https://console.firebase.google.com/project/ecosistema-imob-dev/firestore/databases/imob-dev/data/~2Ftenants"
echo ""
echo "2. Click 'Add document'"
echo ""
echo "3. Set Document ID to: tenant_master"
echo ""
echo "4. Add the following fields:"
echo ""
cat << 'EOF'
{
  "name": "ALTATECH Systems",
  "document": "36.077.869/0001-81",
  "document_type": "cnpj",
  "email": "daniel.garcia@altatechsystems.com",
  "phone": "+5511941491079",
  "slug": "altatech-systems-1766421954",
  "is_active": true,
  "is_platform_admin": true,
  "created_at": "2025-12-22T13:45:54Z",
  "updated_at": "2025-12-22T13:47:08Z"
}
EOF
echo ""
echo "5. After creating tenant_master, update the user document:"
echo "   Path: tenants/tenant_master/users/GMfu2R7TaCqi5qzgZq21"
echo "   Update field: tenant_id = 'tenant_master'"
echo ""
echo "6. Update Firebase custom claims for the user:"
echo "   Run: firebase auth:update-user IXjrZzIapyU6QmPQVTDlecAJRu03 --custom-claims '{\"tenant_id\":\"tenant_master\",\"role\":\"admin\"}'"
echo ""
