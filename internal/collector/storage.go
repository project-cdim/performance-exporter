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
        
package collector

import (
	"net/http"
	"os"
	"strings"

	"github.com/project-cdim/performance-exporter/internal/model"
	"github.com/project-cdim/performance-exporter/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

const (
	storageYamlFilePath string = "configs/storage.yaml"
)

// Definition of Metrics
type storageMetrics struct {
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	volumeCapacityDataAllocatedBytes   prometheus.Gauge
	volumeCapacityDataConsumedBytes    prometheus.Gauge
	volumeCapacityDataGuaranteedBytes  prometheus.Gauge
	volumeCapacityDataProvisionedBytes prometheus.Gauge

	volumeCapacityMetadataAllocatedBytes   prometheus.Gauge
	volumeCapacityMetadataConsumedBytes    prometheus.Gauge
	volumeCapacityMetadataGuaranteedBytes  prometheus.Gauge
	volumeCapacityMetadataProvisionedBytes prometheus.Gauge

	volumeCapacitySnapshotAllocatedBytes   prometheus.Gauge
	volumeCapacitySnapshotConsumedBytes    prometheus.Gauge
	volumeCapacitySnapshotGuaranteedBytes  prometheus.Gauge
	volumeCapacitySnapshotProvisionedBytes prometheus.Gauge

	volumeRemainingCapacityPercent prometheus.Gauge
	driveNegotiatedSpeedGbs        prometheus.Gauge
	statusStateEnabled             prometheus.Gauge
	statusHealthOk                 prometheus.Gauge
	powerStateOn                   prometheus.Gauge
	powerCapabilityTrue            prometheus.Gauge
	ltssmStateL0                   prometheus.Gauge

	diskAmountUsedDisk          prometheus.Gauge
	diskUsageIoReadCount        prometheus.Gauge
	diskUsageIoWriteCount       prometheus.Gauge
	diskUsageIoReadBytes        prometheus.Gauge
	diskUsageIoWriteBytes       prometheus.Gauge
	diskUsageIoReadTime         prometheus.Gauge
	diskUsageIoWriteTime        prometheus.Gauge
	diskUsageIoReadMergedCount  prometheus.Gauge
	diskUsageIoWriteMergedCount prometheus.Gauge
	diskUsageIoBusyRate         prometheus.Gauge

	metricEnergyJoulesSensorResetTime prometheus.Gauge
	metricEnergyJoulesSensingInterval prometheus.Gauge
	metricEnergyJoulesReadingTime     prometheus.Gauge
	metricEnergyJoulesReading         prometheus.Gauge
}

// Define storage metrics.
// Descriptor of metrics "One of the pieces of information embedded in metrics (name, information to be placed in #HELP), other than the numerical values to be displayed later in the graph"
func NewStorageMetrics(reg prometheus.Registerer, hwOutput *model.HwOutput, settings *model.StorageConfig) (*storageMetrics, error) {
	m := &storageMetrics{
		// deviceID is not registered as a metric because it uses the job name
		// type is not registered as a metric because it uses the namespace
		// attribute is not registered as a metric because it is a fixed value

		volumeCapacityDataAllocatedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_data_allocatedBytes",
			Help:      "The number of bytes currently allocated by the storage for this data type",
		}),
		volumeCapacityDataConsumedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_data_consumedBytes",
			Help:      "The number of bytes consumed for this data type.",
		}),
		volumeCapacityDataGuaranteedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_data_guaranteedBytes",
			Help:      "The number of bytes guaranteed by the storage for this data type",
		}),
		volumeCapacityDataProvisionedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_data_provisionedBytes",
			Help:      "The maximum number of bytes that can be allocated for this data type",
		}),

		volumeCapacityMetadataAllocatedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_metadata_allocatedBytes",
			Help:      "The number of bytes currently allocated by the storage for this data type",
		}),
		volumeCapacityMetadataConsumedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_metadata_consumedBytes",
			Help:      "The number of bytes consumed for this data type.",
		}),
		volumeCapacityMetadataGuaranteedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_metadata_guaranteedBytes",
			Help:      "The number of bytes guaranteed by the storage for this data type",
		}),
		volumeCapacityMetadataProvisionedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_metadata_provisionedBytes",
			Help:      "The maximum number of bytes that can be allocated for this data type",
		}),

		volumeCapacitySnapshotAllocatedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_snapshot_allocatedBytes",
			Help:      "The number of bytes currently allocated by the storage for this data type",
		}),
		volumeCapacitySnapshotConsumedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_snapshot_consumedBytes",
			Help:      "The number of bytes consumed for this data type.",
		}),
		volumeCapacitySnapshotGuaranteedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_snapshot_guaranteedBytes",
			Help:      "The number of bytes guaranteed by the storage for this data type",
		}),
		volumeCapacitySnapshotProvisionedBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeCapacity_snapshot_provisionedBytes",
			Help:      "The maximum number of bytes that can be allocated for this data type",
		}),

		volumeRemainingCapacityPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "volumeRemainingCapacityPercent",
			Help:      "Percentage of remaining capacity on this volume",
		}),
		driveNegotiatedSpeedGbs: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "driveNegotiatedSpeedGbs",
			Help:      "The current speed at which this drive communicates with the storage",
		}),
		statusStateEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Storage,
			Name:        "status_state",
			Help:        "Storage status information. Resource status",
			ConstLabels: prometheus.Labels{"value": "Enabled"},
		}),
		statusHealthOk: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Storage,
			Name:        "status_health",
			Help:        "Storage status information. Resource health status",
			ConstLabels: prometheus.Labels{"value": "OK"},
		}),
		powerStateOn: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Storage,
			Name:        "powerState",
			Help:        "Current power state of the drive.",
			ConstLabels: prometheus.Labels{"value": "On"},
		}),
		powerCapabilityTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Storage,
			Name:        "powerCapability",
			Help:        "Whether power control function is enabled.",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		ltssmStateL0: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Storage,
			Name:        "LTSSMState",
			Help:        "CXL device link state",
			ConstLabels: prometheus.Labels{"value": "L0"},
		}),

		diskAmountUsedDisk: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_amountUsedDisk",
			Help:      "Disk usage (kilobyte).",
		}),
		diskUsageIoReadCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_readCount",
			Help:      "Number of reads per disk.",
		}),
		diskUsageIoWriteCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_writeCount",
			Help:      "Number of writes per disk.",
		}),
		diskUsageIoReadBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_readBytes",
			Help:      "Number of bytes read per disk (cumulative).",
		}),
		diskUsageIoWriteBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_writeBytes",
			Help:      "Number of bytes written per disk (cumulative).",
		}),
		diskUsageIoReadTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_readTime",
			Help:      "Time spent reading from disk (in milliseconds).",
		}),
		diskUsageIoWriteTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_writeTime",
			Help:      "Time spent writing to disk (in milliseconds).",
		}),
		diskUsageIoReadMergedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_readMergedCount",
			Help:      "Number of merged reads per disk.",
		}),
		diskUsageIoWriteMergedCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_writeMergedCount",
			Help:      "Number of merged writes per disk.",
		}),
		diskUsageIoBusyRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "disk_usageIO_busyRate",
			Help:      "Disk busy rate, the ratio of time spent on actual I/O (%).",
		}),

		metricEnergyJoulesSensorResetTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "metricEnergyJoules_sensorResetTime",
			Help:      "The date and time when the time property was last reset. (UTC)",
		}),
		metricEnergyJoulesSensingInterval: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "metricEnergyJoules_sensingInterval",
			Help:      "Time interval between sensor readings (seconds).",
		}),
		metricEnergyJoulesReadingTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "metricEnergyJoules_readingTime",
			Help:      "The date and time when the measurement was obtained from the sensor. (UTC)",
		}),
		metricEnergyJoulesReading: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Storage,
			Name:      "metricEnergyJoules_reading",
			Help:      "Measurement of energy consumption (J).",
		}),
	}

	errorItems := []string{}
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	if service.IsExistValue(hwOutput.HwMetrics, "volumeCapacity") {
		if volumeCapacity, ok := hwOutput.HwMetrics["volumeCapacity"].(map[string]any); ok {
			if service.IsExistValue(volumeCapacity, "data") {
				if data, ok := volumeCapacity["data"].(map[string]any); ok {
					if settings.VolumeCapacity.Data.AllocatedBytes && service.IsExistValue(data, "allocatedBytes") {
						reg.MustRegister(m.volumeCapacityDataAllocatedBytes)
					}
					if settings.VolumeCapacity.Data.ConsumedBytes && service.IsExistValue(data, "consumedBytes") {
						reg.MustRegister(m.volumeCapacityDataConsumedBytes)
					}
					if settings.VolumeCapacity.Data.GuaranteedBytes && service.IsExistValue(data, "guaranteedBytes") {
						reg.MustRegister(m.volumeCapacityDataGuaranteedBytes)
					}
					if settings.VolumeCapacity.Data.ProvisionedBytes && service.IsExistValue(data, "provisionedBytes") {
						reg.MustRegister(m.volumeCapacityDataProvisionedBytes)
					}
				} else {
					errorItems = append(errorItems, "data")
				}
			}
			if service.IsExistValue(volumeCapacity, "metadata") {
				if metadata, ok := volumeCapacity["metadata"].(map[string]any); ok {
					if settings.VolumeCapacity.Metadata.AllocatedBytes && service.IsExistValue(metadata, "allocatedBytes") {
						reg.MustRegister(m.volumeCapacityMetadataAllocatedBytes)
					}
					if settings.VolumeCapacity.Metadata.ConsumedBytes && service.IsExistValue(metadata, "consumedBytes") {
						reg.MustRegister(m.volumeCapacityMetadataConsumedBytes)
					}
					if settings.VolumeCapacity.Metadata.GuaranteedBytes && service.IsExistValue(metadata, "guaranteedBytes") {
						reg.MustRegister(m.volumeCapacityMetadataGuaranteedBytes)
					}
					if settings.VolumeCapacity.Metadata.ProvisionedBytes && service.IsExistValue(metadata, "provisionedBytes") {
						reg.MustRegister(m.volumeCapacityMetadataProvisionedBytes)
					}
				} else {
					errorItems = append(errorItems, "metadata")
				}
			}
			if service.IsExistValue(volumeCapacity, "snapshot") {
				if snapshot, ok := volumeCapacity["snapshot"].(map[string]any); ok {
					if settings.VolumeCapacity.Snapshot.AllocatedBytes && service.IsExistValue(snapshot, "allocatedBytes") {
						reg.MustRegister(m.volumeCapacitySnapshotAllocatedBytes)
					}
					if settings.VolumeCapacity.Snapshot.ConsumedBytes && service.IsExistValue(snapshot, "consumedBytes") {
						reg.MustRegister(m.volumeCapacitySnapshotConsumedBytes)
					}
					if settings.VolumeCapacity.Snapshot.GuaranteedBytes && service.IsExistValue(snapshot, "guaranteedBytes") {
						reg.MustRegister(m.volumeCapacitySnapshotGuaranteedBytes)
					}
					if settings.VolumeCapacity.Snapshot.ProvisionedBytes && service.IsExistValue(snapshot, "provisionedBytes") {
						reg.MustRegister(m.volumeCapacitySnapshotProvisionedBytes)
					}
				} else {
					errorItems = append(errorItems, "snapshot")
					return m, service.CreateErroeMessage("snapshot")
				}
			}
		} else {
			errorItems = append(errorItems, "volumeCapacity")
		}
	}

	if settings.VolumeRemainingCapacityPercent && service.IsExistValue(hwOutput.HwMetrics, "volumeRemainingCapacityPercent") {
		reg.MustRegister(m.volumeRemainingCapacityPercent)
	}
	if settings.DriveNegotiatedSpeedGbs && service.IsExistValue(hwOutput.HwMetrics, "driveNegotiatedSpeedGbs") {
		reg.MustRegister(m.driveNegotiatedSpeedGbs)
	}

	if service.IsExistValue(hwOutput.HwMetrics, "status") {
		if status, ok := hwOutput.HwMetrics["status"].(map[string]any); ok {
			if settings.Status.State && service.IsExistValue(status, "state") {
				reg.MustRegister(m.statusStateEnabled)
			}
			if settings.Status.Health && service.IsExistValue(status, "health") {
				reg.MustRegister(m.statusHealthOk)
			}
		} else {
			errorItems = append(errorItems, "status")
		}
	}

	if settings.PowerState && service.IsExistValue(hwOutput.HwMetrics, "powerState") {
		reg.MustRegister(m.powerStateOn)
	}
	if settings.PowerCapability && service.IsExistValue(hwOutput.HwMetrics, "powerCapability") {
		reg.MustRegister(m.powerCapabilityTrue)
	}
	if settings.LtssmState && service.IsExistValue(hwOutput.HwMetrics, "LTSSMState") {
		reg.MustRegister(m.ltssmStateL0)
	}

	if service.IsExistValue(hwOutput.HwMetrics, "disk") {
		if disk, ok := hwOutput.HwMetrics["disk"].(map[string]any); ok {
			if settings.Disk.AmountUsedDisk && service.IsExistValue(disk, "amountUsedDisk") {
				reg.MustRegister(m.diskAmountUsedDisk)
			}
			if service.IsExistValue(disk, "usageIO") {
				if usageIO, ok := disk["usageIO"].(map[string]any); ok {
					if settings.Disk.UsageIO.ReadCount && service.IsExistValue(usageIO, "readCount") {
						reg.MustRegister(m.diskUsageIoReadCount)
					}
					if settings.Disk.UsageIO.WriteCount && service.IsExistValue(usageIO, "writeCount") {
						reg.MustRegister(m.diskUsageIoWriteCount)
					}
					if settings.Disk.UsageIO.ReadBytes && service.IsExistValue(usageIO, "readBytes") {
						reg.MustRegister(m.diskUsageIoReadBytes)
					}
					if settings.Disk.UsageIO.WriteBytes && service.IsExistValue(usageIO, "writeBytes") {
						reg.MustRegister(m.diskUsageIoWriteBytes)
					}
					if settings.Disk.UsageIO.ReadTime && service.IsExistValue(usageIO, "readTime") {
						reg.MustRegister(m.diskUsageIoReadTime)
					}
					if settings.Disk.UsageIO.WriteTime && service.IsExistValue(usageIO, "writeTime") {
						reg.MustRegister(m.diskUsageIoWriteTime)
					}
					if settings.Disk.UsageIO.ReadMergedCount && service.IsExistValue(usageIO, "readMergedCount") {
						reg.MustRegister(m.diskUsageIoReadMergedCount)
					}
					if settings.Disk.UsageIO.WriteMergedCount && service.IsExistValue(usageIO, "writeMergedCount") {
						reg.MustRegister(m.diskUsageIoWriteMergedCount)
					}
					if settings.Disk.UsageIO.BusyRate && service.IsExistValue(usageIO, "busyRate") {
						reg.MustRegister(m.diskUsageIoBusyRate)
					}
				} else {
					errorItems = append(errorItems, "usageIO")
				}
			}
		} else {
			errorItems = append(errorItems, "disk")
		}
	}

	if service.IsExistValue(hwOutput.HwMetrics, "metricEnergyJoules") {
		if metricEnergyJoules, ok := hwOutput.HwMetrics["metricEnergyJoules"].(map[string]any); ok {
			if settings.MetricEnergyJoules.SensorResetTime && service.IsExistValue(metricEnergyJoules, "sensorResetTime") {
				reg.MustRegister(m.metricEnergyJoulesSensorResetTime)
			}
			if settings.MetricEnergyJoules.SensingInterval && service.IsExistValue(metricEnergyJoules, "sensingInterval") {
				reg.MustRegister(m.metricEnergyJoulesSensingInterval)
			}
			if settings.MetricEnergyJoules.ReadingTime && service.IsExistValue(metricEnergyJoules, "readingTime") {
				reg.MustRegister(m.metricEnergyJoulesReadingTime)
			}
			if settings.MetricEnergyJoules.Reading && service.IsExistValue(metricEnergyJoules, "reading") {
				reg.MustRegister(m.metricEnergyJoulesReading)
			}
		} else {
			errorItems = append(errorItems, "metricEnergyJoules")
		}
	}

	if len(errorItems) > 0 {
		return m, service.CreateErroeMessage(strings.Join(errorItems, "/"))
	}
	return m, nil
}

