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
  # Default can be set in controller config, broker's "num.partitions" config is not used
  # Required if no default is specified in controller config
  partitions: 10
  
  # The number of replicas for each of the topic's partitions
  # Must be >= 1
  # Default can be set in controller config, broker's "default.replication.factor" config is not used
  # Required if no default is specified in controller config
  replicationFactor: 3
  
  # (Optional) The configs for the topic, for all available configs see: https://kafka.apache.org/documentation/#topicconfigs
  # Both keys and values must be strings
  # Defaults for each config can be set in controller config
  configs:
    retention.ms: "86400000" # 1 day
    retention.bytes: "1073741824" # 1 GiB
    segment.bytes: "107374182" # 100 MiB
```

KafkaTopics in Kubernetes are namespaced, allowing KafkaTopics to have the same `metadata.name`.
To accommodate this, the controller is configurable with a `nameTemplate` that generates the topic name based on the 
KafkaTopic namespace/name.  The helm chart defaults the `nameTemplate` to `{{ .Namespace }}.{{ .Name }}`, resulting in
topic names such as `my-ns.my-topic` in the example above.

The KafkaTopic name must conform to [RFC 1035](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names).

For Kafka topic naming structures, see Kafka's documentation on [Multitenancy Topic Naming](https://kafka.apache.org/documentation/#multitenancy-topic-naming).
