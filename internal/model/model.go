// Copyright (C) 2025-2026 NEC Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package model

const (
	CPU                      = "CPU"
	Memory                   = "memory"
	Storage                  = "storage"
	NetworkInterface         = "networkInterface"
	Accelerator              = "Accelerator"
	DSP                      = "DSP"
	FPGA                     = "FPGA"
	GPU                      = "GPU"
	UnknownProcessor         = "UnknownProcessor"
	SourceFabricAdapter      = "sourceFabricAdapter"
	DestinationFabricAdapter = "destinationFabricAdapter"

	StatusStateEnabled string = "Enabled"
	StatusHealthOk     string = "OK"
	PowerStateOn       string = "On"
	LtssmStateL0       string = "L0"

	StatusStateEnabledValue   float64 = 1
	StatusStateOtherValue     float64 = 0
	StatusHealthOkValue       float64 = 1
	StatusHealthOtherValue    float64 = 0
	PowerStateOnValue         float64 = 1
	PowerStateOtherValue      float64 = 0
	PowerCapabilityTrueValue  float64 = 1
	PowerCapabilityFalseValue float64 = 0
	LtssmStateL0Value         float64 = 1
	LtssmStateOtherValue      float64 = 0
)

// HwOutput holds metrics information retrieved from hardware.
type HwOutput struct {
	HwMetrics map[string]any
}

// DevicePortListConfig represents the metric acquisition settings for device port list.
type DevicePortListConfig struct {
	LTSSMState bool `yaml:"LTSSMState"`
}

// CommonDeviceConfig represents metric acquisition settings common to all devices.
// It provides fields that are inline-embedded in each device configuration.
type CommonDeviceConfig struct {
	Status          StatusConfig         `yaml:"status"`
	PowerState      bool                 `yaml:"powerState"`
	PowerCapability bool                 `yaml:"powerCapability"`
	DevicePortList  DevicePortListConfig `yaml:"devicePortList"`
}

// ProcessorConfig loads processor metric acquisition settings from processor.yaml.
type ProcessorConfig struct {
	CommonDeviceConfig               `yaml:",inline"`
	UsageRate                        bool                     `yaml:"usageRate"`
	User                             bool                     `yaml:"user"`
	System                           bool                     `yaml:"system"`
	Wait                             bool                     `yaml:"wait"`
	Idle                             bool                     `yaml:"idle"`
	CPUWatts                         bool                     `yaml:"CPUWatts"`
	Turbostat                        bool                     `yaml:"turbostat"`
	MetricBandwidthPercent           bool                     `yaml:"metricBandwidthPercent"`
	MetricOperatingSpeedMHz          bool                     `yaml:"metricOperatingSpeedMHz"`
	MetricLocalMemoryBandwidthBytes  bool                     `yaml:"metricLocalMemoryBandwidthBytes"`
	MetricRemoteMemoryBandwidthBytes bool                     `yaml:"metricRemoteMemoryBandwidthBytes"`
	MetricEnergyJoules               MetricEnergyJoulesConfig `yaml:"metricEnergyJoules"`
}

// StatusConfig represents metric acquisition settings for status.
type StatusConfig struct {
	State  bool `yaml:"state"`
	Health bool `yaml:"health"`
}

// MetricEnergyJoulesConfig represents metric acquisition settings for energy consumption.
type MetricEnergyJoulesConfig struct {
	SensorResetTime bool `yaml:"sensorResetTime"`
	SensingInterval bool `yaml:"sensingInterval"`
	ReadingTime     bool `yaml:"readingTime"`
	Reading         bool `yaml:"reading"`
}

// MemoryConfig loads memory metric acquisition settings from memory.yaml.
type MemoryConfig struct {
	CommonDeviceConfig      `yaml:",inline"`
	Enabled                 bool                     `yaml:"enabled"`
	CXLCapacity             CXLCapacityConfig        `yaml:"CXLCapacity"`
	UsedMemory              bool                     `yaml:"usedMemory"`
	Swap                    bool                     `yaml:"swap"`
	MetricBandwidthPercent  bool                     `yaml:"metricBandwidthPercent"`
	MetricBlockSizeBytes    bool                     `yaml:"metricBlockSizeBytes"`
	MetricOperatingSpeedMHz bool                     `yaml:"metricOperatingSpeedMHz"`
	MetricHealthData        MetricHealthDataConfig   `yaml:"metricHealthData"`
	MetricEnergyJoules      MetricEnergyJoulesConfig `yaml:"metricEnergyJoules"`
}

// CXLCapacityConfig represents metric acquisition settings for CXL memory device capacity.
type CXLCapacityConfig struct {
	Volatile   bool `yaml:"volatile"`
	Persistent bool `yaml:"persistent"`
	Total      bool `yaml:"total"`
}

