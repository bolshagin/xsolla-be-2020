package apiserver_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bolshagin/xsolla-be-2020/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var (
	client     = http.Client{Timeout: 5 * time.Second}
	session    = &model.Session{}
	amount     = 100.0
	purpose    = "test"
	cardNumber = "4111 1111 1111 1111"
	cardCode   = "325"
	cardDate   = "12/23"
)

// Тестирование обработчика эндпойнта /session
// который используется для создании платежной сессии
func Test_HandleSessionCreate(t *testing.T) {
	data := []byte(fmt.Sprintf(`{"amount":%v,"purpose":"%v"}`, amount, purpose))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/session", bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer req.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(session); err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, amount, session.Amount)
	assert.Equal(t, purpose, session.Purpose)
	assert.NotNil(t, session.SessionToken)
}

// Тестирование обработчика эндпойнта /pay
// который используется для выполнения оплаты
func Test_HandlePayment(t *testing.T) {
	data := []byte(fmt.Sprintf(
		`{"session_token":"%v","card_number":"%v","code":"%v","date":"%v"}`,
		session.SessionToken,
		cardNumber,
		cardCode,
		cardDate,
	))

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/pay", bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer req.Body.Close()

	type response struct {
		Payment string `json:"payment"`
	}

	r := &response{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "successful", r.Payment)
}
