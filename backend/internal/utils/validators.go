package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================================
// CRECI Validation (Conselho Regional de Corretores de Imóveis)
// ============================================================================

// ValidateCRECI validates the Brazilian CRECI format
// Format: XXXXX-F/UF or XXXXX/UF
// Where:
// - XXXXX = 5 digits (may have leading zeros)
// - F = optional letter (J for legal entity, empty for individual)
// - UF = state abbreviation (2 uppercase letters)
//
// Examples:
// - "12345-J/SP" (Legal Entity - São Paulo)
// - "00123/RJ" (Individual - Rio de Janeiro)
func ValidateCRECI(creci string) error {
	if creci == "" {
		return errors.New("CRECI é obrigatório")
	}

	// Regex: 5 digits + optional hyphen + optional letter + slash + UF
	// Example: 12345-J/SP or 00123/RJ
	creciRegex := regexp.MustCompile(`^\d{5}(-[A-Z])?/[A-Z]{2}$`)

	if !creciRegex.MatchString(strings.TrimSpace(creci)) {
		return errors.New("CRECI inválido. Formato esperado: XXXXX-F/UF ou XXXXX/UF (ex: 12345-J/SP)")
	}

	return nil
}

// NormalizeCRECI normalizes CRECI by removing spaces and converting to uppercase
func NormalizeCRECI(creci string) string {
	return strings.ToUpper(strings.TrimSpace(creci))
}

// ExtractCRECIState extracts the state (UF) from CRECI
// Example: "12345-J/SP" -> "SP"
func ExtractCRECIState(creci string) string {
	parts := strings.Split(creci, "/")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// ============================================================================
// CPF Validation (Cadastro de Pessoa Física)
// ============================================================================

// ValidateCPF validates Brazilian CPF
// Accepts formats: XXX.XXX.XXX-XX or XXXXXXXXXXX
func ValidateCPF(cpf string) error {
	// Remove punctuation
	cpf = regexp.MustCompile(`[^\d]`).ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return errors.New("CPF deve ter 11 dígitos")
	}

	// Check for known invalid CPFs (all digits the same)
	if cpf == "00000000000" || cpf == "11111111111" || cpf == "22222222222" ||
		cpf == "33333333333" || cpf == "44444444444" || cpf == "55555555555" ||
		cpf == "66666666666" || cpf == "77777777777" || cpf == "88888888888" ||
		cpf == "99999999999" {
		return errors.New("CPF inválido")
	}

	// Validate check digits
	if !validateCPFCheckDigit(cpf) {
		return errors.New("CPF inválido")
	}

	return nil
}

// validateCPFCheckDigit validates CPF check digits
func validateCPFCheckDigit(cpf string) bool {
	// First check digit
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (10 - i)
	}
	remainder := sum % 11
	checkDigit1 := 0
	if remainder >= 2 {
		checkDigit1 = 11 - remainder
	}
	if checkDigit1 != int(cpf[9]-'0') {
		return false
	}

	// Second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (11 - i)
	}
	remainder = sum % 11
	checkDigit2 := 0
	if remainder >= 2 {
		checkDigit2 = 11 - remainder
	}
	if checkDigit2 != int(cpf[10]-'0') {
		return false
	}

	return true
}

// NormalizeCPF formats CPF in the standard XXX.XXX.XXX-XX format
func NormalizeCPF(cpf string) string {
	cpf = regexp.MustCompile(`[^\d]`).ReplaceAllString(cpf, "")
	if len(cpf) != 11 {
		return cpf
	}
	return fmt.Sprintf("%s.%s.%s-%s", cpf[0:3], cpf[3:6], cpf[6:9], cpf[9:11])
}

// ============================================================================
// CNPJ Validation (Cadastro Nacional de Pessoa Jurídica)
// ============================================================================

