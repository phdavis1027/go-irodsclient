package fs

import (
	irods_fs "github.com/phdavis1027/go-irodsclient/irods/fs"
	"github.com/phdavis1027/go-irodsclient/irods/types"
)

// ListProcesses lists all processes
func (fs *FileSystem) ListProcesses(address string, zone string) ([]*types.IRODSProcess, error) {
	conn, err := fs.metaSession.AcquireConnection()
	if err != nil {
		return nil, err
	}
	defer fs.metaSession.ReturnConnection(conn)

	processes, err := irods_fs.StatProcess(conn, address, zone)
	if err != nil {
		return nil, err
	}

	return processes, nil
}

// ListAllProcesses lists all processes
func (fs *FileSystem) ListAllProcesses() ([]*types.IRODSProcess, error) {
	return fs.ListProcesses("", "")
}
