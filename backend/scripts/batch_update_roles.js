/**
 * Batch update user roles
 * Updates all users in a tenant to have the 'broker' role
 */

const admin = require('firebase-admin');

// Initialize Firebase Admin
const serviceAccount = require('../config/firebase-adminsdk.json');

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount)
});

const db = admin.firestore();

// Tenant ID for ALTATECH Imóveis
const TENANT_ID = 'bd71c02b-5fa5-43df-8b46-a1df2206f1ef';

// User IDs to update (from the console logs)
const userIdsToUpdate = [
  'E0H6ixFYcDOppcEHrJft', // Suzana Costa
  'E68JsxLHDzTSCVHLWzus', // (unknown name)
  'Q25xCfitLjmZz1bNQpBo', // (unknown name)
  'QZm6iPxGy93GF6ZeeabY', // (unknown name)
  'f39046f1-c833-4c11-bd92-2c6420830979', // (unknown name)
  'hirzSgIZhQLzuMdUdvJX', // (unknown name)
  'i1kCAou1loc9nGqJFApk', // (unknown name)
  'uI0PK4FgmLfgKIlR7lDR'  // (unknown name)
];

async function batchUpdateRoles() {
  try {
    console.log(`🔄 Updating ${userIdsToUpdate.length} users to 'broker' role...\n`);

    let updated = 0;
    let errors = 0;

    for (const userId of userIdsToUpdate) {
      try {
        const userRef = db
          .collection('tenants')
          .doc(TENANT_ID)
          .collection('users')
          .doc(userId);

        const userDoc = await userRef.get();

        if (!userDoc.exists) {
          console.log(`⚠️  User ${userId} not found`);
          errors++;
          continue;
        }

        const userData = userDoc.data();
        console.log(`✏️  Updating ${userData.name || 'Unknown'} (${userData.email})`);
        console.log(`   Current role: ${userData.role} → New role: broker`);

        await userRef.update({
          role: 'broker',
          updated_at: admin.firestore.FieldValue.serverTimestamp()
        });

        updated++;
        console.log(`   ✅ Updated successfully\n`);

      } catch (error) {
        console.error(`   ❌ Error updating user ${userId}:`, error.message);
        errors++;
      }
    }

    console.log(`\n✅ Batch update complete!`);
    console.log(`   Updated: ${updated} users`);
    console.log(`   Errors: ${errors} users`);

  } catch (error) {
    console.error('❌ Fatal error:', error);
    process.exit(1);
  }

  process.exit(0);
}

// Run the batch update
batchUpdateRoles();
