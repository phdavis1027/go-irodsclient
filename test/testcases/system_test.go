package testcases

import (
	"testing"
	"time"

	"github.com/phdavis1027/go-irodsclient/irods/connection"
	"github.com/phdavis1027/go-irodsclient/irods/fs"
)

func TestSystem(t *testing.T) {
	setup()
	defer shutdown()

	t.Run("test ProcessStat", testProcessStat)
}

func testProcessStat(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	conn := connection.NewIRODSConnection(account, 300*time.Second, "go-irodsclient-test")
	err := conn.Connect()
	failError(t, err)
	defer conn.Disconnect()

	processes, err := fs.StatProcess(conn, "", "")
	failError(t, err)

	for _, process := range processes {
		t.Logf("process - %s\n", process.ToString())
	}
}
