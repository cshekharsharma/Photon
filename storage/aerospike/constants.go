package aerospike

// minimum number of seed nodes/hosts required to be configured
// in the aerospike cluster config. Ideal value is more than 1
// node to avoid single point of failure.
const MinRequiredSeedNode int = 2

// Default connection timeout in seconds
const DefaultConnectionTimeout uint32 = 20

// Default TTL time for each record in aerspike.
// This value is further overridden by provided
// write policy that is sent to Put() and PutBin()
const DefaultRecordTTL uint32 = 1800
