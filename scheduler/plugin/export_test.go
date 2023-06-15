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

package wasm

import (
	wazeroapi "github.com/tetratelabs/wazero/api"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type WasmPlugin struct{ *wasmPlugin }

func NewTestWasmPlugin(p framework.Plugin) *WasmPlugin {
	return &WasmPlugin{wasmPlugin: p.(*wasmPlugin)} // panic on test bug
}

func (w *WasmPlugin) SetGlobals(globals map[string]int32) {
	if err := w.pool.doWithSchedulingGuest(ctx, 0, func(g *guest) {
		// Use test conventions to set a global used to test value range.
		for n, v := range globals {
			g.guest.ExportedGlobal(n + "_global").(wazeroapi.MutableGlobal).Set(uint64(v))
		}
	}); err != nil {
		panic(err)
	}
}

func (w *WasmPlugin) ClearGuestModule() {
	w.guestModule = nil
}

func (w *WasmPlugin) GetSchedulingCycleID() uint32 {
	return w.pool.schedulingCycleID
}

func (w *WasmPlugin) GetBindingCycles() map[uint32]*guest {
	return w.pool.binding
}

func (w *WasmPlugin) GetFreePool() []*guest {
	return w.pool.free
}
