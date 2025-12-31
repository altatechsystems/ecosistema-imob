package models

// PropertyType defines the type of property
type PropertyType string

const (
	PropertyTypeApartment  PropertyType = "apartment"
	PropertyTypeHouse      PropertyType = "house"
	PropertyTypeLand       PropertyType = "land"
	PropertyTypeCommercial PropertyType = "commercial"
	// NOVOS: Lançamentos (MVP+2)
	PropertyTypeNewDevelopment PropertyType = "new_development" // Apartamento em lançamento
	PropertyTypeCondoLot       PropertyType = "condo_lot"       // Lote em condomínio
	PropertyTypeBuildingLot    PropertyType = "building_lot"    // Terreno para construção
)

// PropertyStatus defines the status of a property
type PropertyStatus string

const (
	PropertyStatusAvailable           PropertyStatus = "available"
	PropertyStatusUnavailable         PropertyStatus = "unavailable"
	PropertyStatusPendingConfirmation PropertyStatus = "pending_confirmation"
)

// PropertyVisibility defines the visibility level of a property
type PropertyVisibility string

const (
	// Visibilidade escalonada (AI_DEV_DIRECTIVE Seção 20.2)
	PropertyVisibilityPrivate     PropertyVisibility = "private"     // apenas captador
	PropertyVisibilityNetwork     PropertyVisibility = "network"     // tenant (imobiliária)
	PropertyVisibilityMarketplace PropertyVisibility = "marketplace" // todos os corretores
	PropertyVisibilityPublic      PropertyVisibility = "public"      // internet (SEO)

	// DEPRECATED (manter por compatibilidade temporária)
	PropertyVisibilityHiddenStale       PropertyVisibility = "hidden_stale"
	PropertyVisibilityHiddenUnavailable PropertyVisibility = "hidden_unavailable"
)

// ConstructionStatus defines the construction status for developments
type ConstructionStatus string

const (
	ConstructionStatusPlant      ConstructionStatus = "plant"      // Na planta
	ConstructionStatusFoundation ConstructionStatus = "foundation" // Fundação
	ConstructionStatusStructure  ConstructionStatus = "structure"  // Estrutura
	ConstructionStatusFinishing  ConstructionStatus = "finishing"  // Acabamento
	ConstructionStatusReady      ConstructionStatus = "ready"      // Pronto
)

// TransactionType defines the type of transaction (Tipo de Transação) - MVP+3
type TransactionType string

const (
	TransactionTypeSale TransactionType = "sale" // Venda (padrão no MVP)
	TransactionTypeRent TransactionType = "rent" // Aluguel/Locação
	TransactionTypeBoth TransactionType = "both" // Disponível para venda OU aluguel
)

// RentalType defines the type of rental
type RentalType string

const (
	RentalTypeTraditional RentalType = "traditional" // Locação tradicional (12+ meses, residencial)
	RentalTypeCorporate   RentalType = "corporate"   // Locação corporativa (6-12 meses, mobiliado, para empresas)
	RentalTypeShortTerm   RentalType = "short_term"  // Temporada/curta duração (1-6 meses)
	RentalTypeVacation    RentalType = "vacation"    // Férias/Airbnb-style (dias/semanas)
)

// GuaranteeType defines the type of rental guarantee
type GuaranteeType string

const (
	GuaranteeTypeFiador         GuaranteeType = "fiador"          // Fiador pessoa física com imóvel próprio
	GuaranteeTypeCaucao         GuaranteeType = "caucao"          // Caução (3-6 meses adiantados)
	GuaranteeTypeSeguroFianca   GuaranteeType = "seguro_fianca"   // Seguro Fiança (seguradora)
	GuaranteeTypeFiancaBancaria GuaranteeType = "fianca_bancaria" // Carta de Fiança Bancária
)

// IndexationType defines the indexation type for rent adjustment
type IndexationType string

const (
	IndexationTypeIGPM IndexationType = "igpm" // IGP-M (Índice Geral de Preços do Mercado)
	IndexationTypeIPCA IndexationType = "ipca" // IPCA (Índice Nacional de Preços ao Consumidor Amplo)
	IndexationTypeINPC IndexationType = "inpc" // INPC (Índice Nacional de Preços ao Consumidor)
)

// OwnerStatus defines the completeness status of owner data
type OwnerStatus string

const (
	OwnerStatusIncomplete OwnerStatus = "incomplete" // placeholder
	OwnerStatusPartial    OwnerStatus = "partial"    // alguns dados
	OwnerStatusVerified   OwnerStatus = "verified"   // completo e verificado
)

// BrokerPropertyRole defines the role of a broker in relation to a property
type BrokerPropertyRole string

const (
	// Captador: corretor que originou/captou o imóvel (único por Property)
	BrokerPropertyRoleOriginating BrokerPropertyRole = "originating_broker"

	// Vendedor: corretor responsável por um Listing (pode haver múltiplos)
	BrokerPropertyRoleListing BrokerPropertyRole = "listing_broker"

	// Co-corretor: corretor adicional na negociação (comum no Brasil)
	BrokerPropertyRoleCoBroker BrokerPropertyRole = "co_broker"
)

// LeadChannel defines the channel through which a lead was created
type LeadChannel string

const (
	LeadChannelWhatsApp LeadChannel = "whatsapp"
	LeadChannelForm     LeadChannel = "form"
	LeadChannelPhone    LeadChannel = "phone"
	LeadChannelEmail    LeadChannel = "email"
)

// LeadStatus defines the status of a lead
type LeadStatus string

const (
	LeadStatusNew         LeadStatus = "new"
	LeadStatusContacted   LeadStatus = "contacted"
	LeadStatusQualified   LeadStatus = "qualified"
	LeadStatusNegotiating LeadStatus = "negotiating"
	LeadStatusConverted   LeadStatus = "converted"
	LeadStatusLost        LeadStatus = "lost"
)

// ActorType defines the type of actor performing an action
type ActorType string

const (
	ActorTypeUser   ActorType = "user"   // Broker autenticado
	ActorTypeSystem ActorType = "system" // Job automático
	ActorTypeOwner  ActorType = "owner"  // Owner confirmando via link
)