// ValidateCNPJ validates Brazilian CNPJ
// Accepts formats: XX.XXX.XXX/XXXX-XX or XXXXXXXXXXXXXX
func ValidateCNPJ(cnpj string) error {
	// Remove punctuation
	cnpj = regexp.MustCompile(`[^\d]`).ReplaceAllString(cnpj, "")

	if len(cnpj) != 14 {
		return errors.New("CNPJ deve ter 14 dígitos")
	}

	// Check for known invalid CNPJs (all digits the same)
	if cnpj == "00000000000000" || cnpj == "11111111111111" || cnpj == "22222222222222" ||
		cnpj == "33333333333333" || cnpj == "44444444444444" || cnpj == "55555555555555" ||
		cnpj == "66666666666666" || cnpj == "77777777777777" || cnpj == "88888888888888" ||
		cnpj == "99999999999999" {
		return errors.New("CNPJ inválido")
	}

	// Validate check digits
	if !validateCNPJCheckDigit(cnpj) {
		return errors.New("CNPJ inválido")
	}

	return nil
}

// validateCNPJCheckDigit validates CNPJ check digits
func validateCNPJCheckDigit(cnpj string) bool {
	// First check digit
	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights1[i]
	}
	remainder := sum % 11
	checkDigit1 := 0
	if remainder >= 2 {
		checkDigit1 = 11 - remainder
	}
	if checkDigit1 != int(cnpj[12]-'0') {
		return false
	}

	// Second check digit
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights2[i]
	}
	remainder = sum % 11
	checkDigit2 := 0
	if remainder >= 2 {
		checkDigit2 = 11 - remainder
	}
	if checkDigit2 != int(cnpj[13]-'0') {
		return false
	}

	return true
}

// NormalizeCNPJ formats CNPJ in the standard XX.XXX.XXX/XXXX-XX format
func NormalizeCNPJ(cnpj string) string {
	cnpj = regexp.MustCompile(`[^\d]`).ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 {
		return cnpj
	}
	return fmt.Sprintf("%s.%s.%s/%s-%s", cnpj[0:2], cnpj[2:5], cnpj[5:8], cnpj[8:12], cnpj[12:14])
}

// ============================================================================
// Phone Validation (Brazilian Phone Numbers)
// ============================================================================

// ValidatePhoneBR validates Brazilian phone numbers
// Accepts formats: (XX) XXXXX-XXXX, (XX) XXXX-XXXX, XXXXXXXXXXX
// Mobile: 11 digits (DDD + 9 + 8 digits)
// Landline: 10 digits (DDD + 8 digits)
func ValidatePhoneBR(phone string) error {
	// Remove punctuation
	phone = regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Mobile: 11 digits (DDD + 9 + 8 digits)
	// Landline: 10 digits (DDD + 8 digits)
	if len(phone) != 10 && len(phone) != 11 {
		return errors.New("telefone brasileiro deve ter 10 ou 11 dígitos")
	}

	// Validate DDD (area code: 11 to 99)
	ddd := phone[0:2]
	dddInt, err := strconv.Atoi(ddd)
	if err != nil || dddInt < 11 || dddInt > 99 {
		return errors.New("DDD inválido")
	}

	return nil
}

// NormalizePhoneBR formats Brazilian phone numbers
// Mobile: (XX) XXXXX-XXXX
// Landline: (XX) XXXX-XXXX
func NormalizePhoneBR(phone string) string {
	phone = regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	if len(phone) == 11 {
		// Mobile: (XX) XXXXX-XXXX
		return fmt.Sprintf("(%s) %s-%s", phone[0:2], phone[2:7], phone[7:11])
	} else if len(phone) == 10 {
		// Landline: (XX) XXXX-XXXX
		return fmt.Sprintf("(%s) %s-%s", phone[0:2], phone[2:6], phone[6:10])
	}

	return phone
}

// ============================================================================
// Email Validation
// ============================================================================

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email é obrigatório")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return errors.New("email inválido")
	}

	return nil
}

// NormalizeEmail converts email to lowercase and removes spaces
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
