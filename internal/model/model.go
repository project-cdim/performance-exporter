// Copyright (C) 2025 NEC Corporation.
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
	CPU              = "CPU"
	Memory           = "memory"
	Storage          = "storage"
	NetworkInterface = "networkInterface"
	Accelerator      = "Accelerator"
	DSP              = "DSP"
	FPGA             = "FPGA"
	GPU              = "GPU"
	UnknownProcessor = "UnknownProcessor"

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

// HW control metric information retrieval result
type HwOutput struct {
	HwMetrics map[string]any
}

// Configuration values from processor.yaml
type ProcessorConfig struct {
	Status                           Status             `yaml:"status"`
	PowerState                       bool               `yaml:"powerState"`
	PowerCapability                  bool               `yaml:"powerCapability"`
	LtssmState                       bool               `yaml:"LTSSMState"`
	UsageRate                        bool               `yaml:"usageRate"`
	User                             bool               `yaml:"user"`
	System                           bool               `yaml:"system"`
	Wait                             bool               `yaml:"wait"`
	Idle                             bool               `yaml:"idle"`
	CPUWatts                         bool               `yaml:"CPUWatts"`
	Turbostat                        bool               `yaml:"turbostat"`
	MetricBandwidthPercent           bool               `yaml:"metricBandwidthPercent"`
	MetricOperatingSpeedMHz          bool               `yaml:"metricOperatingSpeedMHz"`
	MetricLocalMemoryBandwidthBytes  bool               `yaml:"metricLocalMemoryBandwidthBytes"`
	MetricRemoteMemoryBandwidthBytes bool               `yaml:"metricRemoteMemoryBandwidthBytes"`
	MetricEnergyJoules               MetricEnergyJoules `yaml:"metricEnergyJoules"`
}

type Status struct {
	State  bool `yaml:"state"`
	Health bool `yaml:"health"`
}

type MetricEnergyJoules struct {
	SensorResetTime bool `yaml:"sensorResetTime"`
	SensingInterval bool `yaml:"sensingInterval"`
	ReadingTime     bool `yaml:"readingTime"`
	Reading         bool `yaml:"reading"`
}

// Configuration values from memory.yaml
type MemoryConfig struct {
	Enabled                 bool               `yaml:"enabled"`
	Status                  Status             `yaml:"status"`
	PowerState              bool               `yaml:"powerState"`
	PowerCapability         bool               `yaml:"powerCapability"`
	LtssmState              bool               `yaml:"LTSSMState"`
	CXLCapacity             CXLCapacity        `yaml:"CXLCapacity"`
	UsedMemory              bool               `yaml:"usedMemory"`
	Swap                    bool               `yaml:"swap"`
	MetricBandwidthPercent  bool               `yaml:"metricBandwidthPercent"`
	MetricBlockSizeBytes    bool               `yaml:"metricBlockSizeBytes"`
	MetricOperatingSpeedMHz bool               `yaml:"metricOperatingSpeedMHz"`
	MetricHealthData        MetricHealthData   `yaml:"metricHealthData"`
	MetricEnergyJoules      MetricEnergyJoules `yaml:"metricEnergyJoules"`
}

type CXLCapacity struct {
	Volatile   bool `yaml:"volatile"`
	Persistent bool `yaml:"persistent"`
	Total      bool `yaml:"total"`
}

type MetricHealthData struct {
	DataLossDetected              bool `yaml:"dataLossDetected"`
	LastShutdownSuccess           bool `yaml:"lastShutdownSuccess"`
	PerformanceDegraded           bool `yaml:"performanceDegraded"`
	PredictedMediaLifeLeftPercent bool `yaml:"predictedMediaLifeLeftPercent"`
}

