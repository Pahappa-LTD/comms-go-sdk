# Ego SMS sdk for Go

**Version:** 1.1.0

Example:
```go
sdk, err := v1.Authenticate("username", "password")

success, err := sdk.SendSMS("+256772123456", "Test message to single number")

numbers := []string{"+256772123456", "0772123457"}
success, err := sdk.SendSMSWithSenderId(numbers, "Test message to many numbers", "MySenderID")

// Same as SendSMS/SendSMSWithSenderId/SendSMSWithPriority, but returns the full ApiResponse
response, err := sdk.QuerySendSMS("+256772123456", "Test message to single number")

balance, err := sdk.GetBalance()
```

`SendSMS`/`QuerySendSMS` and their `WithSenderId`/`WithPriority` variants default to `models.HIGH` priority. Use `SendSMSFull`/`QuerySendSMSFull` to set both a custom sender ID and priority in one call.