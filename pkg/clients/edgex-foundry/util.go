/*
Copyright 2021 The OpenYurt Authors.

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

package edgex_foundry

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	devicev1alpha1 "github.com/openyurtio/device-controller/apis/device.openyurt.io/v1alpha1"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EdgeXObjectName     = "device-controller/edgex-object.name"
	DeviceServicePath   = "/api/v2/deviceservice"
	DeviceProfilePath   = "/api/v2/deviceprofile"
	DevicePath          = "/api/v2/device"
	CommandResponsePath = "/api/v2/device"

	APIVersionV2 = "v2"
)

type ClientURL struct {
	Host string
	Port int
}

func getEdgeDeviceName(d *devicev1alpha1.Device) string {
	var actualDeviceName string
	if _, ok := d.ObjectMeta.Labels[EdgeXObjectName]; ok {
		actualDeviceName = d.ObjectMeta.Labels[EdgeXObjectName]
	} else {
		actualDeviceName = d.GetName()
	}
	return actualDeviceName
}

func toEdgexDeviceService(ds *devicev1alpha1.DeviceService) dtos.DeviceService {
	return dtos.DeviceService{
		Description:   ds.Spec.Description,
		Name:          ds.GetName(),
		LastConnected: ds.Status.LastConnected,
		LastReported:  ds.Status.LastReported,
		Labels:        ds.Spec.Labels,
		AdminState:    string(ds.Spec.AdminState),
		BaseAddress:   ds.Spec.BaseAddress,
	}
}

func toEdgeXDeviceResourceSlice(drs []devicev1alpha1.DeviceResource) []dtos.DeviceResource {
	var ret []dtos.DeviceResource
	for _, dr := range drs {
		ret = append(ret, toEdgeXDeviceResource(dr))
	}
	return ret
}

func toEdgeXDeviceResource(dr devicev1alpha1.DeviceResource) dtos.DeviceResource {
	genericAttrs := make(map[string]interface{})
	for k, v := range dr.Attributes {
		genericAttrs[k] = v
	}

	return dtos.DeviceResource{
		Description: dr.Description,
		Name:        dr.Name,
		Tag:         dr.Tag,
		Properties:  toEdgeXProfileProperty(dr.Properties),
		Attributes:  genericAttrs,
	}
}

func toEdgeXProfileProperty(pp devicev1alpha1.ResourceProperties) dtos.ResourceProperties {
	return dtos.ResourceProperties{
		ReadWrite:    pp.ReadWrite,
		Minimum:      pp.Minimum,
		Maximum:      pp.Maximum,
		DefaultValue: pp.DefaultValue,
		Mask:         pp.Mask,
		Shift:        pp.Shift,
		Scale:        pp.Scale,
		Offset:       pp.Offset,
		Base:         pp.Base,
		Assertion:    pp.Assertion,
		MediaType:    pp.MediaType,
		Units:        pp.Units,
		ValueType:    pp.ValueType,
	}
}

func toKubeDeviceService(ds dtos.DeviceService) devicev1alpha1.DeviceService {
	return devicev1alpha1.DeviceService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(ds.Name),
			Namespace: "default",
			Labels: map[string]string{
				EdgeXObjectName: ds.Name,
			},
		},
		Spec: devicev1alpha1.DeviceServiceSpec{
			Description: ds.Description,
			Labels:      ds.Labels,
			AdminState:  devicev1alpha1.AdminState(ds.AdminState),
			BaseAddress: ds.BaseAddress,
		},
		Status: devicev1alpha1.DeviceServiceStatus{
			EdgeId:        ds.Id,
			LastConnected: ds.LastConnected,
			LastReported:  ds.LastReported,
			AdminState:    devicev1alpha1.AdminState(ds.AdminState),
		},
	}
}

func toEdgeXDevice(d *devicev1alpha1.Device) dtos.Device {
	md := dtos.Device{
		Description:    d.Spec.Description,
		Name:           d.GetName(),
		AdminState:     string(toEdgeXAdminState(d.Spec.AdminState)),
		OperatingState: string(toEdgeXOperatingState(d.Spec.OperatingState)),
		Protocols:      toEdgeXProtocols(d.Spec.Protocols),
		LastConnected:  d.Status.LastConnected,
		LastReported:   d.Status.LastReported,
		Labels:         d.Spec.Labels,
		Location:       d.Spec.Location,
		ServiceName:    d.Spec.Service,
		ProfileName:    d.Spec.Profile,
	}
	if d.Status.EdgeId != "" {
		md.Id = d.Status.EdgeId
	}
	return md
}

func toEdgeXProtocols(
	pps map[string]devicev1alpha1.ProtocolProperties) map[string]dtos.ProtocolProperties {
	ret := map[string]dtos.ProtocolProperties{}
	for k, v := range pps {
		ret[k] = dtos.ProtocolProperties(v)
	}
	return ret
}

func toEdgeXAdminState(as devicev1alpha1.AdminState) models.AdminState {
	if as == devicev1alpha1.Locked {
		return models.Locked
	}
	return models.Unlocked
}

func toEdgeXOperatingState(os devicev1alpha1.OperatingState) models.OperatingState {
	if os == devicev1alpha1.Up {
		return models.Up
	} else if os == devicev1alpha1.Down {
		return models.Down
	}
	return models.Unknown
}

// toKubeDevice serialize the EdgeX Device to the corresponding Kubernetes Device
func toKubeDevice(ed dtos.Device) devicev1alpha1.Device {
	var loc string
	if ed.Location != nil {
		loc = ed.Location.(string)
	}
	return devicev1alpha1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(ed.Name),
			Namespace: "default",
			Labels: map[string]string{
				EdgeXObjectName: ed.Name,
			},
		},
		Spec: devicev1alpha1.DeviceSpec{
			Description:    ed.Description,
			AdminState:     devicev1alpha1.AdminState(ed.AdminState),
			OperatingState: devicev1alpha1.OperatingState(ed.OperatingState),
			Protocols:      toKubeProtocols(ed.Protocols),
			Labels:         ed.Labels,
			Location:       loc,
			Service:        ed.ServiceName,
			Profile:        ed.ProfileName,
			// TODO: Notify
		},
		Status: devicev1alpha1.DeviceStatus{
			LastConnected:  ed.LastConnected,
			LastReported:   ed.LastReported,
			Synced:         true,
			EdgeId:         ed.Id,
			AdminState:     devicev1alpha1.AdminState(ed.AdminState),
			OperatingState: devicev1alpha1.OperatingState(ed.OperatingState),
		},
	}
}

// toKubeProtocols serialize the EdgeX ProtocolProperties to the corresponding
// Kubernetes OperatingState
func toKubeProtocols(
	eps map[string]dtos.ProtocolProperties) map[string]devicev1alpha1.ProtocolProperties {
	ret := map[string]devicev1alpha1.ProtocolProperties{}
	for k, v := range eps {
		ret[k] = devicev1alpha1.ProtocolProperties(v)
	}
	return ret
}

// toKubeDeviceProfile create DeviceProfile in cloud according to devicProfile in edge
func toKubeDeviceProfile(dp *dtos.DeviceProfile) devicev1alpha1.DeviceProfile {
	return devicev1alpha1.DeviceProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(dp.Name),
			Namespace: "default",
			Labels: map[string]string{
				EdgeXObjectName: dp.Name,
			},
		},
		Spec: devicev1alpha1.DeviceProfileSpec{
			Description:     dp.Description,
			Manufacturer:    dp.Manufacturer,
			Model:           dp.Model,
			Labels:          dp.Labels,
			DeviceResources: toKubeDeviceResources(dp.DeviceResources),
			DeviceCommands:  toKubeDeviceCommand(dp.DeviceCommands),
		},
		Status: devicev1alpha1.DeviceProfileStatus{
			EdgeId: dp.Id,
			Synced: true,
		},
	}
}

func toKubeDeviceCommand(dcs []dtos.DeviceCommand) []devicev1alpha1.DeviceCommand {
	var ret []devicev1alpha1.DeviceCommand
	for _, dc := range dcs {
		ret = append(ret, devicev1alpha1.DeviceCommand{
			Name:               dc.Name,
			ReadWrite:          dc.ReadWrite,
			IsHidden:           dc.IsHidden,
			ResourceOperations: toKubeResourceOperations(dc.ResourceOperations),
		})
	}
	return ret
}

func toEdgeXDeviceCommand(dcs []devicev1alpha1.DeviceCommand) []dtos.DeviceCommand {
	var ret []dtos.DeviceCommand
	for _, dc := range dcs {
		ret = append(ret, dtos.DeviceCommand{
			Name:               dc.Name,
			ReadWrite:          dc.ReadWrite,
			IsHidden:           dc.IsHidden,
			ResourceOperations: toEdgeXResourceOperations(dc.ResourceOperations),
		})
	}
	return ret
}

func toKubeResourceOperations(ros []dtos.ResourceOperation) []devicev1alpha1.ResourceOperation {
	var ret []devicev1alpha1.ResourceOperation
	for _, ro := range ros {
		ret = append(ret, devicev1alpha1.ResourceOperation{
			DeviceResource: ro.DeviceResource,
			Mappings:       ro.Mappings,
			DefaultValue:   ro.DefaultValue,
		})
	}
	return ret
}

func toEdgeXResourceOperations(ros []devicev1alpha1.ResourceOperation) []dtos.ResourceOperation {
	var ret []dtos.ResourceOperation
	for _, ro := range ros {
		ret = append(ret, dtos.ResourceOperation{
			DeviceResource: ro.DeviceResource,
			Mappings:       ro.Mappings,
			DefaultValue:   ro.DefaultValue,
		})
	}
	return ret
}

func toKubeDeviceResources(drs []dtos.DeviceResource) []devicev1alpha1.DeviceResource {
	var ret []devicev1alpha1.DeviceResource
	for _, dr := range drs {
		ret = append(ret, toKubeDeviceResource(dr))
	}
	return ret
}

func toKubeDeviceResource(dr dtos.DeviceResource) devicev1alpha1.DeviceResource {
	concreteAttrs := make(map[string]string)
	for k, v := range dr.Attributes {
		switch v.(type) {
		case string:
			concreteAttrs[k] = v.(string)
			continue
		case int:
			concreteAttrs[k] = fmt.Sprintf("%d", v.(int))
			continue
		case float64:
			concreteAttrs[k] = fmt.Sprintf("%f", v.(float64))
			continue
		case fmt.Stringer:
			concreteAttrs[k] = v.(fmt.Stringer).String()
			continue
		}
	}

	return devicev1alpha1.DeviceResource{
		Description: dr.Description,
		Name:        dr.Name,
		Tag:         dr.Tag,
		IsHidden:    dr.IsHidden,
		Properties:  toKubeProfileProperty(dr.Properties),
		Attributes:  concreteAttrs,
	}
}

func toKubeProfileProperty(rp dtos.ResourceProperties) devicev1alpha1.ResourceProperties {
	return devicev1alpha1.ResourceProperties{
		ValueType:    rp.ValueType,
		ReadWrite:    rp.ReadWrite,
		Minimum:      rp.Minimum,
		Maximum:      rp.Maximum,
		DefaultValue: rp.DefaultValue,
		Mask:         rp.Mask,
		Shift:        rp.Shift,
		Scale:        rp.Scale,
		Offset:       rp.Offset,
		Base:         rp.Base,
		Assertion:    rp.Assertion,
		MediaType:    rp.MediaType,
		Units:        rp.Units,
	}
}

// toEdgeXDeviceProfile create DeviceProfile in edge according to devicProfile in cloud
func toEdgeXDeviceProfile(dp *devicev1alpha1.DeviceProfile) dtos.DeviceProfile {
	return dtos.DeviceProfile{
		Description:     dp.Spec.Description,
		Name:            dp.GetName(),
		Manufacturer:    dp.Spec.Manufacturer,
		Model:           dp.Spec.Model,
		Labels:          dp.Spec.Labels,
		DeviceResources: toEdgeXDeviceResourceSlice(dp.Spec.DeviceResources),
		DeviceCommands:  toEdgeXDeviceCommand(dp.Spec.DeviceCommands),
	}
}

func makeEdgeXDeviceProfilesRequest(dps []*devicev1alpha1.DeviceProfile) []*requests.DeviceProfileRequest {
	var req []*requests.DeviceProfileRequest
	for _, dp := range dps {
		req = append(req, &requests.DeviceProfileRequest{
			BaseRequest: common.BaseRequest{
				Versionable: common.Versionable{
					ApiVersion: APIVersionV2,
				},
			},
			Profile: toEdgeXDeviceProfile(dp),
		})
	}
	return req
}

func makeEdgeXDeviceRequest(devs []*devicev1alpha1.Device) []*requests.AddDeviceRequest {
	var req []*requests.AddDeviceRequest
	for _, dev := range devs {
		req = append(req, &requests.AddDeviceRequest{
			BaseRequest: common.BaseRequest{
				Versionable: common.Versionable{
					ApiVersion: APIVersionV2,
				},
			},
			Device: toEdgeXDevice(dev),
		})
	}
	return req
}

func makeEdgeXDeviceService(dss []*devicev1alpha1.DeviceService) []*requests.AddDeviceServiceRequest {
	var req []*requests.AddDeviceServiceRequest
	for _, ds := range dss {
		req = append(req, &requests.AddDeviceServiceRequest{
			BaseRequest: common.BaseRequest{
				Versionable: common.Versionable{
					ApiVersion: APIVersionV2,
				},
			},
			Service: toEdgexDeviceService(ds),
		})
	}
	return req
}
