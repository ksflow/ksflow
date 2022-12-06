
## CRDs

Ksflow introduces the following CRDs:

| CRD           | Short Name | Namespaced | Owns                                                        |
|---------------|------------|------------|-------------------------------------------------------------|
| `KafkaACL`    | `kacl`     | yes        | ACL (kafka)                                                 |
| `KafkaTopic`  | `ktopic`   | yes        | Topic (kafka)                                               |
| `KafkaSchema` | `kschema`  | yes        | Schema (schema-registry)                                    |
| `KafkaUser`   | `kuser`    | yes        | Certificate (kubernetes)                                    |
| `KafkaApp`    | `kapp`     | yes        | KafkaACL, KafkaUser, StatefulSet, ScaledObject (kubernetes) |

The Ksflow controller is configured to point to a single Kafka cluster, operating on topics under a configurable prefix (i.e. `my-cluster.`).
For security, Ksflow relies on Kafka's mTLS client authentication and cert-manager [Certificates](https://cert-manager.io/docs/concepts/certificate/)
to manage pod permissions for consuming/producing from/to Kafka topics.


### KafkaACL
Creates [Kafka ACLs](https://docs.confluent.io/platform/current/kafka/authorization.html) that give one or all
`KafkaUsers` permissions to one or more `KafkaTopics` in the same namespace as the `KafkaACL`.

### KafkaTopic
Creates a kafka topic. A KafkaTopic named `my-topic` in a kubernetes namespace `my-ns` will result in a topic in Kafka named `my-cluster.my-ns.my-topic`.

You can then use features built into Kubernetes to manage Kafka topics. This can include:
- Kubernetes [Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/#object-count-quota) used to limit the number of topics that can be created in a namespace.
- Kubernetes [Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) used to validate or mutate custom resources (i.e. limit total number of bytes that can be stored for a namespace, by computing total `retention.bytes` for all topic-partitions)

### KafkaSchema
Creates a schema in a schema registry.

### KafkaUser
A `KafkaUser` creates a cert-manager `Certificate` resource in the same kubernetes namespace with the same name.
Cert-manager is then responsible for creating the secret, which can be used by pods for authentication to external services (i.e. Kafka, Schema Registry).

Pods may be associated with a `KafkaUser` through an annotation, which causes an admission webhook to inject the necessary
authorization.

### KafkaApp
Creates an application that consumes and/or produces kafka messages. For most production-ready applications, the intent is
that only the following need to be created: `KafkaSchema`, `KafkaTopic`, `KafkaApp`.
