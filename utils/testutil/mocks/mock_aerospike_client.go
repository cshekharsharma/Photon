package mocks

import (
	aerov8 "github.com/aerospike/aerospike-client-go/v8"
	"github.com/stretchr/testify/mock"
)

// ---- Mocking of aerospike client --------- //

type MockAerospikeClient struct {
	mock.Mock
}

// Connection management
func (m *MockAerospikeClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAerospikeClient) Close() {
	m.Called()
}

// Policies Getters
func (m *MockAerospikeClient) GetDefaultPolicy() *aerov8.BasePolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.BasePolicy)
}

func (m *MockAerospikeClient) GetDefaultBatchPolicy() *aerov8.BatchPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.BatchPolicy)
}

func (m *MockAerospikeClient) GetDefaultBatchReadPolicy() *aerov8.BatchReadPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.BatchReadPolicy)
}

func (m *MockAerospikeClient) GetDefaultBatchWritePolicy() *aerov8.BatchWritePolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.BatchWritePolicy)
}

func (m *MockAerospikeClient) GetDefaultBatchDeletePolicy() *aerov8.BatchDeletePolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.BatchDeletePolicy)
}

func (m *MockAerospikeClient) GetDefaultBatchUDFPolicy() *aerov8.BatchUDFPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.BatchUDFPolicy)
}

func (m *MockAerospikeClient) GetDefaultWritePolicy() *aerov8.WritePolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.WritePolicy)
}

func (m *MockAerospikeClient) GetDefaultScanPolicy() *aerov8.ScanPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.ScanPolicy)
}

func (m *MockAerospikeClient) GetDefaultQueryPolicy() *aerov8.QueryPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.QueryPolicy)
}

func (m *MockAerospikeClient) GetDefaultAdminPolicy() *aerov8.AdminPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.AdminPolicy)
}

func (m *MockAerospikeClient) GetDefaultInfoPolicy() *aerov8.InfoPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.InfoPolicy)
}

func (m *MockAerospikeClient) GetDefaultTxnVerifyPolicy() *aerov8.TxnVerifyPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.TxnVerifyPolicy)
}

func (m *MockAerospikeClient) GetDefaultTxnRollPolicy() *aerov8.TxnRollPolicy {
	args := m.Called()
	return args.Get(0).(*aerov8.TxnRollPolicy)
}

