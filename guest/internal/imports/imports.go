//go:build tinygo.wasm

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

package imports

import "sigs.k8s.io/kube-scheduler-wasm-extension/guest/internal/mem"

//go:wasmimport k8s.io/api node
func k8sApiNode(ptr uint32, limit mem.BufLimit) (len uint32)

//go:wasmimport k8s.io/api nodeName
func k8sApiNodeName(ptr uint32, limit mem.BufLimit) (len uint32)

//go:wasmimport k8s.io/api pod
func k8sApiPod(ptr uint32, limit mem.BufLimit) (len uint32)

//go:wasmimport k8s.io/scheduler result.status_reason
func k8sSchedulerResultStatusReason(ptr, size uint32)
