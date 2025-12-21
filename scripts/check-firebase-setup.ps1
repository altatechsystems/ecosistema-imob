# Script de Verificação da Configuração Firebase
# Ecosistema Imob - Setup Checker

Write-Host "`n🔍 Verificando Configuração Firebase...`n" -ForegroundColor Cyan

$projectRoot = Split-Path -Parent $PSScriptRoot
$errors = @()
$warnings = @()
$success = @()

# 1. Verificar credenciais Firebase Admin SDK
Write-Host "1️⃣  Verificando Firebase Admin SDK..." -ForegroundColor Yellow
$adminSDKPath = Join-Path $projectRoot "backend\config\firebase-adminsdk.json"
if (Test-Path $adminSDKPath) {
    $success += "✅ Firebase Admin SDK encontrado"

    # Verificar se é JSON válido
    try {
        $json = Get-Content $adminSDKPath | ConvertFrom-Json
        if ($json.project_id -eq "ecosistema-imob-dev") {
            $success += "✅ Project ID correto: ecosistema-imob-dev"
        } else {
            $warnings += "⚠️  Project ID não é 'ecosistema-imob-dev': $($json.project_id)"
        }
    } catch {
        $errors += "❌ Arquivo JSON inválido"
    }
} else {
    $errors += "❌ Firebase Admin SDK não encontrado em: $adminSDKPath"
    Write-Host "   📥 Baixe em: https://console.firebase.google.com/project/ecosistema-imob-dev/settings/serviceaccounts/adminsdk" -ForegroundColor Gray
}

# 2. Verificar arquivo .env do backend
Write-Host "`n2️⃣  Verificando arquivo .env do backend..." -ForegroundColor Yellow
$envPath = Join-Path $projectRoot "backend\.env"
if (Test-Path $envPath) {
    $success += "✅ Arquivo .env encontrado"

    $envContent = Get-Content $envPath -Raw

    # Verificar variáveis essenciais
    if ($envContent -match "FIREBASE_PROJECT_ID=ecosistema-imob-dev") {
        $success += "✅ FIREBASE_PROJECT_ID configurado"
    } else {
        $errors += "❌ FIREBASE_PROJECT_ID não configurado ou incorreto"
    }

    if ($envContent -match "GOOGLE_APPLICATION_CREDENTIALS=") {
        $success += "✅ GOOGLE_APPLICATION_CREDENTIALS configurado"
    } else {
        $errors += "❌ GOOGLE_APPLICATION_CREDENTIALS não configurado"
    }

    if ($envContent -match "PORT=") {
        $success += "✅ PORT configurado"
    } else {
        $warnings += "⚠️  PORT não configurado (usará default 8080)"
    }
} else {
    $errors += "❌ Arquivo .env não encontrado"
    Write-Host "   📝 Copie de: backend\.env.example" -ForegroundColor Gray
}

# 3. Verificar Firebase CLI
Write-Host "`n3️⃣  Verificando Firebase CLI..." -ForegroundColor Yellow
try {
    $firebaseVersion = firebase --version 2>$null
    if ($firebaseVersion) {
        $success += "✅ Firebase CLI instalado: $firebaseVersion"
    } else {
        $warnings += "⚠️  Firebase CLI não encontrado (necessário para deploy de índices)"
        Write-Host "   📥 Instale com: npm install -g firebase-tools" -ForegroundColor Gray
    }
} catch {
    $warnings += "⚠️  Firebase CLI não encontrado"
    Write-Host "   📥 Instale com: npm install -g firebase-tools" -ForegroundColor Gray
}

# 4. Verificar firestore.indexes.json
Write-Host "`n4️⃣  Verificando firestore.indexes.json..." -ForegroundColor Yellow
$indexesPath = Join-Path $projectRoot "firestore.indexes.json"
if (Test-Path $indexesPath) {
    $success += "✅ firestore.indexes.json encontrado"

    try {
        $indexes = Get-Content $indexesPath | ConvertFrom-Json
        $indexCount = $indexes.indexes.Count
        if ($indexCount -eq 56) {
            $success += "✅ 56 índices Firestore definidos"
        } else {
            $warnings += "⚠️  Esperados 56 índices, encontrados: $indexCount"
        }
    } catch {
        $errors += "❌ Arquivo firestore.indexes.json inválido"
    }
} else {
    $errors += "❌ firestore.indexes.json não encontrado"
}

# 5. Verificar go.mod
Write-Host "`n5️⃣  Verificando dependências Go..." -ForegroundColor Yellow
$goModPath = Join-Path $projectRoot "backend\go.mod"
if (Test-Path $goModPath) {
    $success += "✅ go.mod encontrado"

    $goModContent = Get-Content $goModPath -Raw

    if ($goModContent -match "firebase.google.com/go/v4") {
        $success += "✅ Firebase Go SDK instalado"
    } else {
        $errors += "❌ Firebase Go SDK não encontrado no go.mod"
    }

    if ($goModContent -match "cloud.google.com/go/firestore") {
        $success += "✅ Firestore Go client instalado"
    } else {
        $errors += "❌ Firestore Go client não encontrado no go.mod"
    }

    if ($goModContent -match "github.com/gin-gonic/gin") {
        $success += "✅ Gin framework instalado"
    } else {
        $errors += "❌ Gin framework não encontrado no go.mod"
    }
} else {
    $errors += "❌ go.mod não encontrado"
}

# 6. Verificar modelos criados
Write-Host "`n6️⃣  Verificando modelos Go..." -ForegroundColor Yellow
$modelsPath = Join-Path $projectRoot "backend\internal\models"
if (Test-Path $modelsPath) {
    $modelFiles = @(
        "tenant.go",
        "broker.go",
        "property.go",
        "listing.go",
        "owner.go",
        "property_broker_role.go",
        "lead.go",
        "activity_log.go",
        "enums.go"
    )

    $foundModels = 0
    foreach ($file in $modelFiles) {
        if (Test-Path (Join-Path $modelsPath $file)) {
            $foundModels++
        }
    }

    if ($foundModels -eq $modelFiles.Count) {
        $success += "✅ Todos os 9 modelos criados"
    } else {
        $warnings += "⚠️  Modelos: $foundModels/$($modelFiles.Count) encontrados"
    }
} else {
    $errors += "❌ Pasta de modelos não encontrada"
}

# Resumo final
Write-Host "`n" + ("="*60) -ForegroundColor Cyan
Write-Host "📊 RESUMO DA VERIFICAÇÃO" -ForegroundColor Cyan
Write-Host ("="*60) -ForegroundColor Cyan

if ($success.Count -gt 0) {
    Write-Host "`n✅ SUCESSOS ($($success.Count)):" -ForegroundColor Green
    $success | ForEach-Object { Write-Host "   $_" -ForegroundColor Green }
}

if ($warnings.Count -gt 0) {
    Write-Host "`n⚠️  AVISOS ($($warnings.Count)):" -ForegroundColor Yellow
    $warnings | ForEach-Object { Write-Host "   $_" -ForegroundColor Yellow }
}

if ($errors.Count -gt 0) {
    Write-Host "`n❌ ERROS ($($errors.Count)):" -ForegroundColor Red
    $errors | ForEach-Object { Write-Host "   $_" -ForegroundColor Red }
    Write-Host "`n📖 Consulte: FIREBASE_SETUP_GUIDE.md para instruções completas`n" -ForegroundColor Cyan
    exit 1
} else {
    Write-Host "`n🎉 CONFIGURAÇÃO COMPLETA!" -ForegroundColor Green
    Write-Host "✅ Tudo pronto para iniciar a implementação do backend`n" -ForegroundColor Green
    exit 0
}
