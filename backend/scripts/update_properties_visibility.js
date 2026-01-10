const admin = require('firebase-admin');

// Initialize Firebase Admin
const serviceAccount = require('../config/firebase-adminsdk.json');

admin.initializeApp({
  credential: admin.credential.cert(serviceAccount),
  databaseURL: `https://${serviceAccount.project_id}.firebaseio.com`
});

// Connect to named database "imob-dev"
const db = admin.firestore();
db.settings({ databaseId: 'imob-dev' });

async function updatePropertiesVisibility() {
  try {
    console.log('🔍 Buscando todos os imóveis...');

    // Get all tenants
    const tenantsSnapshot = await db.collection('tenants').get();

    let totalProperties = 0;
    let updatedProperties = 0;

    for (const tenantDoc of tenantsSnapshot.docs) {
      const tenantId = tenantDoc.id;
      console.log(`\n📂 Processando tenant: ${tenantId} (${tenantDoc.data().name || 'Sem nome'})`);

      // Get all properties for this tenant
      const propertiesSnapshot = await db
        .collection('tenants')
        .doc(tenantId)
        .collection('properties')
        .get();

      console.log(`   Encontrados ${propertiesSnapshot.size} imóveis`);

      for (const propertyDoc of propertiesSnapshot.docs) {
        totalProperties++;
        const propertyData = propertyDoc.data();

        // Check current visibility
        if (propertyData.visibility !== 'public') {
          console.log(`   ✏️  Atualizando imóvel ${propertyDoc.id} (${propertyData.reference || 'sem ref'})`);
          console.log(`      Visibilidade: ${propertyData.visibility} → public`);

          await propertyDoc.ref.update({
            visibility: 'public',
            updated_at: admin.firestore.FieldValue.serverTimestamp()
          });

          updatedProperties++;
        }
      }
    }

    console.log('\n✅ Atualização concluída!');
    console.log(`📊 Total de imóveis: ${totalProperties}`);
    console.log(`📝 Imóveis atualizados: ${updatedProperties}`);
    console.log(`✓  Imóveis já públicos: ${totalProperties - updatedProperties}`);

  } catch (error) {
    console.error('❌ Erro ao atualizar imóveis:', error);
    process.exit(1);
  }

  process.exit(0);
}

// Run the update
updatePropertiesVisibility();
