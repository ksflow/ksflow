//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 The Ksflow Authors.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaConnectionConfig) DeepCopyInto(out *KafkaConnectionConfig) {
	*out = *in
	if in.BootstrapServers != nil {
		in, out := &in.BootstrapServers, &out.BootstrapServers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaConnectionConfig.
func (in *KafkaConnectionConfig) DeepCopy() *KafkaConnectionConfig {
	if in == nil {
		return nil
	}
	out := new(KafkaConnectionConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopic) DeepCopyInto(out *KafkaTopic) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopic.
func (in *KafkaTopic) DeepCopy() *KafkaTopic {
	if in == nil {
		return nil
	}
	out := new(KafkaTopic)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaTopic) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicConfig) DeepCopyInto(out *KafkaTopicConfig) {
	*out = *in
	in.KafkaConnectionConfig.DeepCopyInto(&out.KafkaConnectionConfig)
	in.KafkaTopicDefaultsConfig.DeepCopyInto(&out.KafkaTopicDefaultsConfig)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicConfig.
func (in *KafkaTopicConfig) DeepCopy() *KafkaTopicConfig {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicInClusterConfiguration) DeepCopyInto(out *KafkaTopicInClusterConfiguration) {
	*out = *in
	if in.Partitions != nil {
		in, out := &in.Partitions, &out.Partitions
		*out = new(int32)
		**out = **in
	}
	if in.ReplicationFactor != nil {
		in, out := &in.ReplicationFactor, &out.ReplicationFactor
		*out = new(int16)
		**out = **in
	}
	if in.Configs != nil {
		in, out := &in.Configs, &out.Configs
		*out = make(map[string]*string, len(*in))
		for key, val := range *in {
			var outVal *string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(string)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicInClusterConfiguration.
func (in *KafkaTopicInClusterConfiguration) DeepCopy() *KafkaTopicInClusterConfiguration {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicInClusterConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicList) DeepCopyInto(out *KafkaTopicList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KafkaTopic, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicList.
func (in *KafkaTopicList) DeepCopy() *KafkaTopicList {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaTopicList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicSpec) DeepCopyInto(out *KafkaTopicSpec) {
	*out = *in
	in.KafkaTopicInClusterConfiguration.DeepCopyInto(&out.KafkaTopicInClusterConfiguration)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicSpec.
func (in *KafkaTopicSpec) DeepCopy() *KafkaTopicSpec {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicStatus) DeepCopyInto(out *KafkaTopicStatus) {
	*out = *in
	in.KafkaTopicInClusterConfiguration.DeepCopyInto(&out.KafkaTopicInClusterConfiguration)
	in.LastUpdated.DeepCopyInto(&out.LastUpdated)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicStatus.
func (in *KafkaTopicStatus) DeepCopy() *KafkaTopicStatus {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaUser) DeepCopyInto(out *KafkaUser) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaUser.
func (in *KafkaUser) DeepCopy() *KafkaUser {
	if in == nil {
		return nil
	}
	out := new(KafkaUser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaUser) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaUserConfig) DeepCopyInto(out *KafkaUserConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaUserConfig.
func (in *KafkaUserConfig) DeepCopy() *KafkaUserConfig {
	if in == nil {
		return nil
	}
	out := new(KafkaUserConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaUserList) DeepCopyInto(out *KafkaUserList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KafkaUser, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaUserList.
func (in *KafkaUserList) DeepCopy() *KafkaUserList {
	if in == nil {
		return nil
	}
	out := new(KafkaUserList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaUserList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaUserSpec) DeepCopyInto(out *KafkaUserSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaUserSpec.
func (in *KafkaUserSpec) DeepCopy() *KafkaUserSpec {
	if in == nil {
		return nil
	}
	out := new(KafkaUserSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaUserStatus) DeepCopyInto(out *KafkaUserStatus) {
	*out = *in
	in.LastUpdated.DeepCopyInto(&out.LastUpdated)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaUserStatus.
func (in *KafkaUserStatus) DeepCopy() *KafkaUserStatus {
	if in == nil {
		return nil
	}
	out := new(KafkaUserStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KsflowConfig) DeepCopyInto(out *KsflowConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ControllerManagerConfigurationSpec.DeepCopyInto(&out.ControllerManagerConfigurationSpec)
	in.KafkaTopicConfig.DeepCopyInto(&out.KafkaTopicConfig)
	out.KafkaUserConfig = in.KafkaUserConfig
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KsflowConfig.
func (in *KsflowConfig) DeepCopy() *KsflowConfig {
	if in == nil {
		return nil
	}
	out := new(KsflowConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KsflowConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
