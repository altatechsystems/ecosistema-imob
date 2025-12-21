package models

import "time"

// Property represents a physical real estate asset (imóvel)
// Collection: /tenants/{tenantId}/properties/{propertyId}
// IMPORTANTE: Property = ativo físico único (princípio de IMÓVEL ÚNICO)
type Property struct {
	ID       string `firestore:"-" json:"id"`
	TenantID string `firestore:"tenant_id" json:"tenant_id"`

	// Identificação
	Slug           string `firestore:"slug" json:"slug"`                                           // SEO-friendly, unique per tenant
	ExternalSource string `firestore:"external_source,omitempty" json:"external_source,omitempty"` // ex: "union"
	ExternalID     string `firestore:"external_id,omitempty" json:"external_id,omitempty"`         // ex: "77749175"
	Reference      string `firestore:"reference,omitempty" json:"reference,omitempty"`             // ex: "AP00335"

	// Proprietário
	OwnerID string `firestore:"owner_id" json:"owner_id"` // ref Owner

	// Tipo e localização
	PropertyType PropertyType `firestore:"property_type" json:"property_type"` // apartment, house, land, commercial
	Street       string       `firestore:"street,omitempty" json:"street,omitempty"`
	Number       string       `firestore:"number,omitempty" json:"number,omitempty"`
	Complement   string       `firestore:"complement,omitempty" json:"complement,omitempty"`
	Neighborhood string       `firestore:"neighborhood" json:"neighborhood"`
	City         string       `firestore:"city" json:"city"`
	State        string       `firestore:"state" json:"state"` // UF (ex: "SP")
	ZipCode      string       `firestore:"zip_code,omitempty" json:"zip_code,omitempty"`
	Country      string       `firestore:"country" json:"country"` // default "BR"

	// Características
	Bedrooms      int     `firestore:"bedrooms,omitempty" json:"bedrooms,omitempty"`
	Bathrooms     int     `firestore:"bathrooms,omitempty" json:"bathrooms,omitempty"`
	Suites        int     `firestore:"suites,omitempty" json:"suites,omitempty"`
	ParkingSpaces int     `firestore:"parking_spaces,omitempty" json:"parking_spaces,omitempty"`
	TotalArea     float64 `firestore:"total_area,omitempty" json:"total_area,omitempty"`   // m²
	UsableArea    float64 `firestore:"usable_area,omitempty" json:"usable_area,omitempty"` // m²

	// Preço e status (GOVERNANÇA)
	PriceAmount       float64        `firestore:"price_amount" json:"price_amount"`
	PriceCurrency     string         `firestore:"price_currency" json:"price_currency"` // "BRL"
	PriceConfirmedAt  *time.Time     `firestore:"price_confirmed_at,omitempty" json:"price_confirmed_at,omitempty"`
	Status            PropertyStatus `firestore:"status" json:"status"` // available, unavailable, pending_confirmation
	StatusConfirmedAt *time.Time     `firestore:"status_confirmed_at,omitempty" json:"status_confirmed_at,omitempty"`

	// Visibilidade e Co-corretagem (AI_DEV_DIRECTIVE Seção 20)
	Visibility         PropertyVisibility `firestore:"visibility" json:"visibility"`                             // private, network, marketplace, public
	VisibilityPublic   PropertyVisibility `firestore:"visibility_public" json:"visibility_public"`               // DEPRECATED: usar apenas Visibility
	CoBrokerCommission float64            `firestore:"co_broker_commission" json:"co_broker_commission"`         // % oferecida para selling_broker (ex: 40.0 = 40%)
	PendingReason      string             `firestore:"pending_reason,omitempty" json:"pending_reason,omitempty"` // stale_status, stale_price, owner_reported

	// Canonical Listing
	CanonicalListingID string `firestore:"canonical_listing_id,omitempty" json:"canonical_listing_id,omitempty"` // ref Listing

	// Deduplicação
	Fingerprint       string `firestore:"fingerprint" json:"fingerprint"` // hash(street+number+city+property_type+area)
	PossibleDuplicate bool   `firestore:"possible_duplicate" json:"possible_duplicate"`
	DataCompleteness  string `firestore:"data_completeness" json:"data_completeness"` // complete, incomplete, partial

	// PREPARAÇÃO: Lançamentos imobiliários (NULL no MVP, ativa em MVP+2)
	// Construtoras e loteadoras terão campos específicos para empreendimentos
	DevelopmentInfo *DevelopmentInfo `firestore:"development_info,omitempty" json:"development_info,omitempty"`

	// PREPARAÇÃO: Locação/Aluguel (NULL no MVP, ativa em MVP+3)
	// Suporte a anúncios de aluguel e gestão de contratos de locação
	TransactionType *TransactionType `firestore:"transaction_type,omitempty" json:"transaction_type,omitempty"` // sale, rent, both (default: sale no MVP)
	RentalInfo      *RentalInfo      `firestore:"rental_info,omitempty" json:"rental_info,omitempty"`           // Informações de locação

	// RESERVADO: Gestão de Contratos (MVP+4)
	CurrentContractID  *string    `firestore:"current_contract_id,omitempty" json:"current_contract_id,omitempty"`   // ref RentalContract
	ContractHistory    []string   `firestore:"contract_history,omitempty" json:"contract_history,omitempty"`         // IDs de contratos anteriores
	LastRentalEndDate  *time.Time `firestore:"last_rental_end_date,omitempty" json:"last_rental_end_date,omitempty"` // Última data de término
	AverageVacancyDays *int       `firestore:"average_vacancy_days,omitempty" json:"average_vacancy_days,omitempty"` // Média de dias vazio entre contratos

	// Metadata
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

// DevelopmentInfo contains information about real estate developments (lançamentos imobiliários)
// Used by developers and land developers for new projects (MVP+2)
type DevelopmentInfo struct {
	// Identificação da construtora/loteadora
	DeveloperID    string `firestore:"developer_id" json:"developer_id"` // Tenant da construtora
	DeveloperName  string `firestore:"developer_name" json:"developer_name"`
	DeveloperCRECI string `firestore:"developer_creci" json:"developer_creci"` // CRECI PJ

	// Informações do empreendimento
	ProjectName        string `firestore:"project_name" json:"project_name"` // "Residencial Vista Verde"
	ProjectSlug        string `firestore:"project_slug" json:"project_slug"` // URL do empreendimento
	ProjectDescription string `firestore:"project_description" json:"project_description"`

	// Unidades
	TotalUnits     int `firestore:"total_units" json:"total_units"`         // 200 unidades
	UnitsAvailable int `firestore:"units_available" json:"units_available"` // Unidades disponíveis
	UnitsSold      int `firestore:"units_sold" json:"units_sold"`           // Unidades vendidas
	UnitsReserved  int `firestore:"units_reserved" json:"units_reserved"`   // Unidades reservadas

	// Datas e status da obra
	LaunchDate         time.Time          `firestore:"launch_date" json:"launch_date"`                 // Data de lançamento
	DeliveryDate       time.Time          `firestore:"delivery_date" json:"delivery_date"`             // Previsão de entrega
	ConstructionStatus ConstructionStatus `firestore:"construction_status" json:"construction_status"` // plant, foundation, structure, finishing, ready

	// Financiamento
	AcceptsFinancing   bool    `firestore:"accepts_financing" json:"accepts_financing"`     // Aceita financiamento
	DownPaymentMin     float64 `firestore:"down_payment_min" json:"down_payment_min"`       // Entrada mínima (% do valor)
	InstallmentsDuring int     `firestore:"installments_during" json:"installments_during"` // Parcelas durante obra
	InstallmentsAfter  int     `firestore:"installments_after" json:"installments_after"`   // Parcelas após entrega

	// Mídia e documentos
	MasterPlanURL  string   `firestore:"master_plan_url,omitempty" json:"master_plan_url,omitempty"`   // Planta geral do empreendimento
	FloorPlansURLs []string `firestore:"floor_plans_urls,omitempty" json:"floor_plans_urls,omitempty"` // Plantas por tipologia
	Video360URL    string   `firestore:"video_360_url,omitempty" json:"video_360_url,omitempty"`       // Tour virtual 360°
	BrochureURL    string   `firestore:"brochure_url,omitempty" json:"brochure_url,omitempty"`         // Folder do empreendimento

	// Amenidades do condomínio
	Amenities []string `firestore:"amenities,omitempty" json:"amenities,omitempty"` // ["piscina", "quadra", "salao_festas", "playground"]
}

// RentalInfo contains specific information for rental properties (MVP+3)
type RentalInfo struct {
	// ===== VALORES MONETÁRIOS (obrigatórios) =====
	MonthlyRent      float64 `firestore:"monthly_rent" json:"monthly_rent"`                     // Aluguel mensal (R$ 2.500)
	CondoFee         float64 `firestore:"condo_fee,omitempty" json:"condo_fee,omitempty"`       // Condomínio (R$ 800)
	IPTUMonthly      float64 `firestore:"iptu_monthly,omitempty" json:"iptu_monthly,omitempty"` // IPTU mensal (R$ 150)
	TotalMonthlyCost float64 `firestore:"total_monthly_cost" json:"total_monthly_cost"`         // Total (R$ 3.450) = rent + condo + iptu

	// ===== GARANTIAS LOCATÍCIAS (obrigatório) =====
	DepositMonths      int      `firestore:"deposit_months" json:"deposit_months"`           // Caução (em meses, ex: 3 = R$ 7.500)
	AcceptedGuarantees []string `firestore:"accepted_guarantees" json:"accepted_guarantees"` // ["fiador", "caucao", "seguro_fianca", "fianca_bancaria"]

	// ===== TIPO DE LOCAÇÃO =====
	RentalType         RentalType `firestore:"rental_type" json:"rental_type"`                                     // traditional, corporate, short_term, vacation
	MinRentalPeriod    int        `firestore:"min_rental_period" json:"min_rental_period"`                         // Período mínimo (meses, ex: 12)
	Furnished          bool       `firestore:"furnished" json:"furnished"`                                         // Mobiliado? (true/false)
	PartiallyFurnished bool       `firestore:"partially_furnished,omitempty" json:"partially_furnished,omitempty"` // Semi-mobiliado

	// ===== RESTRIÇÕES E POLÍTICAS =====
	AcceptsPets       bool     `firestore:"accepts_pets" json:"accepts_pets"`                                 // Aceita pets?
	PetRestrictions   string   `firestore:"pet_restrictions,omitempty" json:"pet_restrictions,omitempty"`     // "Apenas cães pequenos", "Máximo 1 gato"
	UtilitiesIncluded []string `firestore:"utilities_included,omitempty" json:"utilities_included,omitempty"` // ["water", "gas", "internet"]

	// ===== INDEXAÇÃO (Reajuste Anual) =====
	IndexationType  IndexationType `firestore:"indexation_type,omitempty" json:"indexation_type,omitempty"`   // igpm, ipca, inpc
	AdjustmentMonth int            `firestore:"adjustment_month,omitempty" json:"adjustment_month,omitempty"` // Mês do reajuste (1-12)

	// ===== DISPONIBILIDADE =====
	AvailableFrom      *time.Time `firestore:"available_from,omitempty" json:"available_from,omitempty"` // Disponível a partir de (ex: 2025-02-01)
	ImmediateOccupancy bool       `firestore:"immediate_occupancy" json:"immediate_occupancy"`           // Ocupação imediata?
}
