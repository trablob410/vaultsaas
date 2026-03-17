package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode"

	"github.com/google/uuid"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	if !emailRe.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain uppercase, lowercase, and digit")
	}
	return nil
}

func ValidateUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format")
	}
	return id, nil
}

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

func ValidatePagination(pageStr, limitStr string) (Pagination, error) {
	page := 1
	limit := 20
	var err error

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return Pagination{}, fmt.Errorf("page must be a positive integer")
		}
	}
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return Pagination{}, fmt.Errorf("limit must be between 1 and 100")
		}
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}, nil
}

var regionCodeRe = regexp.MustCompile(`^[a-z]{2}$`)

func ValidateRegionCode(code string) error {
	if !regionCodeRe.MatchString(code) {
		return fmt.Errorf("region code must be 2 lowercase letters")
	}
	return nil
}
