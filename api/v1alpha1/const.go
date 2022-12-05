/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

const (
	KafkaCertificateFilePathCAPem          = "/etc/ksflow/kafka-certs/ca.crt"
	KafkaCertificateFilePathCertificatePem = "/etc/ksflow/kafka-certs/tls.crt"
	KafkaCertificateFilePathPrivateKeyPem  = "/etc/ksflow/kafka-certs/tls.key"
	KafkaCertificateFilePathKeystoreJKS    = "/etc/ksflow/kafka-certs/keystore.jks"
	KafkaCertificateFilePathKeystoreP12    = "/etc/ksflow/kafka-certs/keystore.p12"
	KafkaCertificateFilePathTruststoreJKS  = "/etc/ksflow/kafka-certs/truststore.jks"
	KafkaCertificateFilePathTruststoreP12  = "/etc/ksflow/kafka-certs/truststore.p12"

	KafkaConfigFilePathConsumerProperties = "/etc/ksflow/kafka-configs/consumer.properties"
	KafkaConfigFilePathProducerProperties = "/etc/ksflow/kafka-configs/producer.properties"
)
