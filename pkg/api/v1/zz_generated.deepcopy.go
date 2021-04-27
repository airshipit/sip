// +build !ignore_autogenerated

/*


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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BMCOpts) DeepCopyInto(out *BMCOpts) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BMCOpts.
func (in *BMCOpts) DeepCopy() *BMCOpts {
	if in == nil {
		return nil
	}
	out := new(BMCOpts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JumpHostService) DeepCopyInto(out *JumpHostService) {
	*out = *in
	in.SIPClusterService.DeepCopyInto(&out.SIPClusterService)
	if in.BMC != nil {
		in, out := &in.BMC, &out.BMC
		*out = new(BMCOpts)
		**out = **in
	}
	if in.SSHAuthorizedKeys != nil {
		in, out := &in.SSHAuthorizedKeys, &out.SSHAuthorizedKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JumpHostService.
func (in *JumpHostService) DeepCopy() *JumpHostService {
	if in == nil {
		return nil
	}
	out := new(JumpHostService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancerServiceControlPlane) DeepCopyInto(out *LoadBalancerServiceControlPlane) {
	*out = *in
	in.SIPClusterService.DeepCopyInto(&out.SIPClusterService)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancerServiceControlPlane.
func (in *LoadBalancerServiceControlPlane) DeepCopy() *LoadBalancerServiceControlPlane {
	if in == nil {
		return nil
	}
	out := new(LoadBalancerServiceControlPlane)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancerServiceWorker) DeepCopyInto(out *LoadBalancerServiceWorker) {
	*out = *in
	in.SIPClusterService.DeepCopyInto(&out.SIPClusterService)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancerServiceWorker.
func (in *LoadBalancerServiceWorker) DeepCopy() *LoadBalancerServiceWorker {
	if in == nil {
		return nil
	}
	out := new(LoadBalancerServiceWorker)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeCount) DeepCopyInto(out *NodeCount) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeCount.
func (in *NodeCount) DeepCopy() *NodeCount {
	if in == nil {
		return nil
	}
	out := new(NodeCount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeSet) DeepCopyInto(out *NodeSet) {
	*out = *in
	in.LabelSelector.DeepCopyInto(&out.LabelSelector)
	if in.Count != nil {
		in, out := &in.Count, &out.Count
		*out = new(NodeCount)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeSet.
func (in *NodeSet) DeepCopy() *NodeSet {
	if in == nil {
		return nil
	}
	out := new(NodeSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SIPCluster) DeepCopyInto(out *SIPCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SIPCluster.
func (in *SIPCluster) DeepCopy() *SIPCluster {
	if in == nil {
		return nil
	}
	out := new(SIPCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SIPCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SIPClusterList) DeepCopyInto(out *SIPClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SIPCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SIPClusterList.
func (in *SIPClusterList) DeepCopy() *SIPClusterList {
	if in == nil {
		return nil
	}
	out := new(SIPClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SIPClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SIPClusterService) DeepCopyInto(out *SIPClusterService) {
	*out = *in
	if in.NodeLabels != nil {
		in, out := &in.NodeLabels, &out.NodeLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ClusterIP != nil {
		in, out := &in.ClusterIP, &out.ClusterIP
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SIPClusterService.
func (in *SIPClusterService) DeepCopy() *SIPClusterService {
	if in == nil {
		return nil
	}
	out := new(SIPClusterService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SIPClusterServices) DeepCopyInto(out *SIPClusterServices) {
	*out = *in
	if in.LoadBalancerControlPlane != nil {
		in, out := &in.LoadBalancerControlPlane, &out.LoadBalancerControlPlane
		*out = make([]LoadBalancerServiceControlPlane, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LoadBalancerWorker != nil {
		in, out := &in.LoadBalancerWorker, &out.LoadBalancerWorker
		*out = make([]LoadBalancerServiceWorker, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.JumpHost != nil {
		in, out := &in.JumpHost, &out.JumpHost
		*out = make([]JumpHostService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SIPClusterServices.
func (in *SIPClusterServices) DeepCopy() *SIPClusterServices {
	if in == nil {
		return nil
	}
	out := new(SIPClusterServices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SIPClusterSpec) DeepCopyInto(out *SIPClusterSpec) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make(map[BMHRole]NodeSet, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	in.Services.DeepCopyInto(&out.Services)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SIPClusterSpec.
func (in *SIPClusterSpec) DeepCopy() *SIPClusterSpec {
	if in == nil {
		return nil
	}
	out := new(SIPClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SIPClusterStatus) DeepCopyInto(out *SIPClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SIPClusterStatus.
func (in *SIPClusterStatus) DeepCopy() *SIPClusterStatus {
	if in == nil {
		return nil
	}
	out := new(SIPClusterStatus)
	in.DeepCopyInto(out)
	return out
}
