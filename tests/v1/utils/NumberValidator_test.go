package utils_test

import (
	"testing"

	"github.com/Pahappa-LTD/CommsGoSDK/src/v1/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateNumbersWithValidNumbers(t *testing.T) {
	numbers := []string{"+256772123456", "0772123457", "256772123458"}
	expected := []string{"256772123456", "256772123457", "256772123458"}
	assert.ElementsMatch(t, expected, utils.ValidateNumbers(numbers))
}

func TestValidateNumbersWithInvalidNumbers(t *testing.T) {
	numbers := []string{"123", "not a number", "077212345"}
	assert.Empty(t, utils.ValidateNumbers(numbers))
}

func TestValidateNumbersWithMixedNumbers(t *testing.T) {
	numbers := []string{"+256772123456", "123", "0772123457"}
	expected := []string{"256772123456", "256772123457"}
	assert.ElementsMatch(t, expected, utils.ValidateNumbers(numbers))
}

func TestValidateNumbersWithEmptyArray(t *testing.T) {
	assert.Empty(t, utils.ValidateNumbers([]string{}))
}

func TestValidateNumbersWithDuplicateNumbers(t *testing.T) {
	numbers := []string{"+256772123456", "0772123456"}
	expected := []string{"256772123456"}
	assert.ElementsMatch(t, expected, utils.ValidateNumbers(numbers))
}

func TestValidateNumbersWithDifferentFormats(t *testing.T) {
	numbers := []string{"+256 772 123 456", "0772-123-457", " 256772123458 "}
	expected := []string{"256772123456", "256772123457", "256772123458"}
	assert.ElementsMatch(t, expected, utils.ValidateNumbers(numbers))
}
