#!/usr/bin/env node

/**
 * Script to migrate existing tenants to new schema with tenant_type and subscription fields
 *
 * Usage:
 *   node migrate_tenant_types.js
 *
 * Prerequisites:
 *   - Firebase Admin SDK credentials set in GOOGLE_APPLICATION_CREDENTIALS env var
 *   - Node.js installed
 *   - firebase-admin npm package installed
 */

const admin = require('firebase-admin');

// Initialize Firebase Admin
const serviceAccount = process.env.GOOGLE_APPLICATION_CREDENTIALS;
if (!serviceAccount) {
  console.error('❌ GOOGLE_APPLICATION_CREDENTIALS environment variable not set');
  process.exit(1);
}

admin.initializeApp({
  credential: admin.credential.applicationDefault(),
});

const db = admin.firestore();

async function migrateTenants() {
  console.log('🚀 Starting tenant type migration...');
  console.log('⚠️  This will update all existing tenants with tenant_type and subscription fields\n');

  try {
    // Fetch all tenants
    const tenantsSnapshot = await db.collection('tenants').get();
    const totalTenants = tenantsSnapshot.size;

    console.log(`📊 Found ${totalTenants} tenants to migrate\n`);

    let successCount = 0;
    let errorCount = 0;
    let skippedCount = 0;

    let index = 0;
    for (const tenantDoc of tenantsSnapshot.docs) {
      index++;
      const tenantID = tenantDoc.id;
      const tenantData = tenantDoc.data();

      console.log(`[${index}/${totalTenants}] Processing tenant: ${tenantID}`);

      // Check if already migrated
      if (tenantData.tenant_type) {
        console.log('   ⏭️  Already migrated, skipping...');
        skippedCount++;
        continue;
      }

      try {
        // Infer tenant_type from document_type
        let tenantType = 'pj'; // Default to PJ
        if (tenantData.document_type === 'cpf') {
          tenantType = 'pf';
        }

        // Infer business_type if not exists
        let businessType = tenantData.business_type || '';
        if (!businessType) {
          if (tenantType === 'pf') {
            businessType = 'corretor_autonomo';
          } else {
            // Default PJ to imobiliaria (safest assumption)
            businessType = 'imobiliaria';
          }
        }

        // Prepare updates
        const updates = {
          tenant_type: tenantType,
          business_type: businessType,
          subscription_plan: 'full',
          subscription_status: 'active',
          subscription_started_at: admin.firestore.FieldValue.serverTimestamp(),
        };

        // Apply updates
        await tenantDoc.ref.update(updates);

        console.log('   ✅ Migrated successfully');
        console.log(`      - Tenant Type: ${tenantType}`);
        console.log(`      - Business Type: ${businessType}`);
        console.log('      - Subscription: full (active)');

        successCount++;
      } catch (error) {
        console.log(`   ❌ Error updating tenant ${tenantID}:`, error.message);
        errorCount++;
      }

      console.log(''); // Empty line for readability
    }

    // Summary
    console.log('='.repeat(60));
    console.log('📊 MIGRATION SUMMARY');
    console.log('='.repeat(60));
    console.log(`Total tenants: ${totalTenants}`);
    console.log(`✅ Successfully migrated: ${successCount}`);
    console.log(`⏭️  Skipped (already migrated): ${skippedCount}`);
    console.log(`❌ Errors: ${errorCount}`);
    console.log('='.repeat(60));

    if (errorCount > 0) {
      console.log('\n⚠️  Some tenants failed to migrate. Check the logs above for details.');
      process.exit(1);
    }

    console.log('\n🎉 Migration completed successfully!');
    process.exit(0);
  } catch (error) {
    console.error('❌ Fatal error during migration:', error);
    process.exit(1);
  }
}

// Run migration
migrateTenants();
