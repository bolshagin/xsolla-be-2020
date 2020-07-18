package store_test

import (
	"fmt"
	"os"
	"testing"
)

var (
	cs string
)

// Вспомогательная функция TestMain, которая запускается перед
// выполнением тестов и конфигурирует подключение к тестовой базе
func TestMain(m *testing.M) {
	var (
		dbname   string = "apipayment_test"
		user     string = "dev"
		password string = "12345"
	)
	cs = fmt.Sprintf("%s:%s@/%s?parseTime=true", user, password, dbname)
	os.Exit(m.Run())
}