// Policies Setters
func (m *MockAerospikeClient) SetDefaultPolicy(policy *aerov8.BasePolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultBatchPolicy(policy *aerov8.BatchPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultBatchWritePolicy(policy *aerov8.BatchWritePolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultBatchReadPolicy(policy *aerov8.BatchReadPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultBatchDeletePolicy(policy *aerov8.BatchDeletePolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultBatchUDFPolicy(policy *aerov8.BatchUDFPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultWritePolicy(policy *aerov8.WritePolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultScanPolicy(policy *aerov8.ScanPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultQueryPolicy(policy *aerov8.QueryPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultAdminPolicy(policy *aerov8.AdminPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultInfoPolicy(policy *aerov8.InfoPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultTxnVerifyPolicy(policy *aerov8.TxnVerifyPolicy) {
	m.Called(policy)
}

func (m *MockAerospikeClient) SetDefaultTxnRollPolicy(policy *aerov8.TxnRollPolicy) {
	m.Called(policy)
}

// Cluster/Nodes
func (m *MockAerospikeClient) GetNodes() []*aerov8.Node {
	args := m.Called()
	return args.Get(0).([]*aerov8.Node)
}

func (m *MockAerospikeClient) GetNodeNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAerospikeClient) Cluster() *aerov8.Cluster {
	args := m.Called()
	return args.Get(0).(*aerov8.Cluster)
}

// CRUD
func (m *MockAerospikeClient) Put(policy *aerov8.WritePolicy, key *aerov8.Key, binMap aerov8.BinMap) aerov8.Error {
	args := m.Called(policy, key, binMap)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) PutBins(policy *aerov8.WritePolicy, key *aerov8.Key, bins ...*aerov8.Bin) aerov8.Error {
	args := m.Called(policy, key, bins)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) Get(policy *aerov8.BasePolicy, key *aerov8.Key, binNames ...string) (*aerov8.Record, aerov8.Error) {
	args := m.Called(policy, key, binNames)
	rec := args.Get(0).(*aerov8.Record)
	err := args.Get(1)
	if err == nil {
		return rec, nil
	}
	return rec, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) GetHeader(policy *aerov8.BasePolicy, key *aerov8.Key) (*aerov8.Record, aerov8.Error) {
	args := m.Called(policy, key)
	rec := args.Get(0).(*aerov8.Record)
	err := args.Get(1)
	if err == nil {
		return rec, nil
	}
	return rec, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) Delete(policy *aerov8.WritePolicy, key *aerov8.Key) (bool, aerov8.Error) {
	args := m.Called(policy, key)
	err := args.Get(1)
	if err == nil {
		return args.Bool(0), nil
	}
	return args.Bool(0), err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) Touch(policy *aerov8.WritePolicy, key *aerov8.Key) aerov8.Error {
	args := m.Called(policy, key)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) Exists(policy *aerov8.BasePolicy, key *aerov8.Key) (bool, aerov8.Error) {
	args := m.Called(policy, key)
	err := args.Get(1)
	if err == nil {
		return args.Bool(0), nil
	}
	return args.Bool(0), err.(*aerov8.AerospikeError)
}

// Batch
func (m *MockAerospikeClient) BatchGet(policy *aerov8.BatchPolicy, keys []*aerov8.Key, binNames ...string) ([]*aerov8.Record, aerov8.Error) {
	args := m.Called(policy, keys, binNames)
	records := args.Get(0).([]*aerov8.Record)
	err := args.Get(1)
	if err == nil {
		return records, nil
	}
	return records, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchGetOperate(policy *aerov8.BatchPolicy, keys []*aerov8.Key, ops ...*aerov8.Operation) ([]*aerov8.Record, aerov8.Error) {
	args := m.Called(policy, keys, ops)
	records := args.Get(0).([]*aerov8.Record)
	err := args.Get(1)
	if err == nil {
		return records, nil
	}
	return records, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchGetComplex(policy *aerov8.BatchPolicy, records []*aerov8.BatchRead) aerov8.Error {
	args := m.Called(policy, records)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchGetHeader(policy *aerov8.BatchPolicy, keys []*aerov8.Key) ([]*aerov8.Record, aerov8.Error) {
	args := m.Called(policy, keys)
	records := args.Get(0).([]*aerov8.Record)
	err := args.Get(1)
	if err == nil {
		return records, nil
	}
	return records, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchExists(policy *aerov8.BatchPolicy, keys []*aerov8.Key) ([]bool, aerov8.Error) {
	args := m.Called(policy, keys)
	list := args.Get(0).([]bool)
	err := args.Get(1)
	if err == nil {
		return list, nil
	}
	return list, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchDelete(policy *aerov8.BatchPolicy, deletePolicy *aerov8.BatchDeletePolicy, keys []*aerov8.Key) ([]*aerov8.BatchRecord, aerov8.Error) {
	args := m.Called(policy, deletePolicy, keys)
	records := args.Get(0).([]*aerov8.BatchRecord)
	err := args.Get(1)
	if err == nil {
		return records, nil
	}
	return records, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchOperate(policy *aerov8.BatchPolicy, records []aerov8.BatchRecordIfc) aerov8.Error {
	args := m.Called(policy, records)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) BatchExecute(policy *aerov8.BatchPolicy, udfPolicy *aerov8.BatchUDFPolicy, keys []*aerov8.Key, packageName, functionName string, argsUDF ...aerov8.Value) ([]*aerov8.BatchRecord, aerov8.Error) {
	args := m.Called(policy, udfPolicy, keys, packageName, functionName, argsUDF)
	records := args.Get(0).([]*aerov8.BatchRecord)
	err := args.Get(1)
	if err == nil {
		return records, nil
	}
	return records, err.(*aerov8.AerospikeError)
}

// Operate
func (m *MockAerospikeClient) Operate(policy *aerov8.WritePolicy, key *aerov8.Key, operations ...*aerov8.Operation) (*aerov8.Record, aerov8.Error) {
	args := m.Called(policy, key, operations)
	rec := args.Get(0).(*aerov8.Record)
	err := args.Get(1)
	if err == nil {
		return rec, nil
	}
	return rec, err.(*aerov8.AerospikeError)
}

// Scan
func (m *MockAerospikeClient) ScanAll(policy *aerov8.ScanPolicy, namespace, setName string, binNames ...string) (*aerov8.Recordset, aerov8.Error) {
	args := m.Called(policy, namespace, setName, binNames)
	recordSet := args.Get(0).(*aerov8.Recordset)
	err := args.Get(1)
	if err == nil {
		return recordSet, nil
	}
	return recordSet, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) ScanPartitions(policy *aerov8.ScanPolicy, partitionFilter *aerov8.PartitionFilter, namespace, setName string, binNames ...string) (*aerov8.Recordset, aerov8.Error) {
	args := m.Called(policy, partitionFilter, namespace, setName, binNames)
	recordSet := args.Get(0).(*aerov8.Recordset)
	err := args.Get(1)
	if err == nil {
		return recordSet, nil
	}
	return recordSet, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) ScanNode(policy *aerov8.ScanPolicy, node *aerov8.Node, namespace, setName string, binNames ...string) (*aerov8.Recordset, aerov8.Error) {
	args := m.Called(policy, node, namespace, setName, binNames)
	recordSet := args.Get(0).(*aerov8.Recordset)
	err := args.Get(1)
	if err == nil {
		return recordSet, nil
	}
	return recordSet, err.(*aerov8.AerospikeError)
}

// Query
func (m *MockAerospikeClient) Query(policy *aerov8.QueryPolicy, statement *aerov8.Statement) (*aerov8.Recordset, aerov8.Error) {
	args := m.Called(policy, statement)
	recordSet := args.Get(0).(*aerov8.Recordset)
	err := args.Get(1)
	if err == nil {
		return recordSet, nil
	}
	return recordSet, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryPartitions(policy *aerov8.QueryPolicy, statement *aerov8.Statement, partitionFilter *aerov8.PartitionFilter) (*aerov8.Recordset, aerov8.Error) {
	args := m.Called(policy, statement, partitionFilter)
	recordSet := args.Get(0).(*aerov8.Recordset)
	err := args.Get(1)
	if err == nil {
		return recordSet, nil
	}
	return recordSet, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryNode(policy *aerov8.QueryPolicy, node *aerov8.Node, statement *aerov8.Statement) (*aerov8.Recordset, aerov8.Error) {
	args := m.Called(policy, node, statement)
	recordSet := args.Get(0).(*aerov8.Recordset)
	err := args.Get(1)
	if err == nil {
		return recordSet, nil
	}
	return recordSet, err.(*aerov8.AerospikeError)
}

// UDF
func (m *MockAerospikeClient) Execute(policy *aerov8.WritePolicy, key *aerov8.Key, packageName, functionName string, argsExec ...aerov8.Value) (any, aerov8.Error) {
	args := m.Called(policy, key, packageName, functionName, argsExec)
	result := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return result, nil
	}
	return result, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) RegisterUDF(policy *aerov8.WritePolicy, udfBody []byte, serverPath string, language aerov8.Language) (*aerov8.RegisterTask, aerov8.Error) {
	args := m.Called(policy, udfBody, serverPath, language)
	task := args.Get(0).(*aerov8.RegisterTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) RegisterUDFFromFile(policy *aerov8.WritePolicy, clientPath, serverPath string, language aerov8.Language) (*aerov8.RegisterTask, aerov8.Error) {
	args := m.Called(policy, clientPath, serverPath, language)
	task := args.Get(0).(*aerov8.RegisterTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) RemoveUDF(policy *aerov8.WritePolicy, udfName string) (*aerov8.RemoveTask, aerov8.Error) {
	args := m.Called(policy, udfName)
	task := args.Get(0).(*aerov8.RemoveTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) ListUDF(policy *aerov8.BasePolicy) ([]*aerov8.UDF, aerov8.Error) {
	args := m.Called(policy)
	list := args.Get(0).([]*aerov8.UDF)
	err := args.Get(1)
	if err == nil {
		return list, nil
	}
	return list, err.(*aerov8.AerospikeError)
}

// Execute UDF Query
func (m *MockAerospikeClient) ExecuteUDF(policy *aerov8.QueryPolicy, statement *aerov8.Statement, packageName, functionName string, functionArgs ...aerov8.Value) (*aerov8.ExecuteTask, aerov8.Error) {
	args := m.Called(policy, statement, packageName, functionName, functionArgs)
	task := args.Get(0).(*aerov8.ExecuteTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) ExecuteUDFNode(policy *aerov8.QueryPolicy, node *aerov8.Node, statement *aerov8.Statement, packageName, functionName string, functionArgs ...aerov8.Value) (*aerov8.ExecuteTask, aerov8.Error) {
	args := m.Called(policy, node, statement, packageName, functionName, functionArgs)
	task := args.Get(0).(*aerov8.ExecuteTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryExecute(policy *aerov8.QueryPolicy, writePolicy *aerov8.WritePolicy, statement *aerov8.Statement, ops ...*aerov8.Operation) (*aerov8.ExecuteTask, aerov8.Error) {
	args := m.Called(policy, writePolicy, statement, ops)
	task := args.Get(0).(*aerov8.ExecuteTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

// Admin
func (m *MockAerospikeClient) CreateUser(policy *aerov8.AdminPolicy, user, password string, roles []string) aerov8.Error {
	args := m.Called(policy, user, password, roles)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) DropUser(policy *aerov8.AdminPolicy, user string) aerov8.Error {
	args := m.Called(policy, user)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) ChangePassword(policy *aerov8.AdminPolicy, user, password string) aerov8.Error {
	args := m.Called(policy, user, password)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) GrantRoles(policy *aerov8.AdminPolicy, user string, roles []string) aerov8.Error {
	args := m.Called(policy, user, roles)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) RevokeRoles(policy *aerov8.AdminPolicy, user string, roles []string) aerov8.Error {
	args := m.Called(policy, user, roles)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryUser(policy *aerov8.AdminPolicy, user string) (*aerov8.UserRoles, aerov8.Error) {
	args := m.Called(policy, user)
	userRoles := args.Get(0).(*aerov8.UserRoles)
	err := args.Get(1)
	if err == nil {
		return userRoles, nil
	}
	return userRoles, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryUsers(policy *aerov8.AdminPolicy) ([]*aerov8.UserRoles, aerov8.Error) {
	args := m.Called(policy)
	users := args.Get(0).([]*aerov8.UserRoles)
	err := args.Get(1)
	if err == nil {
		return users, nil
	}
	return users, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryRole(policy *aerov8.AdminPolicy, role string) (*aerov8.Role, aerov8.Error) {
	args := m.Called(policy, role)
	roleInfo := args.Get(0).(*aerov8.Role)
	err := args.Get(1)
	if err == nil {
		return roleInfo, nil
	}
	return roleInfo, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) QueryRoles(policy *aerov8.AdminPolicy) ([]*aerov8.Role, aerov8.Error) {
	args := m.Called(policy)
	roles := args.Get(0).([]*aerov8.Role)
	err := args.Get(1)
	if err == nil {
		return roles, nil
	}
	return roles, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) CreateRole(policy *aerov8.AdminPolicy, roleName string, privileges []aerov8.Privilege, whitelist []string, readQuota, writeQuota uint32) aerov8.Error {
	args := m.Called(policy, roleName, privileges, whitelist, readQuota, writeQuota)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) DropRole(policy *aerov8.AdminPolicy, roleName string) aerov8.Error {
	args := m.Called(policy, roleName)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) GrantPrivileges(policy *aerov8.AdminPolicy, roleName string, privileges []aerov8.Privilege) aerov8.Error {
	args := m.Called(policy, roleName, privileges)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) RevokePrivileges(policy *aerov8.AdminPolicy, roleName string, privileges []aerov8.Privilege) aerov8.Error {
	args := m.Called(policy, roleName, privileges)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) SetWhitelist(policy *aerov8.AdminPolicy, roleName string, whitelist []string) aerov8.Error {
	args := m.Called(policy, roleName, whitelist)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) SetQuotas(policy *aerov8.AdminPolicy, roleName string, readQuota, writeQuota uint32) aerov8.Error {
	args := m.Called(policy, roleName, readQuota, writeQuota)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

// Indexing
func (m *MockAerospikeClient) CreateIndex(policy *aerov8.WritePolicy, namespace, setName, indexName, binName string, indexType aerov8.IndexType) (*aerov8.IndexTask, aerov8.Error) {
	args := m.Called(policy, namespace, setName, indexName, binName, indexType)
	task := args.Get(0).(*aerov8.IndexTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) CreateComplexIndex(policy *aerov8.WritePolicy, namespace, setName, indexName, binName string, indexType aerov8.IndexType, indexCollectionType aerov8.IndexCollectionType, ctx ...*aerov8.CDTContext) (*aerov8.IndexTask, aerov8.Error) {
	args := m.Called(policy, namespace, setName, indexName, binName, indexType, indexCollectionType, ctx)
	task := args.Get(0).(*aerov8.IndexTask)
	err := args.Get(1)
	if err == nil {
		return task, nil
	}
	return task, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) DropIndex(policy *aerov8.WritePolicy, namespace, setName, indexName string) aerov8.Error {
	args := m.Called(policy, namespace, setName, indexName)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) Stats() (map[string]any, aerov8.Error) {
	args := m.Called()
	stats := args.Get(0).(map[string]any)
	err := args.Get(1)
	if err == nil {
		return stats, nil
	}
	return stats, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) WarmUp(count int) (int, aerov8.Error) {
	args := m.Called(count)
	val := args.Int(0)
	err := args.Get(1)
	if err == nil {
		return val, nil
	}
	return val, err.(*aerov8.AerospikeError)
}

func (m *MockAerospikeClient) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAerospikeClient) Commit(txn *aerov8.Txn) (aerov8.CommitStatus, aerov8.Error) {
	args := m.Called(txn)
	return args.Get(0).(aerov8.CommitStatus), args.Get(1).(aerov8.Error)
}

func (m *MockAerospikeClient) Abort(txn *aerov8.Txn) (aerov8.AbortStatus, aerov8.Error) {
	args := m.Called(txn)
	return args.Get(0).(aerov8.AbortStatus), args.Get(1).(aerov8.Error)
}

func (m *MockAerospikeClient) SetXDRFilter(policy *aerov8.InfoPolicy, datacenter, namespace string, filter *aerov8.Expression) aerov8.Error {
	args := m.Called(policy, datacenter, namespace, filter)
	return args.Get(0).(aerov8.Error)
}