// MetricHealthDataConfig represents metric acquisition settings for memory health data.
type MetricHealthDataConfig struct {
	DataLossDetected              bool `yaml:"dataLossDetected"`
	LastShutdownSuccess           bool `yaml:"lastShutdownSuccess"`
	PerformanceDegraded           bool `yaml:"performanceDegraded"`
	PredictedMediaLifeLeftPercent bool `yaml:"predictedMediaLifeLeftPercent"`
}

// DataConfig represents metric acquisition settings for capacity information related to data types.
type DataConfig struct {
	AllocatedBytes   bool `yaml:"allocatedBytes"`
	ConsumedBytes    bool `yaml:"consumedBytes"`
	GuaranteedBytes  bool `yaml:"guaranteedBytes"`
	ProvisionedBytes bool `yaml:"provisionedBytes"`
}

// VolumeCapacityConfig represents metric acquisition settings for capacity information of volumes.
type VolumeCapacityConfig struct {
	Data     DataConfig `yaml:"data"`
	Metadata DataConfig `yaml:"metadata"`
	Snapshot DataConfig `yaml:"snapshot"`
}

// UsageIOConfig represents metric acquisition settings for disk I/O statistics.
type UsageIOConfig struct {
	ReadCount        bool `yaml:"readCount"`
	WriteCount       bool `yaml:"writeCount"`
	ReadBytes        bool `yaml:"readBytes"`
	WriteBytes       bool `yaml:"writeBytes"`
	ReadTime         bool `yaml:"readTime"`
	WriteTime        bool `yaml:"writeTime"`
	ReadMergedCount  bool `yaml:"readMergedCount"`
	WriteMergedCount bool `yaml:"writeMergedCount"`
	BusyRate         bool `yaml:"busyRate"`
}

// DiskConfig represents metric acquisition settings for disks.
type DiskConfig struct {
	AmountUsedDisk bool          `yaml:"amountUsedDisk"`
	UsageIO        UsageIOConfig `yaml:"usageIO"`
}

// StorageConfig loads storage metric acquisition settings from storage.yaml.
type StorageConfig struct {
	VolumeCapacity                 VolumeCapacityConfig `yaml:"volumeCapacity"`
	CommonDeviceConfig             `yaml:",inline"`
	VolumeRemainingCapacityPercent bool                     `yaml:"volumeRemainingCapacityPercent"`
	DriveNegotiatedSpeedGbs        bool                     `yaml:"driveNegotiatedSpeedGbs"`
	Disk                           DiskConfig               `yaml:"disk"`
	MetricEnergyJoules             MetricEnergyJoulesConfig `yaml:"metricEnergyJoules"`
}

// NetworkTrafficConfig represents metric acquisition settings for network traffic statistics.
type NetworkTrafficConfig struct {
	ReceivePackets  bool `yaml:"receivePackets"`
	TransmitPackets bool `yaml:"transmitPackets"`
	BytesSent       bool `yaml:"bytesSent"`
	BytesRecv       bool `yaml:"bytesRecv"`
	Errin           bool `yaml:"errin"`
	Errout          bool `yaml:"errout"`
	Dropin          bool `yaml:"dropin"`
	Dropout         bool `yaml:"dropout"`
}

// NetworkInterfaceInformationConfig represents metric acquisition settings for network interface information.
type NetworkInterfaceInformationConfig struct {
	NetworkSpeed   bool                 `yaml:"networkSpeed"`
	NetworkTraffic NetworkTrafficConfig `yaml:"networkTraffic"`
}

// NetworkInterfaceConfig loads network interface metric acquisition settings from networkInterface.yaml.
type NetworkInterfaceConfig struct {
	CommonDeviceConfig            `yaml:",inline"`
	DeviceEnabled                 bool                              `yaml:"deviceEnabled"`
	NetworkInterfaceInformation   NetworkInterfaceInformationConfig `yaml:"networkInterfaceInformation"`
	MetricsCPUCorePercent         bool                              `yaml:"metricsCPUCorePercent"`
	MetricsHostBusRXPercent       bool                              `yaml:"metricsHostBusRXPercent"`
	MetricsHostBusTXPercent       bool                              `yaml:"metricsHostBusTXPercent"`
	MetricsRXAvgQueueDepthPercent bool                              `yaml:"metricsRXAvgQueueDepthPercent"`
	MetricsTXAvgQueueDepthPercent bool                              `yaml:"metricsTXAvgQueueDepthPercent"`
	MetricsRXBytes                bool                              `yaml:"metricsRXBytes"`
	MetricsRXFrames               bool                              `yaml:"metricsRXFrames"`
	MetricsTXBytes                bool                              `yaml:"metricsTXBytes"`
	MetricsTXFrames               bool                              `yaml:"metricsTXFrames"`
	MetricEnergyJoules            MetricEnergyJoulesConfig          `yaml:"metricEnergyJoules"`
}
