/*
 * Copyright 2025 - IBM Corporation. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"fmt"
	"maps"

	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/dynamic-resource-allocation/resourceslice"
	"k8s.io/klog/v2"
)

type driver struct {
	client coreclientset.Interface
	helper *kubeletplugin.Helper
	state  *DeviceState
}

func NewDriver(ctx context.Context, config *Config) (*driver, error) {
	driver := &driver{
		client: config.coreclient,
	}

	state, err := NewDeviceState(config)
	if err != nil {
		return nil, err
	}
	driver.state = state

	helper, err := kubeletplugin.Start(
		ctx,
		driver,
		kubeletplugin.KubeClient(config.coreclient),
		kubeletplugin.NodeName(config.flags.nodeName),
		kubeletplugin.DriverName(DriverName),
		kubeletplugin.RegistrarDirectoryPath("/var/lib/kubelet/plugins_registry"),
		kubeletplugin.PluginDataDirectoryPath(DriverPluginPath),
		kubeletplugin.GRPCVerbosity(99))
	if err != nil {
		return nil, err
	}
	driver.helper = helper

	devices := make([]resourcev1.Device, 0, len(state.allocatable))
	for device := range maps.Values(state.allocatable) {
		v1Device := resourcev1.Device{
			Name: device.Name,
		}
		devices = append(devices, v1Device)
	}
	resources := resourceslice.DriverResources{
		Pools: map[string]resourceslice.Pool{
			config.flags.nodeName: {
				Slices: []resourceslice.Slice{
					{
						Devices: devices,
					},
				},
			},
		},
	}

	if err := helper.PublishResources(ctx, resources); err != nil {
		return nil, err
	}

	return driver, nil
}

func (d *driver) Shutdown(ctx context.Context) error {
	d.helper.Stop()
	return nil
}

func (d *driver) PrepareResourceClaims(ctx context.Context, claims []*resourcev1.ResourceClaim) (map[types.UID]kubeletplugin.PrepareResult, error) {
	klog.Infof("PrepareResourceClaims is called: number of claims: %d", len(claims))
	result := make(map[types.UID]kubeletplugin.PrepareResult)

	d.state.cdi.cache.Refresh()
	for _, claim := range claims {
		result[claim.UID] = d.prepareResourceClaim(ctx, claim)
	}

	return result, nil
}

func (d *driver) prepareResourceClaim(_ context.Context, claim *resourcev1.ResourceClaim) kubeletplugin.PrepareResult {
	klog.Infof("prepareResourceClaim called for claim: UID=%v, Name=%v/%v", claim.UID, claim.Namespace, claim.Name)

	// Log allocation status details
	if claim.Status.Allocation != nil {
		klog.Infof("Claim has allocation with %d device results", len(claim.Status.Allocation.Devices.Results))

		// Log each device result for debugging
		for i, result := range claim.Status.Allocation.Devices.Results {
			klog.Infof("Device result %d: Request=%v, Driver=%v, Pool=%v, Device=%v",
				i, result.Request, result.Driver, result.Pool, result.Device)
		}
	} else {
		klog.Warningf("Claim %v has no allocation status", claim.UID)
	}

	preparedPBs, err := d.state.Prepare(claim)
	if err != nil {
		klog.Errorf("Failed to prepare devices for claim %v: %v", claim.UID, err)
		return kubeletplugin.PrepareResult{
			Err: fmt.Errorf("error preparing devices for claim %v: %w", claim.UID, err),
		}
	}
	var prepared []kubeletplugin.Device
	for _, preparedPB := range preparedPBs {
		prepared = append(prepared, kubeletplugin.Device{
			Requests:     preparedPB.GetRequestNames(),
			PoolName:     preparedPB.GetPoolName(),
			DeviceName:   preparedPB.GetDeviceName(),
			CDIDeviceIDs: preparedPB.GetCdiDeviceIds(),
		})
	}

	klog.Infof("Returning newly prepared devices for claim '%v': %v", claim.UID, prepared)
	return kubeletplugin.PrepareResult{Devices: prepared}
}

func (d *driver) UnprepareResourceClaims(ctx context.Context, claims []kubeletplugin.NamespacedObject) (map[types.UID]error, error) {
	klog.Infof("UnprepareResourceClaims is called: number of claims: %d", len(claims))
	result := make(map[types.UID]error)

	d.state.cdi.cache.Refresh()

	for _, claim := range claims {
		result[claim.UID] = d.unprepareResourceClaim(ctx, claim)
	}

	return result, nil
}

func (d *driver) unprepareResourceClaim(_ context.Context, claim kubeletplugin.NamespacedObject) error {
	if err := d.state.Unprepare(string(claim.UID)); err != nil {
		return fmt.Errorf("error unpreparing devices for claim %v: %w", claim.UID, err)
	}

	return nil
}

func (d *driver) HandleError(ctx context.Context, err error, claimUID string) {
	klog.ErrorS(err, "Error handling resource claim", "claimUID", claimUID)
}
