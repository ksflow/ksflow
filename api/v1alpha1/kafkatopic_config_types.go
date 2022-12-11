/*
Copyright 2022 The Ksflow Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

// KafkaTopicConfigs defines the Kafka configs for Kafka Topics
type KafkaTopicConfigs struct {
	// CleanupPolicy designates the retention policy to use on log segments.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_cleanup.policy
	// +optional
	CleanupPolicy *[]string `json:"cleanup.policy,omitempty"`

	// CompressionType is the final compression type for a given topic.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_compression.type
	// +optional
	CompressionType *string `json:"compression.type,omitempty"`

	// DeleteRetentionMs is the amount of time to retain delete tombstone markers for log compacted topics.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_delete.retention.ms
	// +optional
	DeleteRetentionMs *uint64 `json:"delete.retention.ms,omitempty"`

	// FileDeleteDelayMs is the amount of time to wait before deleting a file from the filesystem
	// ref: https://kafka.apache.org/documentation/#topicconfigs_file.delete.delay.ms
	// +optional
	FileDeleteDelayMs *uint64 `json:"file.delete.delay.ms,omitempty"`

	// FlushMessages allows specifying an interval at which we will force an fsync of data written to the log.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_flush.messages
	// +optional
	FlushMessages *uint64 `json:"flush.messages,omitempty"`

	// FlushMs allows specifying a time interval at which we will force an fsync of data written to the log.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_flush.ms
	// +optional
	FlushMs *uint64 `json:"flush.ms,omitempty"`

	// FollowerReplicationThrottledReplicas is a list of replicas for which log replication should be throttled on the follower side.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_follower.replication.throttled.replicas
	// +optional
	FollowerReplicationThrottledReplicas *[]string `json:"follower.replication.throttled.replicas,omitempty"`

	// IndexIntervalBytes controls how frequently Kafka adds an index entry to its offset index.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_index.interval.bytes
	// +optional
	IndexIntervalBytes *uint32 `json:"index.interval.bytes,omitempty"`

	// LeaderReplicationThrottledReplicas is a list of replicas for which log replication should be throttled on the leader side.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_leader.replication.throttled.replicas
	// +optional
	LeaderReplicationThrottledReplicas *[]string `json:"leader.replication.throttled.replicas,omitempty"`

	// MaxCompactionLagMs is the maximum time a message will remain ineligible for compaction in the log.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_max.compaction.lag.ms
	// +optional
	MaxCompactionLagMs *uint64 `json:"max.compaction.lag.ms,omitempty"`

	// MaxMessageBytes is the largest record batch size allowed by Kafka (after compression if compression is enabled).
	// ref: https://kafka.apache.org/documentation/#topicconfigs_max.message.bytes
	// +optional
	MaxMessageBytes *uint32 `json:"max.message.bytes,omitempty"`

	// MessageFormatVersion [DEPRECATED] is used to specify the message format version the broker will use to append messages to the logs.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_message.format.version
	// +optional
	MessageFormatVersion *string `json:"message.format.version,omitempty"`

	// MessageTimestampDifferenceMaxMs is the maximum difference allowed between the timestamp when a broker receives a message and the timestamp specified in the message.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_message.timestamp.difference.max.ms
	// +optional
	MessageTimestampDifferenceMaxMs *uint64 `json:"message.timestamp.difference.max.ms,omitempty"`

	// MessageTimestampType defines whether the timestamp in the message is message create time or log append time.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_message.timestamp.type
	// +optional
	MessageTimestampType *string `json:"message.timestamp.type,omitempty"`

	// MinCleanableDirtyRatio controls how frequently the log compactor will attempt to clean the log (assuming log compaction is enabled).
	// ref: https://kafka.apache.org/documentation/#topicconfigs_min.cleanable.dirty.ratio
	// +optional
	MinCleanableDirtyRatio *float64 `json:"min.cleanable.dirty.ratio,omitempty"`

	// MinCompactionLagMs is the minimum time a message will remain uncompacted in the log.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_min.compaction.lag.ms
	// +optional
	MinCompactionLagMs *uint64 `json:"min.compaction.lag.ms,omitempty"`

	// MinInsyncReplicas when a producer sets acks to "all" (or "-1"), this configuration specifies the minimum number of replicas that must acknowledge a write for the write to be considered successful.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_min.insync.replicas
	// +optional
	MinInsyncReplicas *uint32 `json:"min.insync.replicas,omitempty"`

	// Preallocate indicates if we should preallocate the file on disk when creating a new log segment.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_preallocate
	// +optional
	Preallocate *bool `json:"preallocate,omitempty"`

	// RetentionBytes controls the maximum size a partition (which consists of log segments) can grow to before we will discard old log segments to free up space if we are using the "delete" retention policy.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_retention.bytes
	// +optional
	RetentionBytes *uint64 `json:"retention.bytes,omitempty"`

	// RetentionMs controls the maximum time we will retain a log before we will discard old log segments to free up space if we are using the "delete" retention policy.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_retention.ms
	// +optional
	RetentionMs *uint64 `json:"retention.ms,omitempty"`

	// SegmentBytes controls the segment file size for the log.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_segment.bytes
	// +optional
	SegmentBytes *uint32 `json:"segment.bytes,omitempty"`

	// SegmentIndexBytes controls the size of the index that maps offsets to file positions.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_segment.index.bytes
	// +optional
	SegmentIndexBytes *uint32 `json:"segment.index.bytes,omitempty"`

	// SegmentJitterMs is the maximum random jitter subtracted from the scheduled segment roll time to avoid thundering herds of segment rolling.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_segment.jitter.ms
	// +optional
	SegmentJitterMs *uint64 `json:"segment.jitter.ms,omitempty"`

	// SegmentMs controls the period of time after which Kafka will force the log to roll even if the segment file isn't full to ensure that retention can delete or compact old data.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_segment.ms
	// +optional
	SegmentMs *uint64 `json:"segment.ms,omitempty"`

	// UncleanLeaderElectionEnable indicates whether to enable replicas not in the ISR set to be elected as leader as a last resort, even though doing so may result in data loss.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_unclean.leader.election.enable
	// +optional
	UncleanLeaderElectionEnable *bool `json:"unclean.leader.election.enable,omitempty"`

	// MessageDownconversionEnable controls whether down-conversion of message formats is enabled to satisfy consume requests.
	// ref: https://kafka.apache.org/documentation/#topicconfigs_message.downconversion.enable
	// +optional
	MessageDownconversionEnable *bool `json:"message.downconversion.enable,omitempty"`
}
