package validator

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	valid := []string{"user@example.com", "user.name+tag@sub.domain.io", "a@b.cc"}
	for _, e := range valid {
		if err := ValidateEmail(e); err != nil {
			t.Errorf("expected %q valid, got error: %v", e, err)
		}
	}
	invalid := []string{"", "notanemail", "@nodomain.com", "user@", "user @example.com"}
	for _, e := range invalid {
		if err := ValidateEmail(e); err == nil {
			t.Errorf("expected %q invalid, got nil", e)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("Correct1Password"); err != nil {
		t.Error(err)
	}
	if err := ValidatePassword("Short1A"); err == nil {
		t.Error("short password should fail")
	}
	if err := ValidatePassword("alllowercase123"); err == nil {
		t.Error("no upper should fail")
	}
	if err := ValidatePassword("ALLUPPERCASE123"); err == nil {
		t.Error("no lower should fail")
	}
	if err := ValidatePassword("NoDigitsHereXyz"); err == nil {
		t.Error("no digit should fail")
	}
}

func TestValidateUUID(t *testing.T) {
	_, err := ValidateUUID("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Error(err)
	}
	_, err = ValidateUUID("not-a-uuid")
	if err == nil {
		t.Error("invalid UUID should fail")
	}
	_, err = ValidateUUID("")
	if err == nil {
		t.Error("empty UUID should fail")
	}
}

func TestValidatePagination(t *testing.T) {
	p, err := ValidatePagination("", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Page != 1 || p.Limit != 20 || p.Offset != 0 {
		t.Errorf("bad defaults: %+v", p)
	}

	p, err = ValidatePagination("3", "10")
	if err != nil {
		t.Fatal(err)
	}
	if p.Page != 3 || p.Limit != 10 || p.Offset != 20 {
		t.Errorf("bad page 3: %+v", p)
	}

	_, err = ValidatePagination("0", "")
	if err == nil {
		t.Error("page 0 should fail")
	}
	_, err = ValidatePagination("", "200")
	if err == nil {
		t.Error("limit 200 should fail")
	}
	_, err = ValidatePagination("abc", "")
	if err == nil {
		t.Error("non-int page should fail")
	}
}

func TestValidateRegionCode(t *testing.T) {
	if err := ValidateRegionCode("us"); err != nil {
		t.Error(err)
	}
	if err := ValidateRegionCode("eu"); err != nil {
		t.Error(err)
	}
	if err := ValidateRegionCode("USA"); err == nil {
		t.Error("3-char should fail")
	}
	if err := ValidateRegionCode("1a"); err == nil {
		t.Error("digit should fail")
	}
	if err := ValidateRegionCode(""); err == nil {
		t.Error("empty should fail")
	}
	if err := ValidateRegionCode(strings.Repeat("a", 50)); err == nil {
		t.Error("too long should fail")
	}
}
