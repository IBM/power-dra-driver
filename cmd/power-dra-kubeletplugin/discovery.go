/*
 * Copyright 2025 - IBM Corporation. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"
	"strconv"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

const CapsLocation = "/host-sys/devices/vio/ibm,compression-v1/nx_gzip_caps"

// Adds nx-gzip if capabilities are in DeviceTree
func enumerateAllPossibleDevices(numNx int) (AllocatableDevices, error) {
	alldevices := make(AllocatableDevices)

	// If nx-gzip doesn't exist, then don't return any devices.
	if !existsNxGzip() {
		return alldevices, nil
	}

	// prepopulate with a single nx-gzip
	device := resourceapi.Device{
		Name: "cryptonxgzip",
		Attributes: map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
			"index": {
				IntValue: ptr.To(int64(0)),
			},
		},
		Capacity: map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
			"nxgzip": {
				Value: resource.MustParse(strconv.Itoa(numNx)),
			},
		},
	}
	alldevices[device.Name] = device

	return alldevices, nil
}

// Detect NXGZIPCAPS exists
func existsNxGzip() bool {
	_, err := os.Stat(CapsLocation)
	if err != nil {
		klog.V(5).ErrorS(err, "Failed to detect Nest Accelerator nx-gzip feature")
		return false
	}
	return true
}
