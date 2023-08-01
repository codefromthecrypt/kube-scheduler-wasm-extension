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

package main

import (
	"os"

	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/api"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/api/proto"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/enqueue"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/filter"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/prefilter"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/score"
)

type extensionPoints interface {
	api.EnqueueExtensions
	api.PreFilterPlugin
	api.FilterPlugin
	api.ScorePlugin
}

func main() {
	// Multiple tests are here to reduce re-compilation time and size checked
	// into git.
	var plugin extensionPoints
	if len(os.Args) == 2 && os.Args[1] == "params" {
		plugin = paramsPlugin{}
	} else {
		plugin = noopPlugin{}
	}

	enqueue.SetPlugin(plugin)
	prefilter.SetPlugin(plugin)
	filter.SetPlugin(plugin)
	score.SetPlugin(plugin)
}

// noopPlugin doesn't do anything, and this style isn't recommended. This shows the
// impact two things:
//
//   - implementing multiple interfaces
//   - overhead of constructing function parameters
type noopPlugin struct{}

func (noopPlugin) EventsToRegister() (clusterEvents []api.ClusterEvent) { return }

func (noopPlugin) PreFilter(api.CycleState, proto.Pod) (nodeNames []string, status *api.Status) {
	return
}

func (noopPlugin) Filter(api.CycleState, proto.Pod, api.NodeInfo) (status *api.Status) { return }

func (noopPlugin) Score(api.CycleState, proto.Pod, string) (score int32, status *api.Status) { return }

// paramsPlugin doesn't do anything, except evaluate each parameter. This shows
// if protobuf unmarshal caching works (for the pod), and also baseline
// performance of reading each parameter.
type paramsPlugin struct{}

func (paramsPlugin) EventsToRegister() (clusterEvents []api.ClusterEvent) {
	return
}

func (paramsPlugin) PreFilter(state api.CycleState, pod proto.Pod) (nodeNames []string, status *api.Status) {
	_, _ = state.Read("ok")
	_ = pod.Spec()
	return
}

func (paramsPlugin) Filter(state api.CycleState, pod proto.Pod, nodeInfo api.NodeInfo) (status *api.Status) {
	_, _ = state.Read("ok")
	_ = pod.Spec()
	_ = nodeInfo.Node().Spec() // trigger lazy loading
	return
}

func (paramsPlugin) PreScore(state api.CycleState, pod proto.Pod, nodeList proto.NodeList) *api.Status {
	_, _ = state.Read("ok")
	_ = pod.Spec()
	_ = nodeList.Items()
	return nil
}

func (paramsPlugin) Score(state api.CycleState, pod proto.Pod, nodeName string) (score int32, status *api.Status) {
	_, _ = state.Read("ok")
	_ = pod.Spec()
	_ = nodeName
	return
}
