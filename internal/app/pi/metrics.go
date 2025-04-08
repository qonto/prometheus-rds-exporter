package pi

import (
	"github.com/aws/aws-sdk-go-v2/service/pi/types"
)

var DbMetricNames = []string{
	"db.Cache.blks_hit.avg",
	"db.Cache.buffers_alloc.avg",
	"db.Checkpoint.buffers_checkpoint.avg",
	"db.Checkpoint.checkpoint_sync_time.avg",
	"db.Checkpoint.checkpoint_write_time.avg",
	"db.Checkpoint.checkpoints_req.avg",
	"db.Checkpoint.checkpoints_timed.avg",
	"db.Checkpoint.maxwritten_clean.avg",
	"db.Concurrency.deadlocks.avg",
	"db.IO.blk_read_time.avg",
	"db.IO.blks_read.avg",
	"db.IO.buffers_backend.avg",
	"db.IO.buffers_backend_fsync.avg",
	"db.IO.buffers_clean.avg",
	"db.SQL.tup_deleted.avg",
	"db.SQL.tup_fetched.avg",
	"db.SQL.tup_inserted.avg",
	"db.SQL.tup_returned.avg",
	"db.SQL.tup_updated.avg",
	"db.Temp.temp_bytes.avg",
	"db.Temp.temp_files.avg",
	"db.Transactions.blocked_transactions.avg",
	"db.Transactions.max_used_xact_ids.avg",
	"db.Transactions.xact_commit.avg",
	"db.Transactions.xact_rollback.avg",
	"db.Transactions.oldest_inactive_logical_replication_slot_xid_age.avg",
	"db.Transactions.oldest_active_logical_replication_slot_xid_age.avg",
	"db.Transactions.oldest_prepared_transaction_xid_age.avg",
	"db.Transactions.oldest_running_transaction_xid_age.avg",
	"db.Transactions.oldest_hot_standby_feedback_xid_age.avg",
	"db.User.numbackends.avg",
	"db.User.max_connections.avg",
	"db.WAL.archived_count.avg",
	"db.WAL.archive_failed_count.avg",
	"db.state.active_count.avg",
	"db.state.idle_count.avg",
	"db.state.idle_in_transaction_count.avg",
	"db.state.idle_in_transaction_aborted_count.avg",
	"db.state.idle_in_transaction_max_time.avg",
	"db.Checkpoint.checkpoint_sync_latency.avg",
	"db.Checkpoint.checkpoint_write_latency.avg",
	"db.Transactions.active_transactions.avg",
}

type PerformanceInsightsMetrics struct {
	DbCacheBlksHit                             float64
	DbCacheBuffersAlloc                        float64
	DbCheckpointBuffersCheckpoint              float64
	DbCheckpointCheckpointSyncTime             float64
	DbCheckpointCheckpointWriteTime            float64
	DbCheckpointCheckpointsReq                 float64
	DbCheckpointCheckpointsTimed               float64
	DbCheckpointMaxwrittenClean                float64
	DbConcurrencyDeadlocks                     float64
	DbIOBlkReadTime                            float64
	DbIOBlksRead                               float64
	DbIOBuffersBackend                         float64
	DbIOBuffersBackendFsync                    float64
	DbIOBuffersClean                           float64
	DbSQLTupDeleted                            float64
	DbSQLTupFetched                            float64
	DbSQLTupInserted                           float64
	DbSQLTupReturned                           float64
	DbSQLTupUpdated                            float64
	DbTempTempBytes                            float64
	DbTempTempFiles                            float64
	DbTransactionsBlockedTransactions          float64
	DbTransactionsMaxUsedXactIds               float64
	DbTransactionsXactCommit                   float64
	DbTransactionsXactRollback                 float64
	DbTransactionsOldestInactiveLogicalSlotXid float64
	DbTransactionsOldestActiveLogicalSlotXid   float64
	DbTransactionsOldestPreparedXid            float64
	DbTransactionsOldestRunningXid             float64
	DbTransactionsOldestHotStandbyXid          float64
	DbUserNumbackends                          float64
	DbUserMaxConnections                       float64
	DbWALArchivedCount                         float64
	DbWALArchiveFailedCount                    float64
	DbStateActiveCount                         float64
	DbStateIdleCount                           float64
	DbStateIdleInTransactionCount              float64
	DbStateIdleInTransactionAbortedCount       float64
	DbStateIdleInTransactionMaxTime            float64
	DbCheckpointCheckpointSyncLatency          float64
	DbCheckpointCheckpointWriteLatency         float64
	DbTransactionsActiveTransactions           float64
}

func getMetricsDataPoint(point types.MetricKeyDataPoints) float64 {

	length := len(point.DataPoints)
	if length < 1 {
		return 0
	}
	if point.DataPoints[length-1].Value == nil {
		return 0
	}

	return *point.DataPoints[length-1].Value
}

