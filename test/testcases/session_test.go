package testcases

import (
	"testing"

	"github.com/phdavis1027/go-irodsclient/irods/connection"
	"github.com/phdavis1027/go-irodsclient/irods/fs"
	"github.com/phdavis1027/go-irodsclient/irods/session"
	"github.com/phdavis1027/go-irodsclient/irods/types"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

var (
	fsSessionTestID = xid.New().String()
)

func TestSession(t *testing.T) {
	setup()
	defer shutdown()

	makeHomeDir(t, fsSessionTestID)

	t.Run("test Session", testSession)
	t.Run("test many Connections", testManyConnections)
	t.Run("test Connection Metrics", testConnectionMetrics)
}

func testSession(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false
	account.CSNegotiationPolicy = types.CSNegotiationDontCare

	sessionConfig := session.NewIRODSSessionConfigWithDefault("go-irodsclient-test")

	sess, err := session.NewIRODSSession(account, sessionConfig)
	failError(t, err)
	defer sess.Release()

	// first
	conn, err := sess.AcquireConnection()
	failError(t, err)

	homedir := getHomeDir(fsSessionTestID)

	collection, err := fs.GetCollection(conn, homedir)
	failError(t, err)

	assert.Equal(t, homedir, collection.Path)
	assert.NotEmpty(t, collection.ID)

	err = sess.ReturnConnection(conn)
	failError(t, err)
}

func testManyConnections(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false
	account.CSNegotiationPolicy = types.CSNegotiationDontCare

	sessionConfig := session.NewIRODSSessionConfigWithDefault("go-irodsclient-test")

	sess, err := session.NewIRODSSession(account, sessionConfig)
	failError(t, err)
	defer sess.Release()

	homedir := getHomeDir(fsSessionTestID)

	connections := []*connection.IRODSConnection{}

	for i := 0; i < 30; i++ {
		conn, err := sess.AcquireConnection()
		failError(t, err)

		collection, err := fs.GetCollection(conn, homedir)
		failError(t, err)

		connections = append(connections, conn)

		assert.Equal(t, homedir, collection.Path)
		assert.NotEmpty(t, collection.ID)
	}

	assert.Equal(t, sessionConfig.ConnectionMax, sess.ConnectionTotal())

	for _, conn := range connections {
		err = sess.ReturnConnection(conn)
		failError(t, err)
	}
}

func testConnectionMetrics(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false
	account.CSNegotiationPolicy = types.CSNegotiationDontCare

	sessionConfig := session.NewIRODSSessionConfigWithDefault("go-irodsclient-test")

	sess, err := session.NewIRODSSession(account, sessionConfig)
	failError(t, err)
	defer sess.Release()

	metrics := sess.GetMetrics()
	assert.Equal(t, uint64(sessionConfig.ConnectionInitNumber), metrics.GetConnectionsOpened())
	assert.Equal(t, uint64(0), metrics.GetConnectionsOccupied())

	homedir := getHomeDir(fsSessionTestID)

	connections := []*connection.IRODSConnection{}

	for i := 0; i < 30; i++ {
		conn, err := sess.AcquireConnection()
		failError(t, err)

		collection, err := fs.GetCollection(conn, homedir)
		failError(t, err)

		connections = append(connections, conn)

		assert.Equal(t, homedir, collection.Path)
		assert.NotEmpty(t, collection.ID)
	}

	assert.Equal(t, sessionConfig.ConnectionMax, sess.ConnectionTotal())
	assert.Equal(t, uint64(sessionConfig.ConnectionMax), metrics.GetConnectionsOpened())
	assert.Equal(t, uint64(sessionConfig.ConnectionMax), metrics.GetConnectionsOccupied())

	for _, conn := range connections {
		err = sess.ReturnConnection(conn)
		failError(t, err)
	}

	assert.Equal(t, uint64(sessionConfig.ConnectionMaxIdle), metrics.GetConnectionsOpened())
	assert.Equal(t, uint64(0), metrics.GetConnectionsOccupied())
}
