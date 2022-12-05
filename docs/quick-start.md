# Quick Start

Before getting started you will need the following on your path:
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) with access to a Kubernetes cluster.
- [kafka client](https://kafka.apache.org/downloads) with access to a Kafka cluster.
- [helm](https://github.com/helm/helm)
- [helmfile](https://github.com/helmfile/helmfile)

Below is an example of installing all the necessary components on a new Kubernetes cluster.
Your installation may vary depending on your setup and requirements.

## Install
Create a file named `helmfile.yaml` with the following contents:
```yaml
repositories:
- name: kedacore
  url: https://kedacore.github.io/charts
- name: ksflow
  url: https://ksflow.github.io/ksflow-helm
releases:
- name: keda
  namespace: keda
  createNamespace: true
  chart: kedacore/keda
  version: 2.8.2
- name: ksflow
  namespace: ksflow
  createNamespace: true
  chart: ksflow/ksflow
  version: 0.1.0
```
To install:
```shell
helmfile sync
```
To uninstall:
```shell
helmfile destroy
kubectl delete ns keda strimzi ksflow
```
