## CRDs

Ksflow introduces the following CRDs:

| CRD                      | Short Name | Namespaced | Owns          |
|--------------------------|------------|------------|---------------|
| `ClusterKafkaCluster`    | `ckc`      | no         |               |
| `KafkaTopic`             | `kt`       | yes        | topic (kafka) |
| `KafkaACL`               | `ka`       | yes        | ACL (kafka)   |
| `KafkaConsumerConfig`    | `kcc`      | yes        |               |
| `KafkaProducerConfig`    | `kpc`      | yes        |               |
| `KafkaAdminClientConfig` | `kacc`     | yes        |               |

### ClusterKafkaCluster
Ksflow does not manage Kafka brokers, only the things that an application might need to configure (i.e. topics, acls).
Because of this, a ClusterKafkaCluster only defines the properties required for an application to connect to an
external Kafka cluster. Notably it specifies the bootstrap servers and if mTLS is enabled.

**Enabling mTLS is strongly recommended. See [security.md](./security.md) for details.**

### KafkaTopic

### KafkaACL

### KafkaConsumerConfig

### KafkaAdminClientConfig
