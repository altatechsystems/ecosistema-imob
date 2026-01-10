const admin = require('firebase-admin');

// Initialize Firebase Admin for PRODUCTION
const serviceAccount = require('../config/firebase-adminsdk-prod.json');

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount),
  databaseURL: `https://${serviceAccount.project_id}.firebaseio.com`
});

// Connect to named database "imob-prod"
const db = admin.firestore();
db.settings({ databaseId: 'imob-prod' });

async function addTenantIdToUsers() {
  try {
    console.log('🔍 Buscando todos os tenants...\n');

    // Get all tenants
    const tenantsSnapshot = await db.collection('tenants').get();

    console.log(`📊 Encontrados ${tenantsSnapshot.size} tenants\n`);

    let totalUpdated = 0;
    let totalSkipped = 0;

    for (const tenantDoc of tenantsSnapshot.docs) {
      const tenantId = tenantDoc.id;
      console.log(`\n📂 Processando tenant: ${tenantId}`);

      // Get all users in this tenant
      const usersSnapshot = await db.collection('tenants').doc(tenantId).collection('users').get();

      console.log(`   Encontrados ${usersSnapshot.size} usuários`);

      for (const userDoc of usersSnapshot.docs) {
        const userData = userDoc.data();

        if (!userData.tenant_id) {
          console.log(`   ✏️  Adicionando tenant_id ao usuário: ${userDoc.id} (${userData.email || userData.name})`);

          await userDoc.ref.update({
            tenant_id: tenantId,
            updated_at: admin.firestore.FieldValue.serverTimestamp()
          });

          totalUpdated++;
        } else {
          console.log(`   ✓  Usuário já tem tenant_id: ${userDoc.id}`);
          totalSkipped++;
        }
      }
    }

    console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    console.log('✅ Atualização concluída!');
    console.log(`📝 Usuários atualizados: ${totalUpdated}`);
    console.log(`✓  Usuários já tinham tenant_id: ${totalSkipped}`);
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  } catch (error) {
    console.error('❌ Erro ao atualizar usuários:', error);
    process.exit(1);
  }

  process.exit(0);
}

// Run the update
addTenantIdToUsers();