// Set the metrics values for storage.
func setStorageMetrics(m *storageMetrics, hwOutput *model.HwOutput, settings *model.StorageConfig) error {
	errorItems := []string{}

	// attribute is not registered because it is a fixed value.

	if service.IsExistValue(hwOutput.HwMetrics, "volumeCapacity") {
		if volumeCapacity, ok := hwOutput.HwMetrics["volumeCapacity"].(map[string]any); ok {
			if service.IsExistValue(volumeCapacity, "data") {
				if data, ok := volumeCapacity["data"].(map[string]any); ok {
					if settings.VolumeCapacity.Data.AllocatedBytes && service.IsExistValue(data, "allocatedBytes") {
						if allocatedBytes, ok := data["allocatedBytes"].(float64); ok {
							m.volumeCapacityDataAllocatedBytes.Set(allocatedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_data_allocatedBytes")
						}
					}
					if settings.VolumeCapacity.Data.ConsumedBytes && service.IsExistValue(data, "consumedBytes") {
						if consumedBytes, ok := data["consumedBytes"].(float64); ok {
							m.volumeCapacityDataConsumedBytes.Set(consumedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_data_consumedBytes")
						}
					}
					if settings.VolumeCapacity.Data.GuaranteedBytes && service.IsExistValue(data, "guaranteedBytes") {
						if guaranteedBytes, ok := data["guaranteedBytes"].(float64); ok {
							m.volumeCapacityDataGuaranteedBytes.Set(guaranteedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_data_guaranteedBytes")
						}
					}
					if settings.VolumeCapacity.Data.ProvisionedBytes && service.IsExistValue(data, "provisionedBytes") {
						if provisionedBytes, ok := data["provisionedBytes"].(float64); ok {
							m.volumeCapacityDataProvisionedBytes.Set(provisionedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_data_provisionedBytes")
						}
					}
				} else {
					errorItems = append(errorItems, "volumeCapacity_data")
				}
			}
			if service.IsExistValue(volumeCapacity, "metadata") {
				if metadata, ok := volumeCapacity["metadata"].(map[string]any); ok {
					if settings.VolumeCapacity.Metadata.AllocatedBytes && service.IsExistValue(metadata, "allocatedBytes") {
						if allocatedBytes, ok := metadata["allocatedBytes"].(float64); ok {
							m.volumeCapacityMetadataAllocatedBytes.Set(allocatedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_metadata_allocatedBytes")
						}
					}
					if settings.VolumeCapacity.Metadata.ConsumedBytes && service.IsExistValue(metadata, "consumedBytes") {
						if consumedBytes, ok := metadata["consumedBytes"].(float64); ok {
							m.volumeCapacityMetadataConsumedBytes.Set(consumedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_metadata_consumedBytes")
						}
					}
					if settings.VolumeCapacity.Metadata.GuaranteedBytes && service.IsExistValue(metadata, "guaranteedBytes") {
						if guaranteedBytes, ok := metadata["guaranteedBytes"].(float64); ok {
							m.volumeCapacityMetadataGuaranteedBytes.Set(guaranteedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_metadata_guaranteedBytes")
						}
					}
					if settings.VolumeCapacity.Metadata.ProvisionedBytes && service.IsExistValue(metadata, "provisionedBytes") {
						if provisionedBytes, ok := metadata["provisionedBytes"].(float64); ok {
							m.volumeCapacityMetadataProvisionedBytes.Set(provisionedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_metadata_provisionedBytes")
						}
					}
				} else {
					errorItems = append(errorItems, "volumeCapacity_metadata")
				}
			}
			if service.IsExistValue(volumeCapacity, "snapshot") {
				if snapshot, ok := volumeCapacity["snapshot"].(map[string]any); ok {
					if settings.VolumeCapacity.Snapshot.AllocatedBytes && service.IsExistValue(snapshot, "allocatedBytes") {
						if allocatedBytes, ok := snapshot["allocatedBytes"].(float64); ok {
							m.volumeCapacitySnapshotAllocatedBytes.Set(allocatedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_snapshot_allocatedBytes")
						}
					}
					if settings.VolumeCapacity.Snapshot.ConsumedBytes && service.IsExistValue(snapshot, "consumedBytes") {
						if consumedBytes, ok := snapshot["consumedBytes"].(float64); ok {
							m.volumeCapacitySnapshotConsumedBytes.Set(consumedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_snapshot_consumedBytes")
						}
					}
					if settings.VolumeCapacity.Snapshot.GuaranteedBytes && service.IsExistValue(snapshot, "guaranteedBytes") {
						if guaranteedBytes, ok := snapshot["guaranteedBytes"].(float64); ok {
							m.volumeCapacitySnapshotGuaranteedBytes.Set(guaranteedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_snapshot_guaranteedBytes")
						}
					}
					if settings.VolumeCapacity.Snapshot.ProvisionedBytes && service.IsExistValue(snapshot, "provisionedBytes") {
						if provisionedBytes, ok := snapshot["provisionedBytes"].(float64); ok {
							m.volumeCapacitySnapshotProvisionedBytes.Set(provisionedBytes)
						} else {
							errorItems = append(errorItems, "volumeCapacity_snapshot_provisionedBytes")
						}
					}
				} else {
					errorItems = append(errorItems, "volumeCapacity_snapshot")
				}
			}
		} else {
			errorItems = append(errorItems, "volumeCapacity")
		}
	}

	if settings.VolumeRemainingCapacityPercent && service.IsExistValue(hwOutput.HwMetrics, "volumeRemainingCapacityPercent") {
		if volumeRemainingCapacityPercent, ok := hwOutput.HwMetrics["volumeRemainingCapacityPercent"].(float64); ok {
			m.volumeRemainingCapacityPercent.Set(volumeRemainingCapacityPercent)
		} else {
			errorItems = append(errorItems, "volumeRemainingCapacityPercent")
		}
	}
	if settings.DriveNegotiatedSpeedGbs && service.IsExistValue(hwOutput.HwMetrics, "driveNegotiatedSpeedGbs") {
		if driveNegotiatedSpeedGbs, ok := hwOutput.HwMetrics["driveNegotiatedSpeedGbs"].(float64); ok {
			m.driveNegotiatedSpeedGbs.Set(driveNegotiatedSpeedGbs)
		} else {
			errorItems = append(errorItems, "driveNegotiatedSpeedGbs")
		}
	}

	if service.IsExistValue(hwOutput.HwMetrics, "status") {
		if status, ok := hwOutput.HwMetrics["status"].(map[string]any); ok {
			if settings.Status.State && service.IsExistValue(status, "state") {
				if state, ok := status["state"].(string); ok {
					switch state {
					case model.StatusStateEnabled:
						m.statusStateEnabled.Set(model.StatusStateEnabledValue)
					default:
						m.statusStateEnabled.Set(model.StatusStateOtherValue)
					}
				} else {
					errorItems = append(errorItems, "status_state")
				}
			}
			if settings.Status.Health && service.IsExistValue(status, "health") {
				if health, ok := status["health"].(string); ok {
					switch health {
					case model.StatusHealthOk:
						m.statusHealthOk.Set(model.StatusHealthOkValue)
					default:
						m.statusHealthOk.Set(model.StatusHealthOtherValue)
					}
				} else {
					errorItems = append(errorItems, "status_health")
				}
			}
		} else {
			errorItems = append(errorItems, "status")
		}
	}
	if settings.PowerState && service.IsExistValue(hwOutput.HwMetrics, "powerState") {
		if powerState, ok := hwOutput.HwMetrics["powerState"].(string); ok {
			switch powerState {
			case model.PowerStateOn:
				m.powerStateOn.Set(model.PowerStateOnValue)
			default:
				m.powerStateOn.Set(model.PowerStateOtherValue)
			}
		} else {
			errorItems = append(errorItems, "powerState")
		}
	}
	if settings.PowerCapability && service.IsExistValue(hwOutput.HwMetrics, "powerCapability") {
		if powerCapability, ok := hwOutput.HwMetrics["powerCapability"].(bool); ok {
			powerCapabilityValue := model.PowerCapabilityFalseValue
			if powerCapability {
				powerCapabilityValue = model.PowerCapabilityTrueValue
			}
			m.powerCapabilityTrue.Set(powerCapabilityValue)
		} else {
			errorItems = append(errorItems, "powerCapability")
		}
	}
	if settings.LtssmState && service.IsExistValue(hwOutput.HwMetrics, "LTSSMState") {
		if ltssmState, ok := hwOutput.HwMetrics["LTSSMState"].(string); ok {
			switch ltssmState {
			case model.LtssmStateL0:
				m.ltssmStateL0.Set(model.LtssmStateL0Value)
			default:
				m.ltssmStateL0.Set(model.LtssmStateOtherValue)
			}
		} else {
			errorItems = append(errorItems, "LTSSMState")
		}
	}

	if service.IsExistValue(hwOutput.HwMetrics, "disk") {
		if disk, ok := hwOutput.HwMetrics["disk"].(map[string]any); ok {
			if settings.Disk.AmountUsedDisk && service.IsExistValue(disk, "amountUsedDisk") {
				if amountUsedDisk, ok := disk["amountUsedDisk"].(float64); ok {
					m.diskAmountUsedDisk.Set(amountUsedDisk)
				} else {
					errorItems = append(errorItems, "amountUsedDisk")
				}
			}
			if service.IsExistValue(disk, "usageIO") {
				if usageIO, ok := disk["usageIO"].(map[string]any); ok {
					if settings.Disk.UsageIO.ReadCount && service.IsExistValue(usageIO, "readCount") {
						if readCount, ok := usageIO["readCount"].(float64); ok {
							m.diskUsageIoReadCount.Set(readCount)
						} else {
							errorItems = append(errorItems, "readCount")
						}
					}
					if settings.Disk.UsageIO.WriteCount && service.IsExistValue(usageIO, "writeCount") {
						if writeCount, ok := usageIO["writeCount"].(float64); ok {
							m.diskUsageIoWriteCount.Set(writeCount)
						} else {
							errorItems = append(errorItems, "writeCount")
						}
					}
					if settings.Disk.UsageIO.ReadBytes && service.IsExistValue(usageIO, "readBytes") {
						if readBytes, ok := usageIO["readBytes"].(float64); ok {
							m.diskUsageIoReadBytes.Set(readBytes)
						} else {
							errorItems = append(errorItems, "readBytes")
						}
					}
					if settings.Disk.UsageIO.WriteBytes && service.IsExistValue(usageIO, "writeBytes") {
						if writeBytes, ok := usageIO["writeBytes"].(float64); ok {
							m.diskUsageIoWriteBytes.Set(writeBytes)
						} else {
							errorItems = append(errorItems, "writeBytes")
						}
					}
					if settings.Disk.UsageIO.ReadTime && service.IsExistValue(usageIO, "readTime") {
						if readTime, ok := usageIO["readTime"].(float64); ok {
							m.diskUsageIoReadTime.Set(readTime)
						} else {
							errorItems = append(errorItems, "readTime")
						}
					}
					if settings.Disk.UsageIO.WriteTime && service.IsExistValue(usageIO, "writeTime") {
						if writeTime, ok := usageIO["writeTime"].(float64); ok {
							m.diskUsageIoWriteTime.Set(writeTime)
						} else {
							errorItems = append(errorItems, "writeTime")
						}
					}
					if settings.Disk.UsageIO.ReadMergedCount && service.IsExistValue(usageIO, "readMergedCount") {
						if readMergedCount, ok := usageIO["readMergedCount"].(float64); ok {
							m.diskUsageIoReadMergedCount.Set(readMergedCount)
						} else {
							errorItems = append(errorItems, "readMergedCount")
						}
					}
					if settings.Disk.UsageIO.WriteMergedCount && service.IsExistValue(usageIO, "writeMergedCount") {
						if writeMergedCount, ok := usageIO["writeMergedCount"].(float64); ok {
							m.diskUsageIoWriteMergedCount.Set(writeMergedCount)
						} else {
							errorItems = append(errorItems, "writeMergedCount")
						}
					}
					if settings.Disk.UsageIO.BusyRate && service.IsExistValue(usageIO, "busyRate") {
						if busyRate, ok := usageIO["busyRate"].(float64); ok {
							m.diskUsageIoBusyRate.Set(busyRate)
						} else {
							errorItems = append(errorItems, "busyRate")
						}
					}
				} else {
					errorItems = append(errorItems, "usageIO")
				}
			}
		} else {
			errorItems = append(errorItems, "disk")
		}
	}

	if service.IsExistValue(hwOutput.HwMetrics, "metricEnergyJoules") {
		if metricEnergyJoules, ok := hwOutput.HwMetrics["metricEnergyJoules"].(map[string]any); ok {
			if settings.MetricEnergyJoules.SensorResetTime && service.IsExistValue(metricEnergyJoules, "sensorResetTime") {
				if metricEnergyJoulesSensorResetTime, ok := service.TimeStr2Float(metricEnergyJoules, "sensorResetTime"); ok {
					m.metricEnergyJoulesSensorResetTime.Set(metricEnergyJoulesSensorResetTime)
				} else {
					errorItems = append(errorItems, "sensorResetTime")
				}
			}
			if settings.MetricEnergyJoules.SensingInterval && service.IsExistValue(metricEnergyJoules, "sensingInterval") {
				if metricEnergyJoulesSensingInterval, ok := metricEnergyJoules["sensingInterval"].(float64); ok {
					m.metricEnergyJoulesSensingInterval.Set(metricEnergyJoulesSensingInterval)
				} else {
					errorItems = append(errorItems, "sensingInterval")
				}
			}
			if settings.MetricEnergyJoules.ReadingTime && service.IsExistValue(metricEnergyJoules, "readingTime") {
				if metricEnergyJoulesReadingTime, ok := service.TimeStr2Float(metricEnergyJoules, "readingTime"); ok {
					m.metricEnergyJoulesReadingTime.Set(metricEnergyJoulesReadingTime)
				} else {
					errorItems = append(errorItems, "readingTime")
				}
			}
			if settings.MetricEnergyJoules.Reading && service.IsExistValue(metricEnergyJoules, "reading") {
				if metricEnergyJoulesReading, ok := metricEnergyJoules["reading"].(float64); ok {
					m.metricEnergyJoulesReading.Set(metricEnergyJoulesReading)
				} else {
					errorItems = append(errorItems, "reading")
				}
			}
		} else {
			errorItems = append(errorItems, "metricEnergyJoules")
		}
	}

	if len(errorItems) > 0 {
		return service.CreateErroeMessage(strings.Join(errorItems, "/"))
	}
	return nil
}

// Read the storage configuration file and store it in a Struct
func loadStorageConfig(filepath string, settings *model.StorageConfig) error {
	buf, readErr := os.ReadFile(filepath)
	if readErr != nil {
		// If the file does not exist, treat all items as false and return nil.
		return nil
	}

	unmarshalErr := yaml.Unmarshal(buf, &settings)
	if unmarshalErr != nil {
		service.Log.Error(unmarshalErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0031", "Failed to unmarshal yaml.")
	}

	return nil
}

// Create prometheus metrics from the HW control metric information and storage metric acquisition settings, and return them to prometheus
func StorageMetricsHandler(context *gin.Context, hwOutput *model.HwOutput) {
	settings := model.StorageConfig{}

	readErr := loadStorageConfig(storageYamlFilePath, &settings)
	if readErr != nil {
		service.Log.Error(readErr.Error())
		context.JSON(service.GetStatusCode(readErr), service.ToJson(readErr))
		return
	}

	// Create a non-global registry.
	reg := prometheus.NewRegistry()
	// Create new metrics and register them using the custom registry.
	m, err := NewStorageMetrics(reg, hwOutput, &settings)
	if err != nil {
		service.Log.Error(err.Error())
		context.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}
	// Set values for the new created metrics.
	setErr := setStorageMetrics(m, hwOutput, &settings)
	if setErr != nil {
		service.Log.Error(setErr.Error())
		context.JSON(service.GetStatusCode(setErr), service.ToJson(setErr))
		return
	}

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(context.Writer, context.Request)
	service.Log.Info(context.Request.URL.Path + "[" + context.Request.Method + "] completed successfully.")
}
