## Kafka Topics

### KafkaTopic

A `KafkaTopic` defines a topic in the Kafka cluster that the ksflow controller is configured to use.
Below is an example KafkaTopic.

```yaml
apiVersion: ksflow.io/v1alpha1
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: my-ns
spec:
  # (Optional) What should happen to the underlying kafka topic if the KafkaTopic is deleted
  # Must be either "Retain" or "Delete"
  # Defaults to "Delete"
  reclaimPolicy: Retain
  
  # (Optional) The number of partitions in the topic
  # Must be >= 1
  # Defaults to "num.partitions", see: https://kafka.apache.org/documentation/#brokerconfigs_num.partitions
  partitions: 1
  
  # (Optional) The number of replicas for each of the topic's partitions
  # Must be >= 1
  # Defaults to "default.replication.factor", see: https://kafka.apache.org/documentation/#brokerconfigs_default.replication.factor
  replicationFactor: 1
  
  # (Optional) The configs for the topic, for all available configs see: https://kafka.apache.org/documentation/#topicconfigs
  # Both keys and values must be strings
  # Defaults to empty, "configs: {}"
  configs:
    retention.ms: "86400000" # 1 day
    retention.bytes: "1073741824" # 1 GiB
    segment.bytes: "107374182" # 100 MiB
```

KafkaTopics in Kubernetes are namespaced, meaning that multiple KafkaTopics can have the same name.
To deal with this, each KafkaTopic will result in an actual topic name of `<namespace>.<name>` in Kafka.
The above example will therefore create a topic named "my-ns.my-topic".

Both the KafkaTopic name and namespace must conform to [RFC 1035](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names).

### ClusterKafkaTopic

A `ClusterKafkaTopic` is the same as a KafkaTopic, the only difference being that it has a cluster scope.
Below is a minimal example ClusterKafkaTopic.

```yaml
apiVersion: ksflow.io/v1alpha1
kind: ClusterKafkaTopic
metadata:
  name: my-cluster-kafka-topic
spec: {}
```

The above ClusterKafkaTopic will result in a topic named "my-cluster-kafka-topic" in Kafka.
As with KafkaTopics, ClusterKafkaTopic names must conform to RFC 1035.
