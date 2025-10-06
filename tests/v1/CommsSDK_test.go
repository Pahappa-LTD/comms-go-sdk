package v1_test

import (
	"fmt"
	"testing"

	v1 "github.com/Pahappa-LTD/CommsGoSDK/src/v1"
	"github.com/Pahappa-LTD/CommsGoSDK/src/v1/models"
	"github.com/stretchr/testify/assert"
)

func TestSendSMSToSingleNumber(t *testing.T) {
	v1.UseSandBox()
	sdk, err := v1.Authenticate("agabu-idaniel", "dcfa634d7936ec699a3b26f6cd924801b09b285a31949f99")
	assert.NoError(t, err)

	success, err := sdk.SendSMS("+256772123456", "Test message")
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestSendSMSToMultipleNumbers(t *testing.T) {
	v1.UseSandBox()
	sdk, err := v1.Authenticate("agabu-idaniel", "dcfa634d7936ec699a3b26f6cd924801b09b285a31949f99")
	assert.NoError(t, err)

	numbers := []string{"+256772123456", "0772123457"}
	success, err := sdk.SendSMSWithSenderId(numbers, "Test message", "MySenderID")
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestSendSMSWithShortNumberLength(t *testing.T) {
	v1.UseSandBox()
	sdk, err := v1.Authenticate("agabu-idaniel", "dcfa634d7936ec699a3b26f6cd924801b09b285a31949f99")
	assert.NoError(t, err)

	res, err := sdk.QuerySendSMS("123", "Test message", "EgoSMS", models.HIGHEST)
	assert.Error(t, err)

	fmt.Println(res)
}

func TestSendSMSWithCustomMessagePriority(t *testing.T) {
	v1.UseSandBox()
	sdk, err := v1.Authenticate("agabu-idaniel", "dcfa634d7936ec699a3b26f6cd924801b09b285a31949f99")
	assert.NoError(t, err)

	success, err := sdk.SendSMSWithPriority("+256772123456", "Test message", models.LOW)
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestSendSMSWithInvalidCredentials(t *testing.T) {
	v1.UseSandBox()
	_, err := v1.Authenticate("invalid_user", "invalid_password")
	assert.Error(t, err)
}

func TestCheckBalanceAfterSendingSMS(t *testing.T) {
	v1.UseSandBox()
	sdk, err := v1.Authenticate("agabu-idaniel", "dcfa634d7936ec699a3b26f6cd924801b09b285a31949f99")
	assert.NoError(t, err)

	balanceBefore, err := sdk.GetBalance()
	assert.NoError(t, err)

	_, err = sdk.SendSMS("+256772123456", "Test message")
	assert.NoError(t, err)

	balanceAfter, err := sdk.GetBalance()
	assert.NoError(t, err)

	assert.Less(t, *balanceAfter, *balanceBefore)
}
