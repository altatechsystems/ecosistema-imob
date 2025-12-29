package union

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/altatech/ecosistema-imob/backend/internal/models"
	"github.com/google/uuid"
)

// PropertyPayload represents normalized property data ready for import
type PropertyPayload struct {
	Property    models.Property
	Owner       OwnerPayload
	Photos      []string // URLs from XML
	Title       string   // Título do anúncio (from XML)
	Description string   // Descrição completa do anúncio (from XML)
}

// OwnerPayload represents owner data (may be incomplete/placeholder)
type OwnerPayload struct {
	Name            string
	Phone           string
	Email           string
	Company         string
	OwnerStatus     models.OwnerStatus // incomplete, partial, verified
	EnrichedFromXLS bool
}

// NormalizeProperty converts XMLImovel + optional XLSRecord to PropertyPayload
func NormalizeProperty(xml *XMLImovel, xls *XLSRecord, tenantID string) PropertyPayload {
	now := time.Now()
	propertyID := uuid.New().String()

	// Determine purpose
	purpose := determinePurpose(xml)

	// Determine status
	status := determineStatus(xml)

	// Normalize type
	propertyType := normalizeType(xml.Tipo)

	// Normalize prices (XLS values in thousands)
	salePrice := xml.Valorvenda
	rentalPrice := xml.Valorlocacao
	condoFee := xml.Valorcondominio

	// Determine transaction type and rental info
	var transactionType *models.TransactionType
	var rentalInfo *models.RentalInfo

	if purpose == "both" || purpose == "rent" || purpose == "seasonal" {
		tt := models.TransactionTypeSale // default
		if purpose == "rent" {
			tt = models.TransactionTypeRent
		} else if purpose == "both" {
			tt = models.TransactionTypeBoth
		}
		transactionType = &tt

		// Create rental info if rental price available
		if rentalPrice > 0 {
			rentalInfo = &models.RentalInfo{
				MonthlyRent:        rentalPrice,
				CondoFee:           condoFee,
				TotalMonthlyCost:   rentalPrice + condoFee,
				DepositMonths:      3, // default
				AcceptedGuarantees: []string{"fiador", "caucao"},
				RentalType:         models.RentalTypeTraditional,
				MinRentalPeriod:    12,
				Furnished:          false,
				AcceptsPets:        false,
				ImmediateOccupancy: true,
			}

			if purpose == "seasonal" {
				rentalInfo.RentalType = models.RentalTypeVacation
				rentalInfo.MinRentalPeriod = 1
			}
		}
	}

	// Determine property status
	propertyStatus := models.PropertyStatusAvailable
	if status == "inactive" {
		propertyStatus = models.PropertyStatusUnavailable
	}

	// Determine visibility (default: network for imports)
	visibility := models.PropertyVisibilityNetwork

	// Extract photo URLs from XMLFoto structs
	photoURLs := make([]string, 0, len(xml.Fotos))
	for _, foto := range xml.Fotos {
		if foto.URL != "" {
			photoURLs = append(photoURLs, foto.URL)
		}
	}

	// Determine data completeness
	dataCompleteness := "partial"
	if xml.Titulo != "" && xml.Anuncioparainternet != "" && len(photoURLs) > 0 {
		dataCompleteness = "complete"
	}

	// Create property
	property := models.Property{
		ID:       propertyID,
		TenantID: tenantID,

		// External identifiers (CRITICAL FOR DEDUPLICATION)
		ExternalSource: "union",
		ExternalID:     xml.Codigoimovel, // main dedup key
		Reference:      xml.Referencia,

		// Owner (will be set by ImportService)
		OwnerID: "", // populated by ImportService.createOwner

		// Type and location
		PropertyType: propertyType,
		Street:       xml.Endereco,
		Number:       xml.Numero,
		Complement:   xml.Complemento,
		Neighborhood: xml.Bairro,
		City:         xml.Cidade,
		State:        xml.UnidadeFederativa,
		ZipCode:      xml.CEP,
		Country:      "BR",

		// Characteristics
		Bedrooms:      xml.Dormitorios,
		Bathrooms:     xml.Banheiro2,
		Suites:        xml.Suite,
		ParkingSpaces: xml.Garagem,
		TotalArea:     xml.Areatotal,
		UsableArea:    xml.Areautil,

		// Pricing (use sale price as primary)
		PriceAmount:   salePrice,
		PriceCurrency: "BRL",

		// Status and visibility
		Status:             propertyStatus,
		Visibility:         visibility,
		CoBrokerCommission: 0, // to be defined later

		// Transaction type and rental info
		TransactionType: transactionType,
		RentalInfo:      rentalInfo,

		// Slug (generated from title or reference)
		Slug: createSlug(xml.Titulo, xml.Referencia),

		// Deduplication fields
		Fingerprint:       generateFingerprint(xml),
		PossibleDuplicate: false, // will be set by deduplication service
		DataCompleteness:  dataCompleteness,

		// Timestamps
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Build owner payload
	owner := buildOwnerPayload(xml, xls)

	// Extract title and description from XML
	title := xml.Titulo
	if title == "" {
		title = fmt.Sprintf("Imóvel %s", xml.Referencia)
	}

	description := xml.Anuncioparainternet
	if description == "" {
		description = fmt.Sprintf("Imóvel importado - Ref: %s", xml.Referencia)
	}

	payload := PropertyPayload{
		Property:    property,
		Owner:       owner,
		Photos:      photoURLs,
		Title:       title,
		Description: description,
	}

	return payload
}

// determinePurpose determines property purpose from XML flags
func determinePurpose(xml *XMLImovel) string {
	if xml.Venda == 1 && xml.Locacao == 1 {
		return "both"
	}
	if xml.Locacao == 1 {
		return "rent"
	}
	if xml.Temporada == 1 {
		return "seasonal"
	}
	return "sale" // default
}

// determineStatus determines property status
func determineStatus(xml *XMLImovel) string {
	// If neither sale nor rent is enabled, property is inactive
	if xml.Venda == 0 && xml.Locacao == 0 && xml.Temporada == 0 {
		return "inactive"
	}
	return "available" // default for import
}

// normalizeType normalizes property type to PropertyType enum
func normalizeType(tipo string) models.PropertyType {
	tipo = strings.ToLower(strings.TrimSpace(tipo))

	// Map to PropertyType enum values
	if strings.Contains(tipo, "apartamento") {
		return models.PropertyTypeApartment
	}
	if strings.Contains(tipo, "casa") || strings.Contains(tipo, "sobrado") {
		return models.PropertyTypeHouse
	}
	if strings.Contains(tipo, "terreno") {
		return models.PropertyTypeLand
	}
	if strings.Contains(tipo, "comercial") || strings.Contains(tipo, "sala") ||
		strings.Contains(tipo, "loja") || strings.Contains(tipo, "galpão") ||
		strings.Contains(tipo, "galpao") || strings.Contains(tipo, "ponto") {
		return models.PropertyTypeCommercial
	}

	// Default to apartment for unknown types
	return models.PropertyTypeApartment
}


// buildOwnerPayload builds owner data from XML and optionally XLS
func buildOwnerPayload(xml *XMLImovel, xls *XLSRecord) OwnerPayload {
	owner := OwnerPayload{
		OwnerStatus:     models.OwnerStatusIncomplete,
		EnrichedFromXLS: false,
	}

	// Try to enrich from XLS first (priority)
	if xls != nil && xls.Proprietario != "" {
		owner.Name = xls.Proprietario
		owner.Phone = xls.CelularTelefone
		owner.Email = xls.Email
		owner.Company = xls.Empresa
		owner.EnrichedFromXLS = true

		// Determine owner status
		if owner.Email != "" && owner.Phone != "" {
			owner.OwnerStatus = models.OwnerStatusVerified
		} else if owner.Phone != "" || owner.Email != "" {
			owner.OwnerStatus = models.OwnerStatusPartial
		} else {
			owner.OwnerStatus = models.OwnerStatusPartial // at least has name
		}

		return owner
	}

	// Fallback to XML data (usually incomplete)
	if xml.Captador != "" {
		owner.Name = "Proprietário de " + xml.Referencia + " (Captado por: " + xml.Captador + ")"
		owner.OwnerStatus = models.OwnerStatusPartial
		return owner
	}

	// Placeholder owner
	owner.Name = "Proprietário de " + xml.Referencia
	owner.OwnerStatus = models.OwnerStatusIncomplete

	return owner
}

// generateFingerprint generates deduplication fingerprint
// Based on: normalized address + type + total area
func generateFingerprint(xml *XMLImovel) string {
	// Normalize address components
	street := strings.ToLower(strings.TrimSpace(xml.Endereco))
	number := strings.ToLower(strings.TrimSpace(xml.Numero))
	neighborhood := strings.ToLower(strings.TrimSpace(xml.Bairro))
	city := strings.ToLower(strings.TrimSpace(xml.Cidade))

	// Normalize type
	propertyType := normalizeType(xml.Tipo)

	// Normalize area (round to avoid float precision issues)
	area := fmt.Sprintf("%.0f", xml.Areatotal)

	// Build fingerprint string
	fingerprintStr := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		street, number, neighborhood, city, string(propertyType), area)

	// Hash it
	hash := sha256.Sum256([]byte(fingerprintStr))
	return fmt.Sprintf("%x", hash)
}

// cleanString cleans string (trim, normalize spaces)
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// createSlug creates URL-friendly slug
func createSlug(title, reference string) string {
	if title == "" {
		title = reference
	}

	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()

	// Remove multiple hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
	}

	// Ensure uniqueness by appending reference
	if reference != "" {
		slug = slug + "-" + strings.ToLower(reference)
	}

	return slug
}
