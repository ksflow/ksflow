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
  # (Optional) The number of partitions in the topic
  # Must be >= 1
  # Defaults is set in helm chart to 1.
  partitions: 10
  
  # (Optional) The number of replicas for each of the topic's partitions
  # Must be >= 1
  # Defaults is set in helm chart to 1.
  replicationFactor: 3
  
  # (Optional) The configs for the topic, for all available configs see: https://kafka.apache.org/documentation/#topicconfigs
  # Both keys and values must be strings
  # Defaults is set in helm chart to empty, "configs: {}"
  configs:
    retention.ms: "86400000" # 1 day
    retention.bytes: "1073741824" # 1 GiB
    segment.bytes: "107374182" # 100 MiB
```

KafkaTopics in Kubernetes are namespaced, meaning that multiple KafkaTopics can have the same name.
To deal with this, each KafkaTopic will result in an actual topic name of `<namespace>.<name>` in Kafka.
The above example will therefore create a topic named "my-ns.my-topic".

Both the KafkaTopic name and namespace must conform to [RFC 1035](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names).
