package store_test

import (
	"github.com/bolshagin/xsolla-be-2020/model"
	"github.com/bolshagin/xsolla-be-2020/store"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// Функция для тестирования создания платежной сессии
func TestSessionRepo_Create(t *testing.T) {
	st, teardown := store.TestStore(t, cs)
	defer teardown("sessions")

	s := &model.Session{
		SessionToken: "1231231223123123",
		Amount:       1000,
		Purpose:      "test",
		CreatedAt:    time.Now(),
	}

	err := st.Session().Create(s)

	assert.NoError(t, err)
	assert.NotNil(t, s)
}

// Функция для тестирование метода поиска платежной сессии по переданному токену
func TestSessionRepo_FindByToken(t *testing.T) {
	st, teardown := store.TestStore(t, cs)
	defer teardown("sessions")

	token := "1234567"
	_, err := st.Session().FindByToken(token)
	assert.Error(t, err)

	s := &model.Session{
		Amount:       1000,
		SessionToken: "ca197d71-142c-4bef-abd8-65f0bdd53f0b",
		Purpose:      "test",
		CreatedAt:    time.Now(),
	}
	st.Session().Create(s)

	s, err = st.Session().FindByToken(s.SessionToken)
	assert.NoError(t, err)
	assert.NotNil(t, s)
}
