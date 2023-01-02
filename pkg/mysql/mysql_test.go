package mysql_test

import (
	"testing"

	"github.com/limingyao/excellent-go/pkg/mysql"
)

func TestNewFromEnv(t *testing.T) {
	_ = mysql.NewFromEnv()
}
