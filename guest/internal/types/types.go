/*
   Copyright 2023 The Kubernetes Authors.

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

package types

// Override the default GC with a more performant one.
// Note: this requires tinygo flags: -gc=custom -tags=custommalloc
import (
	_ "github.com/wasilibs/nottinygc"

	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/api"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/internal/imports"
	protoapi "sigs.k8s.io/kube-scheduler-wasm-extension/kubernetes/proto/api"
	meta "sigs.k8s.io/kube-scheduler-wasm-extension/kubernetes/proto/meta"
)

var _ api.NodeInfo = (*NodeInfo)(nil)

type NodeInfo struct {
	n *protoapi.Node
}

func (n *NodeInfo) Node() *protoapi.Node {
	return n.node()
}

func (n *NodeInfo) node() *protoapi.Node {
	if node := n.n; node != nil {
		return node
	}

	b := imports.NodeInfoNode()
	var msg protoapi.Node
	if err := msg.UnmarshalVT(b); err != nil {
		panic(err)
	}
	n.n = &msg
	return n.n
}

var _ api.Pod = (*Pod)(nil)

type Pod struct {
	p *protoapi.Pod
}

func (p *Pod) Metadata() *meta.ObjectMeta {
	return p.pod().Metadata
}

func (p *Pod) Spec() *protoapi.PodSpec {
	return p.pod().Spec
}

func (p *Pod) Status() *protoapi.PodStatus {
	return p.pod().Status
}

var (
	// currentPod is updated when its currentPodID changes.
	currentPod   protoapi.Pod
	currentPodID uint32
)

func updatePod(currentID uint32, bytes []byte) error {
	currentPodID = currentID
	return currentPod.UnmarshalVT(bytes)
}

// Pod lazy updates p if it was nil or different from the last call.
func (p *Pod) pod() *protoapi.Pod {
	if pod := p.p; pod != nil {
		return pod
	}

	if err := imports.ConditionallyUpdatePod(currentPodID, updatePod); err != nil {
		panic(err.Error())
	}
	p.p = &currentPod
	return p.p
}
