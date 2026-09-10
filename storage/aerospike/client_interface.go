package aerospike

import (
	as "github.com/aerospike/aerospike-client-go/v8"
)

// AerospikeClientInterface defines the interface for Aerospike client operations.
// This interface is taken from aerospike V8 client, and may be updated in the future.
type AerospikeClientInterface interface {
	// Basic connection methods
	IsConnected() bool
	Close()

	// Policies
	GetDefaultPolicy() *as.BasePolicy
	GetDefaultBatchPolicy() *as.BatchPolicy
	GetDefaultBatchReadPolicy() *as.BatchReadPolicy
	GetDefaultBatchWritePolicy() *as.BatchWritePolicy
	GetDefaultBatchDeletePolicy() *as.BatchDeletePolicy
	GetDefaultBatchUDFPolicy() *as.BatchUDFPolicy
	GetDefaultWritePolicy() *as.WritePolicy
	GetDefaultScanPolicy() *as.ScanPolicy
	GetDefaultQueryPolicy() *as.QueryPolicy
	GetDefaultAdminPolicy() *as.AdminPolicy
	GetDefaultInfoPolicy() *as.InfoPolicy
	GetDefaultTxnVerifyPolicy() *as.TxnVerifyPolicy
	GetDefaultTxnRollPolicy() *as.TxnRollPolicy

	SetDefaultPolicy(policy *as.BasePolicy)
	SetDefaultBatchPolicy(policy *as.BatchPolicy)
	SetDefaultBatchWritePolicy(policy *as.BatchWritePolicy)
	SetDefaultBatchReadPolicy(policy *as.BatchReadPolicy)
	SetDefaultBatchDeletePolicy(policy *as.BatchDeletePolicy)
	SetDefaultBatchUDFPolicy(policy *as.BatchUDFPolicy)
	SetDefaultWritePolicy(policy *as.WritePolicy)
	SetDefaultScanPolicy(policy *as.ScanPolicy)
	SetDefaultQueryPolicy(policy *as.QueryPolicy)
	SetDefaultAdminPolicy(policy *as.AdminPolicy)
	SetDefaultInfoPolicy(policy *as.InfoPolicy)
	SetDefaultTxnVerifyPolicy(policy *as.TxnVerifyPolicy)
	SetDefaultTxnRollPolicy(policy *as.TxnRollPolicy)

	// Node/Cluster
	GetNodes() []*as.Node
	GetNodeNames() []string
	Cluster() *as.Cluster

	// CRUD
	Put(policy *as.WritePolicy, key *as.Key, binMap as.BinMap) as.Error
	PutBins(policy *as.WritePolicy, key *as.Key, bins ...*as.Bin) as.Error
	Get(policy *as.BasePolicy, key *as.Key, binNames ...string) (*as.Record, as.Error)
	GetHeader(policy *as.BasePolicy, key *as.Key) (*as.Record, as.Error)
	Delete(policy *as.WritePolicy, key *as.Key) (bool, as.Error)
	Touch(policy *as.WritePolicy, key *as.Key) as.Error
	Exists(policy *as.BasePolicy, key *as.Key) (bool, as.Error)

	// Batch
	BatchGet(policy *as.BatchPolicy, keys []*as.Key, binNames ...string) ([]*as.Record, as.Error)
	BatchGetOperate(policy *as.BatchPolicy, keys []*as.Key, ops ...*as.Operation) ([]*as.Record, as.Error)
	BatchGetComplex(policy *as.BatchPolicy, records []*as.BatchRead) as.Error
	BatchGetHeader(policy *as.BatchPolicy, keys []*as.Key) ([]*as.Record, as.Error)
	BatchExists(policy *as.BatchPolicy, keys []*as.Key) ([]bool, as.Error)
	BatchDelete(policy *as.BatchPolicy, deletePolicy *as.BatchDeletePolicy, keys []*as.Key) ([]*as.BatchRecord, as.Error)
	BatchOperate(policy *as.BatchPolicy, records []as.BatchRecordIfc) as.Error
	BatchExecute(policy *as.BatchPolicy, udfPolicy *as.BatchUDFPolicy, keys []*as.Key, packageName, functionName string, args ...as.Value) ([]*as.BatchRecord, as.Error)

	// Operate
	Operate(policy *as.WritePolicy, key *as.Key, operations ...*as.Operation) (*as.Record, as.Error)

	// Scan
	ScanAll(policy *as.ScanPolicy, namespace, setName string, binNames ...string) (*as.Recordset, as.Error)
	ScanPartitions(policy *as.ScanPolicy, partitionFilter *as.PartitionFilter, namespace, setName string, binNames ...string) (*as.Recordset, as.Error)
	ScanNode(policy *as.ScanPolicy, node *as.Node, namespace, setName string, binNames ...string) (*as.Recordset, as.Error)

	// Query
	Query(policy *as.QueryPolicy, statement *as.Statement) (*as.Recordset, as.Error)
	QueryPartitions(policy *as.QueryPolicy, statement *as.Statement, partitionFilter *as.PartitionFilter) (*as.Recordset, as.Error)
	QueryNode(policy *as.QueryPolicy, node *as.Node, statement *as.Statement) (*as.Recordset, as.Error)

	// UDF
	RegisterUDF(policy *as.WritePolicy, udfBody []byte, serverPath string, language as.Language) (*as.RegisterTask, as.Error)
	RegisterUDFFromFile(policy *as.WritePolicy, clientPath, serverPath string, language as.Language) (*as.RegisterTask, as.Error)
	RemoveUDF(policy *as.WritePolicy, udfName string) (*as.RemoveTask, as.Error)
	ListUDF(policy *as.BasePolicy) ([]*as.UDF, as.Error)
	Execute(policy *as.WritePolicy, key *as.Key, packageName, functionName string, args ...as.Value) (any, as.Error)

	// Query Execute UDF
	ExecuteUDF(policy *as.QueryPolicy, statement *as.Statement, packageName, functionName string, functionArgs ...as.Value) (*as.ExecuteTask, as.Error)
	ExecuteUDFNode(policy *as.QueryPolicy, node *as.Node, statement *as.Statement, packageName, functionName string, functionArgs ...as.Value) (*as.ExecuteTask, as.Error)
	QueryExecute(policy *as.QueryPolicy, writePolicy *as.WritePolicy, statement *as.Statement, ops ...*as.Operation) (*as.ExecuteTask, as.Error)

	// Admin (User/Role Management)
	CreateUser(policy *as.AdminPolicy, user, password string, roles []string) as.Error
	DropUser(policy *as.AdminPolicy, user string) as.Error
	ChangePassword(policy *as.AdminPolicy, user, password string) as.Error
	GrantRoles(policy *as.AdminPolicy, user string, roles []string) as.Error
	RevokeRoles(policy *as.AdminPolicy, user string, roles []string) as.Error
	QueryUser(policy *as.AdminPolicy, user string) (*as.UserRoles, as.Error)
	QueryUsers(policy *as.AdminPolicy) ([]*as.UserRoles, as.Error)
	QueryRole(policy *as.AdminPolicy, role string) (*as.Role, as.Error)
	QueryRoles(policy *as.AdminPolicy) ([]*as.Role, as.Error)
	CreateRole(policy *as.AdminPolicy, roleName string, privileges []as.Privilege, whitelist []string, readQuota, writeQuota uint32) as.Error
	DropRole(policy *as.AdminPolicy, roleName string) as.Error
	GrantPrivileges(policy *as.AdminPolicy, roleName string, privileges []as.Privilege) as.Error
	RevokePrivileges(policy *as.AdminPolicy, roleName string, privileges []as.Privilege) as.Error
	SetWhitelist(policy *as.AdminPolicy, roleName string, whitelist []string) as.Error
	SetQuotas(policy *as.AdminPolicy, roleName string, readQuota, writeQuota uint32) as.Error

	// Indexing
	CreateIndex(policy *as.WritePolicy, namespace, setName, indexName, binName string, indexType as.IndexType) (*as.IndexTask, as.Error)
	CreateComplexIndex(policy *as.WritePolicy, namespace, setName, indexName, binName string, indexType as.IndexType, indexCollectionType as.IndexCollectionType, ctx ...*as.CDTContext) (*as.IndexTask, as.Error)
	DropIndex(policy *as.WritePolicy, namespace, setName, indexName string) as.Error

	// Other Utilities
	Stats() (map[string]any, as.Error)
	WarmUp(count int) (int, as.Error)
	String() string

	// Multi-Record Transaction (New)
	Commit(txn *as.Txn) (as.CommitStatus, as.Error)
	Abort(txn *as.Txn) (as.AbortStatus, as.Error)

	// Set XDR Filter
	SetXDRFilter(policy *as.InfoPolicy, datacenter, namespace string, filter *as.Expression) as.Error
}
