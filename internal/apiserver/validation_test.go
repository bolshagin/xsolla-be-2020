package apiserver_test

import (
	"github.com/bolshagin/xsolla-be-2020/internal/apiserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Тестирование функции по проверке номера карты
func TestIsCreditCard(t *testing.T) {
	testCases := []struct {
		name       string
		cardNumber string
		isValid    bool
	}{
		{
			name:       "mastercard valid",
			cardNumber: "5500 0000 0000 0004",
			isValid:    true,
		},
		{
			name:       "mastercard invalid",
			cardNumber: "5500 0000 0000 0004 1234",
			isValid:    false,
		},
		{
			name:       "only text",
			cardNumber: "asdasdasdasdasd",
			isValid:    false,
		},
		{
			name:       "empty string",
			cardNumber: "",
			isValid:    false,
		},
		{
			name:       "visa valid",
			cardNumber: "4111 1111 1111 1111",
			isValid:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isValid, apiserver.IsCreditCard(tc.cardNumber))
		})
	}
}

// Тестирование функции по проверке параметра платежа "Дата"
func TestIsCardDate(t *testing.T) {
	testCases := []struct {
		name    string
		date    string
		isValid bool
	}{
		{
			name:    "valid code",
			date:    "12/20",
			isValid: true,
		},
		{
			name:    "invalid code 13 month",
			date:    "13/20",
			isValid: false,
		},
		{
			name:    "invalid code zeros",
			date:    "00/00",
			isValid: false,
		},
		{
			name:    "empty",
			date:    "",
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isValid, apiserver.IsCardDate(tc.date))
		})
	}
}

// Тестирование функции по проверке параметра платежа "CVC/CVV"
func TestIsCardCode(t *testing.T) {
	testCases := []struct {
		name    string
		code    string
		isValid bool
	}{
		{
			name:    "valid code",
			code:    "056",
			isValid: true,
		},
		{
			name:    "invalid code letters",
			code:    "asd",
			isValid: false,
		},
		{
			name:    "invalid code one number",
			code:    "1",
			isValid: false,
		},
		{
			name:    "empty",
			code:    "",
			isValid: false,
		},
		{
			name:    "invalid more 3 symbols",
			code:    "1234",
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isValid, apiserver.IsCardCode(tc.code))
		})
	}
}
