## Security

Ensure you are familiar with [Kafka Security](https://kafka.apache.org/documentation/#security).

To secure your Kafka Cluster, ensure that:
1. `mtls` is enabled on each `ClusterKafkaCluster` defined
2. Kafka brokers have mTLS enabled for client authentication with proper credentials
3. Containers running on Kubernetes are using mTLS for their Kafka clients with proper credentials
4. Ensure your kafka cluster has `allow.everyone.if.no.acl.found` set to `false`  
5. Configure ACLs to provide desired access to the proper Principals

### Configuring Kafka Brokers for mTLS

Ksflow currently only supports mTLS (what Kafka calls SSL). To configure the Kafka brokers, review the documentation
for the Kafka provider you are using:
* If running Kafka yourself, see https://kafka.apache.org/documentation/#security_ssl
* If using AWS MSK, see [MSK Authentication](https://docs.aws.amazon.com/msk/latest/developerguide/msk-authentication.html).
* If another provider, see their documentation

### Configuring Kubernetes for mTLS

All pods, including the ksflow controller will need credentials to connect to Kafka. There are many ways to
accomplish this, a few are:

* (recommended) [Istio's SPIRE integration](https://istio.io/latest/docs/ops/integrations/spire/) with [spire-controller-manager](https://github.com/spiffe/spire-controller-manager)
* [Otterize](https://github.com/otterize/spire-integration-operator) with SPIRE
* [cert-manager](https://github.com/cert-manager/cert-manager) and [reloader](https://github.com/stakater/Reloader)

Ultimately configuring mTLS for pods in Kubernetes is not specific to pods talking to Kafka, so the solution
may vary based on how your Kubernetes cluster is structured. Just ensure that the Principals defined in the ACLs
match up with the certificates used by the pods.

### Configure ACLs

After verifying `allow.everyone.if.no.acl.found` is set to `false`, the next step is to add the ksflow controller
as a [non-super user ACL Administrator](https://docs.confluent.io/platform/current/kafka/authorization.html#creating-non-super-user-acl-administrators)
as the controller pod is responsible for managing the ACLs.

The ksflow controller watches for `KafkaACL`s and for each, does the following:
1. Verifies the KafkaACL is allowed (currently according to hard-coded rules)
2. If valid, applies the ACL to the Kafka cluster
3. Updates the status of the KafkaACL object to reflect success or failure

Currently, namespaced resources are allowed to define KafkaACLs against topics under certain topic prefixes.
Specifically, a prefix can be specified in the ClusterKafkaCluster spec which will restrict the topics.

For example, let's say there is a ClusterKafkaCluster specified with prefix `com.example`, and a
`KafkaTopic` named `my-topic` is defined in namespace `my-namespace` referencing that ClusterKafkaCluster. In this case,
the topic that will be created (if successful) is `com.example.my-namespace.my-topic`. The steps are:
1. The ksflow controller will create a Kubernetes Job in `my-namespace` which is responsible for creating the topic. The Job object is owned by the KafkaTopic.
2. The ksflow controller will watch the Job and update the status of the KafkaTopic accordingly to reflect success or failure.

This approach allows the pod's permissions to manage the topics, failing if ACLs are insufficient.

Currently, the ServiceAccount used for the Job's pods is hard-coded to `default`, so ensure there is a KafkaACL defined
to give it permissions to create topics. In the future the serviceAccount should be made configurable in the KafkaTopic,
along with other config (i.e. resources, nodeSelector, etc).
