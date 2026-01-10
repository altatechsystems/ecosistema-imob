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

async function checkLeads() {
  try {
    const tenantId = 'bd71c02b-5fa5-43df-8b46-a1df2206f1ef';

    console.log('🔍 Verificando leads...\n');

    // Check leads in tenant subcollection
    console.log(`📂 Path: tenants/${tenantId}/leads`);
    const leadsRef = db.collection('tenants').doc(tenantId).collection('leads');
    const leadsSnapshot = await leadsRef.limit(5).get();

    console.log(`   Encontrados: ${leadsSnapshot.size} leads\n`);

    if (leadsSnapshot.size > 0) {
      console.log('📝 Primeiros leads:');
      leadsSnapshot.forEach(doc => {
        const data = doc.data();
        console.log(`\n   Lead ID: ${doc.id}`);
        console.log(`   Nome: ${data.name || 'N/A'}`);
        console.log(`   Email: ${data.email || 'N/A'}`);
        console.log(`   Telefone: ${data.phone || 'N/A'}`);
        console.log(`   Status: ${data.status || 'N/A'}`);
        console.log(`   Canal: ${data.channel || 'N/A'}`);
        console.log(`   Property ID: ${data.property_id || 'N/A'}`);
        console.log(`   Created: ${data.created_at ? data.created_at.toDate() : 'N/A'}`);
      });
    }

    // Try to count all leads
    const allLeadsSnapshot = await leadsRef.get();
    console.log(`\n📊 Total de leads no tenant: ${allLeadsSnapshot.size}`);

  } catch (error) {
    console.error('❌ Erro:', error);
  }

  process.exit(0);
}

checkLeads();
