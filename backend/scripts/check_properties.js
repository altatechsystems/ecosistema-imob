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

async function checkProperties() {
  try {
    const tenantId = 'bd71c02b-5fa5-43df-8b46-a1df2206f1ef';
    const propertyId = '01617591-ba30-4d04-a116-d27f4fb0c801';

    console.log('🔍 Verificando estrutura do Firestore...\n');

    // Check specific property
    console.log(`📍 Buscando imóvel específico: ${propertyId}`);
    const propertyRef = db.collection('tenants').doc(tenantId).collection('properties').doc(propertyId);
    const propertyDoc = await propertyRef.get();

    if (propertyDoc.exists) {
      console.log('✅ Imóvel encontrado!');
      const data = propertyDoc.data();
      console.log(`   Referência: ${data.reference}`);
      console.log(`   Visibilidade: ${data.visibility}`);
      console.log(`   Status: ${data.status}`);
    } else {
      console.log('❌ Imóvel NÃO encontrado neste caminho');
    }

    // List all collections under tenant
    console.log(`\n📂 Listando coleções do tenant ${tenantId}:`);
    const tenantRef = db.collection('tenants').doc(tenantId);
    const collections = await tenantRef.listCollections();
    collections.forEach(collection => {
      console.log(`   - ${collection.id}`);
    });

    // Try to find properties in different paths
    console.log('\n🔎 Buscando imóveis em diferentes caminhos...');

    // Path 1: tenants/{tenant_id}/properties
    const path1 = await db.collection('tenants').doc(tenantId).collection('properties').limit(5).get();
    console.log(`   Path: tenants/{tenant_id}/properties → ${path1.size} documentos`);

    // Path 2: properties (root level)
    const path2 = await db.collection('properties').where('tenant_id', '==', tenantId).limit(5).get();
    console.log(`   Path: properties (root) → ${path2.size} documentos`);

    // Path 3: Collection group query
    const path3 = await db.collectionGroup('properties').where('tenant_id', '==', tenantId).limit(5).get();
    console.log(`   Path: collectionGroup('properties') → ${path3.size} documentos`);

    if (path3.size > 0) {
      console.log('\n📝 Primeiros imóveis encontrados (collection group):');
      path3.forEach(doc => {
        const data = doc.data();
        console.log(`   - ${doc.id}: ${data.reference} (${data.visibility})`);
        console.log(`     Caminho: ${doc.ref.path}`);
      });
    }

  } catch (error) {
    console.error('❌ Erro:', error);
  }

  process.exit(0);
}

checkProperties();
