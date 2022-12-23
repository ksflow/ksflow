## Kafka Topic (kt)

A `KafkaTopic` defines a topic in the Kafka cluster that the ksflow controller is configured to use.
Below is an example KafkaTopic.

```yaml
apiVersion: ksflow.io/v1alpha1
kind: KafkaTopic
metadata:
  name: my-topic
  namespace: my-ns
spec:
  # The number of partitions in the topic
  # Must be >= 1
  # Default can be set in controller config, which helm chart sets to 1 by default (broker's "num.partitions" config is not used)
  partitions: 10
  
  # The number of replicas for each of the topic's partitions
  # Must be >= 1
  # Default can be set in controller config, which helm chart sets to 1 by default (broker's "default.replication.factor" config is not used)
  replicationFactor: 3
  
  # The configs for the topic, for all available configs see: https://kafka.apache.org/documentation/#topicconfigs
  # Both keys and values must be strings
  # Default can be set in controller config, which helm chart sets to {} by default
  configs:
    retention.ms: "86400000" # 1 day
    retention.bytes: "1073741824" # 1 GiB
    segment.bytes: "107374182" # 100 MiB
```

KafkaTopics in Kubernetes are namespaced, meaning that multiple KafkaTopics can have the same name.
To accommodate this, each KafkaTopic will result in an actual topic name of `<namespace>.<name>` in Kafka.
The above example will therefore create a topic named "my-ns.my-topic".

Both the KafkaTopic name and namespace must conform to [RFC 1035](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names).
