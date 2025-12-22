package query

import (
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// BySurname filters individuals by last name (surname).
// Case-insensitive substring match.
func (fq *FilterQuery) BySurname(surname string) *FilterQuery {
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		indiSurname := strings.ToLower(indi.GetSurname())
		return strings.Contains(indiSurname, strings.ToLower(surname))
	})
}

// BySurnameExact filters individuals by exact last name match (case-insensitive).
func (fq *FilterQuery) BySurnameExact(surname string) *FilterQuery {
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return strings.EqualFold(indi.GetSurname(), surname)
	})
}

// ByGivenName filters individuals by first name (given name).
// Case-insensitive substring match.
func (fq *FilterQuery) ByGivenName(givenName string) *FilterQuery {
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		indiGiven := strings.ToLower(indi.GetGivenName())
		return strings.Contains(indiGiven, strings.ToLower(givenName))
	})
}

// ByGivenNameExact filters individuals by exact first name match (case-insensitive).
func (fq *FilterQuery) ByGivenNameExact(givenName string) *FilterQuery {
	return fq.Where(func(indi *gedcom.IndividualRecord) bool {
		return strings.EqualFold(indi.GetGivenName(), givenName)
	})
}

