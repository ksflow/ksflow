![Ksflow](images/ksflow-logo-3800x670-transparent.png)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Ksflow is a tool to simplify configuration management related to running Kubernetes pods that process Kafka topics.

Application security and configuration are standardized, while still allowing applications to use their preferred languages and Kafka clients.

To accomplish this, Ksflow provides a Kubernetes controller and the introduces the following CRDs:

| CRD           | Short Name | Namespaced | Owns                                                        |
|---------------|------------|------------|-------------------------------------------------------------|
| `KafkaACL`    | `kacl`     | yes        | ACL (kafka)                                                 |
| `KafkaTopic`  | `ktopic`   | yes        | Topic (kafka)                                               |
| `KafkaSchema` | `kschema`  | yes        | Schema (schema-registry)                                    |
| `KafkaUser`   | `kuser`    | yes        | Certificate (kubernetes)                                    |
| `KafkaApp`    | `kapp`     | yes        | KafkaACL, KafkaUser, StatefulSet, ScaledObject (kubernetes) |

The Ksflow controller is configured to point to a single Kafka cluster, operating on topics under a configurable prefix (i.e. `my-cluster.`).
For security Ksflow relies on Kafka's mTLS client authentication and cert-manager [Certificates](https://cert-manager.io/docs/concepts/certificate/)
to manage pod permissions for consuming/producing from/to Kafka topics.


#### KafkaACL
Creates [Kafka ACLs](https://docs.confluent.io/platform/current/kafka/authorization.html) that give one or all
`KafkaUsers` permissions to one or more `KafkaTopics` in the same namespace as the `KafkaACL`.

#### KafkaTopic
Creates a kafka topic. A KafkaTopic named `my-topic` in a kubernetes namespace `my-ns` will result in a topic in Kafka named `my-cluster.my-ns.my-topic`.

You can then use features built into Kubernetes to manage Kafka topics. This can include:
- Kubernetes [Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/#object-count-quota) used to limit the number of topics that can be created in a namespace.
- Kubernetes [Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) used to validate or mutate custom resources (i.e. limit total number of bytes that can be stored for a namespace, by computing total `retention.bytes` for all topic-partitions)

#### KafkaSchema
Creates a schema in a schema registry (i.e. Confluent, Apicurio).

#### KafkaUser
A `KafkaUser` creates a cert-manager `Certificate` resource in the same kubernetes namespace with the same name.
Cert-manager is then responsible for creating the secret, which can be used by pods for authentication to external services (i.e. Kafka, Schema Registry).

Pods may be associated with a `KafkaUser` through an annotation, which causes an admission webhook to inject the necessary
authorization.

#### KafkaApp
Creates an application that consumes and/or produces kafka messages. For most production-ready applications, the intent is
that only the following need to be created: `KafkaSchema`, `KafkaTopic`, `KafkaApp`.

## Goals
- Simple (no kafka proxies, no SDKs, not replacing kafka clients)
- Secure (Kafka ACLs, no developer application dependencies, mTLS, cert-manager)
- Cheap (no required sidecars or PVs for pods, low resource requirements for ksflow pods)
- Autoscaling support (keda)
- Support many Kubernetes clusters using the same Kafka (i.e. topic prefixes, ACLs?)
- Agnostic to programming language, kafka-client, and cloud-provider/on-prem
- Kubernetes-native (CRDs)
- Schema registry support (confluent, apicurio)

## Not Goals
- Simplify application logic for processing streams (i.e. stateful aggregations)
- Abstract Kafka or Kubernetes capabilities to provide other options (i.e. Pulsar, Jetstream, etc.)

## External dependencies
| Tool                                                         | Required | Purpose       |
|--------------------------------------------------------------|----------|---------------|
| [Cert Manager](https://github.com/cert-manager/cert-manager) | yes      | Security      |
| [KEDA](https://github.com/kedacore/keda)                     | yes      | Autoscaling   |
| [Prometheus](https://github.com/prometheus/prometheus)       | no       | Monitoring    |
| [Grafana](https://github.com/grafana/grafana)                | no       | Observability |

## Documentation
- [Quick Start](./docs/quick-start.md)
- [Topic Configuration](https://kafka.apache.org/documentation/#topicconfigs)
- [Consumer Configuration](https://kafka.apache.org/documentation/#consumerconfigs)
- [Producer Configuration](https://kafka.apache.org/documentation/#producerconfigs)

*Note: For non-jvm confluent clients see https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md*