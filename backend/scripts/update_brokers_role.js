/**
 * Script to update users with CRECI from "manager" role to "broker" role
 *
 * This fixes the issue where all real estate agents were incorrectly classified as managers
 * when they should be brokers.
 */

const admin = require('firebase-admin');

// Initialize Firebase Admin
const serviceAccount = require('../config/firebase-adminsdk.json');

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount)
});

const db = admin.firestore();

async function updateBrokersRole() {
  try {
    console.log('🔍 Searching for users with CRECI and manager role...\n');

    // Get all tenants
    const tenantsSnapshot = await db.collection('tenants').get();

    let totalUpdated = 0;

    for (const tenantDoc of tenantsSnapshot.docs) {
      const tenantId = tenantDoc.id;
      const tenantName = tenantDoc.data().name;

      console.log(`\n📋 Checking tenant: ${tenantName} (${tenantId})`);

      // Get all users in this tenant that have CRECI
      const usersSnapshot = await db
        .collection('tenants')
        .doc(tenantId)
        .collection('users')
        .where('creci', '!=', '')
        .get();

      console.log(`   Found ${usersSnapshot.size} users with CRECI`);

      let tenantUpdated = 0;

      for (const userDoc of usersSnapshot.docs) {
        const userData = userDoc.data();
        const currentRole = userData.role || 'unknown';

        // Only update if role is not already 'broker'
        if (currentRole !== 'broker') {
          console.log(`   ✏️  Updating ${userData.name} (${userData.email})`);
          console.log(`      CRECI: ${userData.creci}`);
          console.log(`      Current role: ${currentRole} → New role: broker`);

          await db
            .collection('tenants')
            .doc(tenantId)
            .collection('users')
            .doc(userDoc.id)
            .update({
              role: 'broker',
              updated_at: admin.firestore.FieldValue.serverTimestamp()
            });

          tenantUpdated++;
          totalUpdated++;
        } else {
          console.log(`   ✓ ${userData.name} already has broker role`);
        }
      }

      console.log(`   Updated ${tenantUpdated} users in ${tenantName}`);
    }

    console.log(`\n✅ SUCCESS! Updated ${totalUpdated} users total`);
    console.log('\nAll users with CRECI now have the "broker" role.');

  } catch (error) {
    console.error('❌ Error updating brokers role:', error);
    process.exit(1);
  }

  process.exit(0);
}

// Run the update
updateBrokersRole();
