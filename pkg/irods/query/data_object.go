package query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/iychoi/go-irodsclient/pkg/irods/common"
	"github.com/iychoi/go-irodsclient/pkg/irods/connection"
	"github.com/iychoi/go-irodsclient/pkg/irods/message"
	"github.com/iychoi/go-irodsclient/pkg/irods/types"
	"github.com/iychoi/go-irodsclient/pkg/irods/util"
)

// ListDataObjects lists data objects in the given collection
func ListDataObjects(conn *connection.IRODSConnection, path string) ([]*types.IRODSDataObject, error) {
	if conn == nil || !conn.IsConnected() {
		return nil, fmt.Errorf("connection is nil or disconnected")
	}

	dataObjects := []*types.IRODSDataObject{}

	continueQuery := true
	continueIndex := 0
	for continueQuery {
		// data object
		query := message.NewIRODSMessageQuery(common.MaxQueryRows, continueIndex, 0, 0)
		query.AddSelect(common.ICAT_COLUMN_D_DATA_ID, 1)
		query.AddSelect(common.ICAT_COLUMN_DATA_NAME, 1)
		query.AddSelect(common.ICAT_COLUMN_DATA_SIZE, 1)

		// replica
		query.AddSelect(common.ICAT_COLUMN_DATA_REPL_NUM, 1)
		query.AddSelect(common.ICAT_COLUMN_D_DATA_CHECKSUM, 1)
		query.AddSelect(common.ICAT_COLUMN_D_DATA_STATUS, 1)
		query.AddSelect(common.ICAT_COLUMN_D_RESC_NAME, 1)
		query.AddSelect(common.ICAT_COLUMN_D_DATA_PATH, 1)
		query.AddSelect(common.ICAT_COLUMN_D_RESC_HIER, 1)
		query.AddSelect(common.ICAT_COLUMN_D_CREATE_TIME, 1)
		query.AddSelect(common.ICAT_COLUMN_D_MODIFY_TIME, 1)

		pathCondVal := fmt.Sprintf("= '%s'", path)
		query.AddCondition(common.ICAT_COLUMN_COLL_NAME, pathCondVal)

		queryMessage, err := query.GetMessage()
		if err != nil {
			return nil, fmt.Errorf("Could not make a data object query message - %s", err.Error())
		}

		err = conn.SendMessage(queryMessage)
		if err != nil {
			return nil, fmt.Errorf("Could not send a data object query message - %s", err.Error())
		}

		// Server responds with results
		queryResultMessage, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("Could not receive a data object query result message - %s", err.Error())
		}

		queryResult := message.IRODSMessageQueryResult{}
		err = queryResult.FromMessage(queryResultMessage)
		if err != nil {
			return nil, fmt.Errorf("Could not receive a data object query result message - %s", err.Error())
		}

		if queryResult.RowCount == 0 {
			break
		}

		if queryResult.AttributeCount > len(queryResult.SQLResult) {
			return nil, fmt.Errorf("Could not receive data object attributes - requires %d, but received %d attributes", queryResult.AttributeCount, len(queryResult.SQLResult))
		}

		pagenatedDataObjects := make([]*types.IRODSDataObject, queryResult.RowCount, queryResult.RowCount)

		for attr := 0; attr < queryResult.AttributeCount; attr++ {
			sqlResult := queryResult.SQLResult[attr]
			if len(sqlResult.Values) != queryResult.RowCount {
				return nil, fmt.Errorf("Could not receive data object rows - requires %d, but received %d attributes", queryResult.RowCount, len(sqlResult.Values))
			}

			for row := 0; row < queryResult.RowCount; row++ {
				value := sqlResult.Values[row]

				if pagenatedDataObjects[row] == nil {
					// create a new
					replica := &types.IRODSReplica{
						Number:            -1,
						CheckSum:          "",
						Status:            "",
						ResourceName:      "",
						Path:              "",
						ResourceHierarchy: "",
						CreateTime:        time.Time{},
						ModifyTime:        time.Time{},
					}

					pagenatedDataObjects[row] = &types.IRODSDataObject{
						ID:       -1,
						Path:     "",
						Name:     "",
						Size:     0,
						Replicas: []*types.IRODSReplica{replica},
					}
				}

				switch sqlResult.AttributeIndex {
				case int(common.ICAT_COLUMN_D_DATA_ID):
					objID, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Could not parse data object id - %s", value)
					}
					pagenatedDataObjects[row].ID = objID
				case int(common.ICAT_COLUMN_DATA_NAME):
					pagenatedDataObjects[row].Path = util.MakeIRODSPath(path, value)
					pagenatedDataObjects[row].Name = value
				case int(common.ICAT_COLUMN_DATA_SIZE):
					objSize, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Could not parse data object size - %s", value)
					}
					pagenatedDataObjects[row].Size = objSize
				case int(common.ICAT_COLUMN_DATA_REPL_NUM):
					repNum, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Could not parse data object replica number - %s", value)
					}
					pagenatedDataObjects[row].Replicas[0].Number = repNum
				case int(common.ICAT_COLUMN_D_DATA_CHECKSUM):
					pagenatedDataObjects[row].Replicas[0].CheckSum = value
				case int(common.ICAT_COLUMN_D_DATA_STATUS):
					pagenatedDataObjects[row].Replicas[0].Status = value
				case int(common.ICAT_COLUMN_D_RESC_NAME):
					pagenatedDataObjects[row].Replicas[0].ResourceName = value
				case int(common.ICAT_COLUMN_D_DATA_PATH):
					pagenatedDataObjects[row].Replicas[0].Path = value
				case int(common.ICAT_COLUMN_D_RESC_HIER):
					pagenatedDataObjects[row].Replicas[0].ResourceHierarchy = value
				case int(common.ICAT_COLUMN_D_CREATE_TIME):
					cT, err := util.GetIRODSDateTime(value)
					if err != nil {
						return nil, fmt.Errorf("Could not parse create time - %s", value)
					}
					pagenatedDataObjects[row].Replicas[0].CreateTime = cT
				case int(common.ICAT_COLUMN_D_MODIFY_TIME):
					mT, err := util.GetIRODSDateTime(value)
					if err != nil {
						return nil, fmt.Errorf("Could not parse modify time - %s", value)
					}
					pagenatedDataObjects[row].Replicas[0].ModifyTime = mT
				default:
					// ignore
				}
			}
		}

		dataObjects = append(dataObjects, pagenatedDataObjects...)

		continueIndex = queryResult.ContinueIndex
		if continueIndex == 0 {
			continueQuery = false
		}
	}

	// merge data objects per file
	mergedDataObjectsMap := map[int64]*types.IRODSDataObject{}
	for _, object := range dataObjects {
		existingObj, exists := mergedDataObjectsMap[object.ID]
		if exists {
			// merge
			existingObj.Replicas = append(existingObj.Replicas, object.Replicas[0])
		} else {
			// add
			mergedDataObjectsMap[object.ID] = object
		}
	}

	// convert map to array
	mergedDataObjects := []*types.IRODSDataObject{}
	for _, object := range mergedDataObjectsMap {
		mergedDataObjects = append(mergedDataObjects, object)
	}

	return mergedDataObjects, nil
}

