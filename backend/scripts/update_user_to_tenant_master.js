#!/usr/bin/env node

/**
 * Script to update user's tenant_id to tenant_master
 * Run with: node update_user_to_tenant_master.js <firebase_uid>
 */

const admin = require('firebase-admin');
const path = require('path');

// Get Firebase UID from command line
const firebaseUid = process.argv[2];

if (!firebaseUid) {
  console.error('❌ Error: Please provide Firebase UID as argument');
  console.log('Usage: node update_user_to_tenant_master.js <firebase_uid>');
  console.log('Example: node update_user_to_tenant_master.js IXjrZzIapyU6QmPQVTDlecAJRu03');
  process.exit(1);
}

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

async function updateUserToTenantMaster() {
  console.log(`🚀 Updating user ${firebaseUid} to tenant_master...\n`);

  try {
    // 1. Get user from Firebase Auth
    console.log('🔍 Looking up user in Firebase Auth...');
    const userRecord = await auth.getUser(firebaseUid);
    console.log('✅ Found user:', userRecord.email);
    console.log('   Current custom claims:', userRecord.customClaims, '\n');

    // 2. Find user in Firestore (search all tenants)
    console.log('🔍 Searching for user in Firestore...');
    const usersQuery = await db.collectionGroup('users')
      .where('firebase_uid', '==', firebaseUid)
      .limit(1)
      .get();

    if (usersQuery.empty) {
      console.log('❌ User not found in Firestore');
      console.log('Creating new user document...\n');

      // Create new user document
      const newUserId = db.collection('tenants').doc().id;
      const newUserRef = db.collection('tenants').doc('tenant_master').collection('users').doc(newUserId);

      await newUserRef.set({
        firebase_uid: firebaseUid,
        tenant_id: 'tenant_master',
        email: userRecord.email,
        name: userRecord.displayName || 'Admin User',
        role: 'admin',
        is_active: true,
        created_at: admin.firestore.FieldValue.serverTimestamp(),
        updated_at: admin.firestore.FieldValue.serverTimestamp()
      });

      console.log('✅ User created in tenant_master/users/' + newUserId);

      // Update custom claims
      await auth.setCustomUserClaims(firebaseUid, {
        tenant_id: 'tenant_master',
        user_id: newUserId,
        role: 'admin'
      });

      console.log('✅ Custom claims updated\n');

    } else {
      // User found - update it
      const userDoc = usersQuery.docs[0];
      const userData = userDoc.data();
      const oldPath = userDoc.ref.path;
      const userId = userDoc.id;

      console.log('✅ Found user at:', oldPath);
      console.log('   User ID:', userId);
      console.log('   Current tenant_id:', userData.tenant_id, '\n');

      // Create user in tenant_master
      console.log('📝 Creating user in tenant_master...');
      const newUserRef = db.collection('tenants').doc('tenant_master').collection('users').doc(userId);

      await newUserRef.set({
        ...userData,
        tenant_id: 'tenant_master',
        updated_at: admin.firestore.FieldValue.serverTimestamp()
      });

      console.log('✅ User created in tenant_master/users/' + userId, '\n');

      // Update custom claims
      console.log('🔐 Updating Firebase custom claims...');
      await auth.setCustomUserClaims(firebaseUid, {
        tenant_id: 'tenant_master',
        user_id: userId,
        role: userData.role || 'admin'
      });

      console.log('✅ Custom claims updated\n');
    }

    // Verify
    const updatedUserRecord = await auth.getUser(firebaseUid);
    console.log('✅ Verification - new custom claims:', updatedUserRecord.customClaims);
    console.log('\n✅ Migration completed successfully!\n');

  } catch (error) {
    console.error('❌ Error:', error);
    process.exit(1);
  }

  process.exit(0);
}

// Run the migration
updateUserToTenantMaster();