func fillMetricsData(piMetrics []types.MetricKeyDataPoints) PerformanceInsightsMetrics {

	var output PerformanceInsightsMetrics

	for _, data := range piMetrics {
		value := getMetricsDataPoint(data)
		switch *data.Key.Metric {
		case "db.Cache.blks_hit.avg":
			output.DbCacheBlksHit = value
		case "db.Cache.buffers_alloc.avg":
			output.DbCacheBuffersAlloc = value
		case "db.Checkpoint.buffers_checkpoint.avg":
			output.DbCheckpointBuffersCheckpoint = value
		case "db.Checkpoint.checkpoint_sync_time.avg":
			output.DbCheckpointCheckpointSyncTime = value
		case "db.Checkpoint.checkpoint_write_time.avg":
			output.DbCheckpointCheckpointWriteTime = value
		case "db.Checkpoint.checkpoints_req.avg":
			output.DbCheckpointCheckpointsReq = value
		case "db.Checkpoint.checkpoints_timed.avg":
			output.DbCheckpointCheckpointsTimed = value
		case "db.Checkpoint.maxwritten_clean.avg":
			output.DbCheckpointMaxwrittenClean = value
		case "db.Concurrency.deadlocks.avg":
			output.DbConcurrencyDeadlocks = value
		case "db.IO.blk_read_time.avg":
			output.DbIOBlkReadTime = value
		case "db.IO.blks_read.avg":
			output.DbIOBlksRead = value
		case "db.IO.buffers_backend.avg":
			output.DbIOBuffersBackend = value
		case "db.IO.buffers_backend_fsync.avg":
			output.DbIOBuffersBackendFsync = value
		case "db.IO.buffers_clean.avg":
			output.DbIOBuffersClean = value
		case "db.SQL.tup_deleted.avg":
			output.DbSQLTupDeleted = value
		case "db.SQL.tup_fetched.avg":
			output.DbSQLTupFetched = value
		case "db.SQL.tup_inserted.avg":
			output.DbSQLTupInserted = value
		case "db.SQL.tup_returned.avg":
			output.DbSQLTupReturned = value
		case "db.SQL.tup_updated.avg":
			output.DbSQLTupUpdated = value
		case "db.Temp.temp_bytes.avg":
			output.DbTempTempBytes = value
		case "db.Temp.temp_files.avg":
			output.DbTempTempFiles = value
		case "db.Transactions.blocked_transactions.avg":
			output.DbTransactionsBlockedTransactions = value
		case "db.Transactions.max_used_xact_ids.avg":
			output.DbTransactionsMaxUsedXactIds = value
		case "db.Transactions.xact_commit.avg":
			output.DbTransactionsXactCommit = value
		case "db.Transactions.xact_rollback.avg":
			output.DbTransactionsXactRollback = value
		case "db.Transactions.oldest_inactive_logical_replication_slot_xid_age.avg":
			output.DbTransactionsOldestInactiveLogicalSlotXid = value
		case "db.Transactions.oldest_active_logical_replication_slot_xid_age.avg":
			output.DbTransactionsOldestActiveLogicalSlotXid = value
		case "db.Transactions.oldest_prepared_transaction_xid_age.avg":
			output.DbTransactionsOldestPreparedXid = value
		case "db.Transactions.oldest_running_transaction_xid_age.avg":
			output.DbTransactionsOldestRunningXid = value
		case "db.Transactions.oldest_hot_standby_feedback_xid_age.avg":
			output.DbTransactionsOldestHotStandbyXid = value
		case "db.User.numbackends.avg":
			output.DbUserNumbackends = value
		case "db.User.max_connections.avg":
			output.DbUserMaxConnections = value
		case "db.WAL.archived_count.avg":
			output.DbWALArchivedCount = value
		case "db.WAL.archive_failed_count.avg":
			output.DbWALArchiveFailedCount = value
		case "db.state.active_count.avg":
			output.DbStateActiveCount = value
		case "db.state.idle_count.avg":
			output.DbStateIdleCount = value
		case "db.state.idle_in_transaction_count.avg":
			output.DbStateIdleInTransactionCount = value
		case "db.state.idle_in_transaction_aborted_count.avg":
			output.DbStateIdleInTransactionAbortedCount = value
		case "db.state.idle_in_transaction_max_time.avg":
			output.DbStateIdleInTransactionMaxTime = value
		case "db.Checkpoint.checkpoint_sync_latency.avg":
			output.DbCheckpointCheckpointSyncLatency = value
		case "db.Checkpoint.checkpoint_write_latency.avg":
			output.DbCheckpointCheckpointWriteLatency = value
		case "db.Transactions.active_transactions.avg":
			output.DbTransactionsActiveTransactions = value
		}
	}
	return output
}