// GetDataObjectMeta returns a data object metadata for the path
func GetDataObjectMeta(conn *connection.IRODSConnection, path string) ([]*types.IRODSMeta, error) {
	if conn == nil || !conn.IsConnected() {
		return nil, fmt.Errorf("connection is nil or disconnected")
	}

	metas := []*types.IRODSMeta{}

	continueQuery := true
	continueIndex := 0
	for continueQuery {
		query := message.NewIRODSMessageQuery(common.MaxQueryRows, continueIndex, 0, 0)
		query.AddSelect(common.ICAT_COLUMN_META_DATA_ATTR_ID, 1)
		query.AddSelect(common.ICAT_COLUMN_META_DATA_ATTR_NAME, 1)
		query.AddSelect(common.ICAT_COLUMN_META_DATA_ATTR_VALUE, 1)
		query.AddSelect(common.ICAT_COLUMN_META_DATA_ATTR_UNITS, 1)

		pathCondVal := fmt.Sprintf("= '%s'", path)
		query.AddCondition(common.ICAT_COLUMN_COLL_NAME, pathCondVal)

		nameCondVal := fmt.Sprintf("= '%s'", path)
		query.AddCondition(common.ICAT_COLUMN_DATA_NAME, nameCondVal)

		queryMessage, err := query.GetMessage()
		if err != nil {
			return nil, fmt.Errorf("Could not make a data object metadata query message - %s", err.Error())
		}

		err = conn.SendMessage(queryMessage)
		if err != nil {
			return nil, fmt.Errorf("Could not send a data object metadata query message - %s", err.Error())
		}

		// Server responds with results
		queryResultMessage, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("Could not receive a data object metadata query result message - %s", err.Error())
		}

		queryResult := message.IRODSMessageQueryResult{}
		err = queryResult.FromMessage(queryResultMessage)
		if err != nil {
			return nil, fmt.Errorf("Could not receive a data object metadata query result message - %s", err.Error())
		}

		if queryResult.RowCount == 0 {
			break
		}

		if queryResult.AttributeCount > len(queryResult.SQLResult) {
			return nil, fmt.Errorf("Could not receive data object metadata attributes - requires %d, but received %d attributes", queryResult.AttributeCount, len(queryResult.SQLResult))
		}

		pagenatedMetas := make([]*types.IRODSMeta, queryResult.RowCount, queryResult.RowCount)

		for attr := 0; attr < queryResult.AttributeCount; attr++ {
			sqlResult := queryResult.SQLResult[attr]
			if len(sqlResult.Values) != queryResult.RowCount {
				return nil, fmt.Errorf("Could not receive data object metadata rows - requires %d, but received %d attributes", queryResult.RowCount, len(sqlResult.Values))
			}

			for row := 0; row < queryResult.RowCount; row++ {
				value := sqlResult.Values[row]

				if pagenatedMetas[row] == nil {
					// create a new
					pagenatedMetas[row] = &types.IRODSMeta{
						AVUID: -1,
						Name:  "",
						Value: "",
						Units: "",
					}
				}

				switch sqlResult.AttributeIndex {
				case int(common.ICAT_COLUMN_META_DATA_ATTR_ID):
					avuID, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("Could not parse data object metadata id - %s", value)
					}
					pagenatedMetas[row].AVUID = avuID
				case int(common.ICAT_COLUMN_META_DATA_ATTR_NAME):
					pagenatedMetas[row].Name = value
				case int(common.ICAT_COLUMN_META_DATA_ATTR_VALUE):
					pagenatedMetas[row].Value = value
				case int(common.ICAT_COLUMN_META_DATA_ATTR_UNITS):
					pagenatedMetas[row].Units = value
				default:
					// ignore
				}
			}

			metas = append(metas, pagenatedMetas...)

			continueIndex = queryResult.ContinueIndex
			if continueIndex == 0 {
				continueQuery = false
			}
		}
	}

	return metas, nil
}