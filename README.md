![Ksflow](images/ksflow-logo-3800x670-transparent.png)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Ksflow is a tool to simplify processing Kafka topics on Kubernetes.

Several CRDs are introduced, which developers can use to define all the kafka-related configuration their application requires.

Concerns such as security, configuration, and scaling are standardized for developer's pods in a language-agnostic way.

## Motivation

1. There are many situations where Kafka runs external to the Kubernetes cluster where you are processing data.
   * Using a managed Kafka (i.e. [AWS MSK](https://aws.amazon.com/msk/), [Confluent Cloud](https://www.confluent.io/confluent-cloud/))
   * Running your own Kafka outside of Kubernetes (i.e. on VMs)
   * Running Kafka on a separate Kubernetes cluster (i.e. [Strimzi](https://strimzi.io/) or [Confluent for Kubernetes (CFK)](https://docs.confluent.io/operator/current/overview.html))
   
   In such cases, often what is desired is a **lightweight tool focused on developer experience**.
2. The Kafka ecosystem is very friendly to languages that run on the JVM, while **support for other commonly used languages**
(i.e. Python, Go, Rust) is often less of a focus.
3. The tools that support language-agnostic stream-processing applications on Kubernetes, such as [Dapr](https://github.com/dapr/dapr) and [Numaflow](https://github.com/numaproj/numaflow),
provide significant value, but come with various tradeoffs. Specifically, they abstract the event store
(i.e. kafka, redis streams, jetstream, kinesis, pulsar, etc.) by injecting themselves into the traffic and providing an
alternate developer API. For teams **solely focused on Kafka** the provided abstraction often introduces limitations
and/or requires significant refactoring to existing applications without providing sufficient additional value.
4. Strimzi not only manages Kafka clusters, but other resource types (i.e. topics, ACLs, etc). However, often it is not
**cheap** to run Strimzi since "the Topic Operator and User Operator can only watch a single namespace".
Additionally, Strimzi is written in Java, making it more difficult to work with since most kubernetes controllers are
**written in Go** (i.e. [strimzi-client-go](https://github.com/RedHatInsights/strimzi-client-go) is a "work in progress").

## Documentation
- [CRDs](./docs/crds.md)
- [Quick Start](./docs/quick-start.md)
- [Install](./docs/install.md)
- [Security](./docs/security.md)
- [Topic Names](./docs/topic-names.md)
