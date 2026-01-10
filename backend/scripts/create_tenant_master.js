#!/usr/bin/env node

/**
 * Script to create tenant_master document in Firestore
 * Run with: node create_tenant_master.js
 */

const admin = require('firebase-admin');
const path = require('path');

// Initialize Firebase Admin
const serviceAccountPath = path.join(__dirname, '..', 'config', 'firebase-adminsdk.json');
const serviceAccount = require(serviceAccountPath);

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount),
  databaseURL: 'https://ecosistema-imob-dev.firebaseio.com'
});

const db = admin.firestore();
const auth = admin.auth();

// Set the Firestore database to use
db.settings({
  databaseId: 'imob-dev'
});

async function createTenantMaster() {
  console.log('🚀 Starting tenant_master creation process...\n');

  try {
    // 1. Create tenant_master document
    console.log('📝 Creating tenant_master document...');
    const tenantMasterRef = db.collection('tenants').doc('tenant_master');

    await tenantMasterRef.set({
      name: 'ALTATECH Systems',
      document: '36.077.869/0001-81',
      document_type: 'cnpj',
      email: 'daniel.garcia@altatechsystems.com',
      phone: '+5511941491079',
      slug: 'altatech-systems',
      is_active: true,
      is_platform_admin: true,
      created_at: admin.firestore.FieldValue.serverTimestamp(),
      updated_at: admin.firestore.FieldValue.serverTimestamp()
    });
    console.log('✅ tenant_master document created successfully\n');

    // 2. Find the user document
    console.log('🔍 Finding user document...');
    const oldTenantId = '391b12f8-ebe4-426a-8c99-ec5a10b1f361';
    const userId = 'GMfu2R7TaCqi5qzgZq21';
    const firebaseUid = 'IXjrZzIapyU6QmPQVTDlecAJRu03';

    const oldUserRef = db.collection('tenants').doc(oldTenantId).collection('users').doc(userId);
    const oldUserSnap = await oldUserRef.get();

    if (!oldUserSnap.exists) {
      console.log('❌ User not found in old tenant');
      return;
    }

    const userData = oldUserSnap.data();
    console.log('✅ Found user:', userData.name);
    console.log('   Firebase UID:', firebaseUid);
    console.log('   Email:', userData.email, '\n');

    // 3. Create user in tenant_master
    console.log('📝 Creating user in tenant_master...');
    const newUserRef = db.collection('tenants').doc('tenant_master').collection('users').doc(userId);

    await newUserRef.set({
      ...userData,
      tenant_id: 'tenant_master',
      updated_at: admin.firestore.FieldValue.serverTimestamp()
    });
    console.log('✅ User created in tenant_master\n');

    // 4. Update Firebase custom claims
    console.log('🔐 Updating Firebase custom claims...');
    await auth.setCustomUserClaims(firebaseUid, {
      tenant_id: 'tenant_master',
      user_id: userId,
      role: 'admin'
    });
    console.log('✅ Custom claims updated\n');

    // 5. Verify the update
    console.log('🔍 Verifying custom claims...');
    const userRecord = await auth.getUser(firebaseUid);
    console.log('Custom claims:', userRecord.customClaims);
    console.log('\n✅ Migration completed successfully!\n');

    console.log('📋 Summary:');
    console.log('   - Created tenant_master document');
    console.log('   - Migrated user to tenant_master/users');
    console.log('   - Updated Firebase custom claims');
    console.log('\n⚠️  Note: The old tenant documents still exist. You can delete them manually if needed.\n');

  } catch (error) {
    console.error('❌ Error:', error);
    process.exit(1);
  }

  process.exit(0);
}

// Run the migration
createTenantMaster();
