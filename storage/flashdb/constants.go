package flashdb

const (
	ShardCount        uint8  = 128  // default number of shards
	maxDeletePerShard uint16 = 1024 // max deletes per shard per janitor run
)