// Configuration values from storage.yaml
type StorageConfig struct {
	VolumeCapacity struct {
		Data struct {
			AllocatedBytes   bool `yaml:"allocatedBytes"`
			ConsumedBytes    bool `yaml:"consumedBytes"`
			GuaranteedBytes  bool `yaml:"guaranteedBytes"`
			ProvisionedBytes bool `yaml:"provisionedBytes"`
		} `yaml:"data"`
		Metadata struct {
			AllocatedBytes   bool `yaml:"allocatedBytes"`
			ConsumedBytes    bool `yaml:"consumedBytes"`
			GuaranteedBytes  bool `yaml:"guaranteedBytes"`
			ProvisionedBytes bool `yaml:"provisionedBytes"`
		} `yaml:"metadata"`
		Snapshot struct {
			AllocatedBytes   bool `yaml:"allocatedBytes"`
			ConsumedBytes    bool `yaml:"consumedBytes"`
			GuaranteedBytes  bool `yaml:"guaranteedBytes"`
			ProvisionedBytes bool `yaml:"provisionedBytes"`
		} `yaml:"snapshot"`
	} `yaml:"volumeCapacity"`
	VolumeRemainingCapacityPercent bool   `yaml:"volumeRemainingCapacityPercent"`
	DriveNegotiatedSpeedGbs        bool   `yaml:"driveNegotiatedSpeedGbs"`
	Status                         Status `yaml:"status"`
	PowerState                     bool   `yaml:"powerState"`
	PowerCapability                bool   `yaml:"powerCapability"`
	LtssmState                     bool   `yaml:"LTSSMState"`
	Disk                           struct {
		AmountUsedDisk bool `yaml:"amountUsedDisk"`
		UsageIO        struct {
			ReadCount        bool `yaml:"readCount"`
			WriteCount       bool `yaml:"writeCount"`
			ReadBytes        bool `yaml:"readBytes"`
			WriteBytes       bool `yaml:"writeBytes"`
			ReadTime         bool `yaml:"readTime"`
			WriteTime        bool `yaml:"writeTime"`
			ReadMergedCount  bool `yaml:"readMergedCount"`
			WriteMergedCount bool `yaml:"writeMergedCount"`
			BusyRate         bool `yaml:"busyRate"`
		} `yaml:"usageIO"`
	} `yaml:"disk"`
	MetricEnergyJoules MetricEnergyJoules `yaml:"metricEnergyJoules"`
}

// Configuration values from networkInterface.yaml
type NetworkInterfaceConfig struct {
	DeviceEnabled   bool   `yaml:"deviceEnabled"`
	Status          Status `yaml:"status"`
	PowerState      bool   `yaml:"powerState"`
	PowerCapability bool   `yaml:"powerCapability"`
	LtssmState      bool   `yaml:"LTSSMState"`

	NetworkInterfaceInformation struct {
		NetworkSpeed   bool `yaml:"networkSpeed"`
		NetworkTraffic struct {
			ReceivePackets  bool `yaml:"receivePackets"`
			TransmitPackets bool `yaml:"transmitPackets"`
			BytesSent       bool `yaml:"bytesSent"`
			BytesRecv       bool `yaml:"bytesRecv"`
			Errin           bool `yaml:"errin"`
			Errout          bool `yaml:"errout"`
			Dropin          bool `yaml:"dropin"`
			Dropout         bool `yaml:"dropout"`
		} `yaml:"networkTraffic"`
	} `yaml:"networkInterfaceInformation"`

	MetricsCPUCorePercent         bool               `yaml:"metricsCPUCorePercent"`
	MetricsHostBusRXPercent       bool               `yaml:"metricsHostBusRXPercent"`
	MetricsHostBusTXPercent       bool               `yaml:"metricsHostBusTXPercent"`
	MetricsRXAvgQueueDepthPercent bool               `yaml:"metricsRXAvgQueueDepthPercent"`
	MetricsTXAvgQueueDepthPercent bool               `yaml:"metricsTXAvgQueueDepthPercent"`
	MetricsRXBytes                bool               `yaml:"metricsRXBytes"`
	MetricsRXFrames               bool               `yaml:"metricsRXFrames"`
	MetricsTXBytes                bool               `yaml:"metricsTXBytes"`
	MetricsTXFrames               bool               `yaml:"metricsTXFrames"`
	MetricEnergyJoules            MetricEnergyJoules `yaml:"metricEnergyJoules"`
}
