/*
 * Copyright 2025 - IBM Corporation. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"testing"

	resourceapi "k8s.io/api/resource/v1"
)

// A claim can be satisfied by devices from several drivers. NodePrepareResources
// must prepare only this driver's devices and skip the rest. Device names are
// scoped by driver and pool, so another driver may legitimately use the same
// device name as an Nx device; skipping by driver (not by allocatable presence)
// keeps that foreign device from being silently prepared as an Nx device.
func TestPrepareDevicesSkipsOtherDriverResults(t *testing.T) {
	const nxDevice = "nx-0"
	localResult := resourceapi.DeviceRequestAllocationResult{
		Request: "nx",
		Driver:  DriverName,
		Pool:    "nx-pool",
		Device:  nxDevice,
	}
	foreign := func(device string) resourceapi.DeviceRequestAllocationResult {
		return resourceapi.DeviceRequestAllocationResult{
			Request: "gpu",
			Driver:  "gpu.example.com",
			Pool:    "gpu-pool",
			Device:  device,
		}
	}

	tests := []struct {
		name    string
		results []resourceapi.DeviceRequestAllocationResult
	}{
		// The foreign device is not locally allocatable; the old code failed loudly.
		{"foreign device not allocatable", []resourceapi.DeviceRequestAllocationResult{foreign("gpu-0"), localResult}},
		// The foreign device reuses the local device name (valid across drivers); the
		// old code passed the allocatable lookup and silently prepared it as an Nx device.
		{"foreign name collides, foreign first", []resourceapi.DeviceRequestAllocationResult{foreign(nxDevice), localResult}},
		{"foreign name collides, local first", []resourceapi.DeviceRequestAllocationResult{localResult, foreign(nxDevice)}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := &DeviceState{
				allocatable: AllocatableDevices{nxDevice: resourceapi.Device{Name: nxDevice}},
				cdi:         &CDIHandler{},
			}
			claim := &resourceapi.ResourceClaim{
				Status: resourceapi.ResourceClaimStatus{
					Allocation: &resourceapi.AllocationResult{
						Devices: resourceapi.DeviceAllocationResult{Results: tc.results},
					},
				},
			}

			prepared, err := state.prepareDevices(claim)
			if err != nil {
				t.Fatalf("prepareDevices failed on a multi-driver claim: %v", err)
			}
			if len(prepared) != 1 {
				t.Fatalf("prepared %d devices, want exactly one local Nx device: %#v", len(prepared), prepared)
			}
			got := prepared[0]
			if got.DeviceName != nxDevice || got.PoolName != "nx-pool" ||
				len(got.RequestNames) != 1 || got.RequestNames[0] != "nx" {
				t.Fatalf("prepared %#v, want request=nx pool=nx-pool device=%s", got, nxDevice)
			}
		})
	}
}
