package testcases

import (
	"os"
	"testing"

	"github.com/phdavis1027/go-irodsclient/irods/fs"
	"github.com/phdavis1027/go-irodsclient/irods/session"
	"github.com/phdavis1027/go-irodsclient/irods/types"
	"github.com/phdavis1027/go-irodsclient/irods/util"
	"github.com/phdavis1027/go-irodsclient/test/server"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

var (
	redirectionToResourceAPITestID = xid.New().String()
)

func TestRedirectionToResourceAPI(t *testing.T) {
	setup()
	defer shutdown()

	log.SetLevel(log.DebugLevel)

	makeHomeDir(t, redirectionToResourceAPITestID)

	t.Run("test DownloadDataObjectFromResourceServer", testDownloadDataObjectFromResourceServer)
	t.Run("test UploadDataObjectFromResourceServer", testUploadDataObjectFromResourceServer)
}

func testDownloadDataObjectFromResourceServer(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	sessionConfig := session.NewIRODSSessionConfigWithDefault("go-irodsclient-test")

	sess, err := session.NewIRODSSessionWithAddressResolver(account, sessionConfig, server.AddressResolver)
	failError(t, err)
	defer sess.Release()

	conn, err := sess.AcquireConnection()
	failError(t, err)

	homedir := getHomeDir(redirectionToResourceAPITestID)

	// gen very large file
	filename := "test_large_file.bin"
	fileSize := 100 * 1024 * 1024 // 100MB

	filepath, err := createLocalTestFile(filename, int64(fileSize))
	failError(t, err)

	// upload
	irodsPath := homedir + "/" + filename

	callbackCalled := 0
	callBack := func(current int64, total int64) {
		callbackCalled++
	}

	err = fs.UploadDataObjectParallel(sess, filepath, irodsPath, "", 4, false, callBack)
	failError(t, err)
	assert.Greater(t, callbackCalled, 10) // at least called 10 times

	checksumOriginal, err := util.HashLocalFile(filepath, string(types.ChecksumAlgorithmSHA1))
	failError(t, err)

	err = os.Remove(filepath)
	failError(t, err)

	coll, err := fs.GetCollection(conn, homedir)
	failError(t, err)

	obj, err := fs.GetDataObject(conn, coll, filename)
	failError(t, err)

	assert.NotEmpty(t, obj.ID)
	assert.Equal(t, int64(fileSize), obj.Size)

	// get
	err = fs.DownloadDataObjectFromResourceServer(sess, irodsPath, "", filename, int64(fileSize), callBack)
	failError(t, err)

	checksumNew, err := util.HashLocalFile(filename, string(types.ChecksumAlgorithmSHA1))
	failError(t, err)

	err = os.Remove(filename)
	failError(t, err)

	// delete
	err = fs.DeleteDataObject(conn, irodsPath, true)
	failError(t, err)

	assert.Equal(t, checksumOriginal, checksumNew)

	sess.ReturnConnection(conn)
}

func testUploadDataObjectFromResourceServer(t *testing.T) {
	account := GetTestAccount()

	account.ClientServerNegotiation = false

	sessionConfig := session.NewIRODSSessionConfigWithDefault("go-irodsclient-test")

	sess, err := session.NewIRODSSessionWithAddressResolver(account, sessionConfig, server.AddressResolver)
	failError(t, err)
	defer sess.Release()

	conn, err := sess.AcquireConnection()
	failError(t, err)

	homedir := getHomeDir(redirectionToResourceAPITestID)

	// gen very large file
	filename := "test_large_file.bin"
	fileSize := 100 * 1024 * 1024 // 100MB

	filepath, err := createLocalTestFile(filename, int64(fileSize))
	failError(t, err)

	// upload
	irodsPath := homedir + "/" + filename

	callbackCalled := 0
	callBack := func(current int64, total int64) {
		callbackCalled++
	}

	err = fs.UploadDataObjectToResourceServer(sess, filepath, irodsPath, "", false, callBack)
	failError(t, err)
	assert.Greater(t, callbackCalled, 10) // at least called 10 times

	checksumOriginal, err := util.HashLocalFile(filepath, string(types.ChecksumAlgorithmSHA1))
	failError(t, err)

	err = os.Remove(filepath)
	failError(t, err)

	coll, err := fs.GetCollection(conn, homedir)
	failError(t, err)

	obj, err := fs.GetDataObject(conn, coll, filename)
	failError(t, err)

	assert.NotEmpty(t, obj.ID)
	assert.Equal(t, int64(fileSize), obj.Size)

	// get
	err = fs.DownloadDataObjectFromResourceServer(sess, irodsPath, "", filename, int64(fileSize), callBack)
	failError(t, err)

	checksumNew, err := util.HashLocalFile(filename, string(types.ChecksumAlgorithmSHA1))
	failError(t, err)

	err = os.Remove(filename)
	failError(t, err)

	// delete
	err = fs.DeleteDataObject(conn, irodsPath, true)
	failError(t, err)

	assert.Equal(t, checksumOriginal, checksumNew)

	sess.ReturnConnection(conn)
}
