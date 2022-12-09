![Ksflow](images/ksflow-logo-3800x670-transparent.png)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Ksflow is a tool to simplify processing Kafka topics using Kubernetes and Istio.

Concerns such as security, configuration and scaling are standardized for developer's pods in a language-agnostic way.

## Features
#### Simple
- Support existing Kafka client libraries
#### Secure
- Kafka ACLs
- mTLS handled by [Istio](https://github.com/istio/istio)
#### Cheap
- Single container written in Go
#### Scale
- Autoscaling managed by [KEDA](https://github.com/kedacore/keda)
- Multi-cluster supported by leveraging topic prefixes
#### Developer Friendly
- Agnostic to programming language, kafka-client, and cloud-provider/on-prem
- Kubernetes-native (CRDs)
- Schema registry support

## Documentation
- [CRDs](./docs/crds.md)
- [Dependencies](./docs/dependencies.md)
- Install with the [ksflow helm chart](https://github.com/ksflow/ksflow-helm/charts/ksflow)
