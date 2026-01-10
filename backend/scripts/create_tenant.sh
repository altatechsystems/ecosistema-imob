#!/bin/bash

# Script to create a tenant document in Firestore
# This is needed for the frontend-admin to work

TENANT_ID="bd71c02b-5fa5-43df-8b46-a1df2206f1ef"
PROJECT_ID="altatech-systems-imob-prod"

echo "Creating tenant document in Firestore..."
echo "Tenant ID: $TENANT_ID"
echo "Project ID: $PROJECT_ID"
echo ""

# Create tenant document using gcloud firestore
gcloud firestore documents create \
  --project="$PROJECT_ID" \
  --collection-path="tenants" \
  --document-id="$TENANT_ID" \
  --fields='
    name=string:Altatech Imobiliária,
    slug=string:altatech,
    email=string:contato@altatech.com.br,
    phone=string:+5535998671079,
    document=string:00000000000000,
    document_type=string:cnpj,
    business_type=string:imobiliaria,
    is_active=boolean:true,
    created_at=timestamp:2026-01-09T00:00:00Z,
    updated_at=timestamp:2026-01-09T00:00:00Z
  '

echo ""
echo "Tenant created successfully!"
echo "You can now access the frontend-admin"
