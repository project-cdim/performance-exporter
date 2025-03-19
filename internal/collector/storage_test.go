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
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/project-cdim/performance-exporter/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

func setStorageConfig(filepath string) model.StorageConfig {
	yamlContent := model.StorageConfig{}
	buf, _ := os.ReadFile(filepath)

	unmarshalErr := yaml.Unmarshal(buf, &yamlContent)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr.Error())
		os.Exit(1)
	}
	return yamlContent
}

func createStorageMetrics(hwOutput model.HwOutput, yamlContent model.StorageConfig) *storageMetrics {
	reg := prometheus.NewRegistry()
	m, _ := NewStorageMetrics(reg, &hwOutput, &yamlContent)
	return m
}

func TestNewStorageMetrics(t *testing.T) {
	const settingsPath = "testdata/storage/settings/"
	const hwInfoPath = "testdata/storage/hw_info/"

	settings := setStorageConfig(settingsPath + "normal.yaml")

	type args struct {
		reg      prometheus.Registerer
		hwOutput model.HwOutput
		settings *model.StorageConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"nomal",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "normal.json"),
				&settings,
			},
			false,
		},
		{
			"Normal case: When there are no metrics",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_exist_metrics.json"),
				&settings,
			},
			false,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity_data is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity_data.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity_metadata is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity_metadata.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity_snapshot is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity_snapshot.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of status is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_status.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricEnergyJoules is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_metricEnergyJoules.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of disk is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_disk.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of disk's usageIO is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_usageIO.json"),
				&settings,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStorageMetrics(tt.args.reg, &tt.args.hwOutput, tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorageMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setStorageMetrics(t *testing.T) {
	const settingsPath = "testdata/storage/settings/"
	const hwInfoPath = "testdata/storage/hw_info/"

	type args struct {
		m        *storageMetrics
		hwOutput model.HwOutput
		settings model.StorageConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Pattern of configuration files. Confirm that panic does not occur
		{
			"nomal",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: All values in the configuration file are false",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setStorageConfig(settingsPath+"all_false.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setStorageConfig(settingsPath + "all_false.yaml"),
			},
			false,
		},
		{
			"Normal case: The configuration file is for liveness monitoring",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setStorageConfig(settingsPath+"status_monitoring.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setStorageConfig(settingsPath + "status_monitoring.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file is empty, treat all items as false",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setStorageConfig(settingsPath+"empty.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setStorageConfig(settingsPath + "empty.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setStorageConfig(settingsPath+"aaa.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setStorageConfig(settingsPath + "aaa.yaml"),
			},
			false,
		},
		// Pattern of obtaining HW control metric information. Confirm that panic does not occur
		{
			"Normal case: HW control JSON does not have volumeCapacity_data",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_exist_volumeCapacity_data.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_volumeCapacity_data.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: HW control JSON does not have volumeCapacity_metadata",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_exist_volumeCapacity_metadata.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_volumeCapacity_metadata.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: HW control JSON does not have volumeCapacity_snapshot",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_exist_volumeCapacity_snapshot.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_volumeCapacity_snapshot.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: HW control JSON does not have performance metrics",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_exist_metrics.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_metrics.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The value of performance metrics in HW control JSON is null",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"is_null_metrics_value.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "is_null_metrics_value.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: Liveness monitoring part of HW control JSON. When prometheus side is 1",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"monitoring_1.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_1.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: Liveness monitoring part of HW control JSON. When prometheus side is 0",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"monitoring_0.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_0.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Error case: The data type of the metrics value in HW control JSON is different from expected",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"unexpected_data_type_usagerate.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "unexpected_data_type_usagerate.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_volumeCapacity.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity_data in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_volumeCapacity_data.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity_data.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity_metadata in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_volumeCapacity_metadata.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity_metadata.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of volumeCapacity_snapshot in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_volumeCapacity_snapshot.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_volumeCapacity_snapshot.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of status in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_status.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_status.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricEnergyJoules in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_metricEnergyJoules.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_metricEnergyJoules.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of disk in HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_disk.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_disk.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of usageIO in the disk of HW control JSON is not map[string]any",
			args{
				createStorageMetrics(
					setHwMetric(hwInfoPath+"not_object_usageIO.json"),
					setStorageConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_usageIO.json"),
				setStorageConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setStorageMetrics(tt.args.m, &tt.args.hwOutput, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("setStorageMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadStorageConfig(t *testing.T) {
	const settingsPath = "testdata/storage/settings/"
	yamlContent := model.StorageConfig{}

	type args struct {
		filepath string
		settings model.StorageConfig
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected model.StorageConfig
	}{
		{
			"nomal",
			args{
				settingsPath + "normal.yaml",
				yamlContent,
			},
			false,
			setStorageConfigAllValue(true),
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				settingsPath + "aaa.yaml",
				yamlContent,
			},
			false,
			setStorageConfigAllValue(false),
		},
		{
			"Error case: Return an error if the content of the configuration file is not in YAML format",
			args{
				settingsPath + "textonly.yaml",
				yamlContent,
			},
			true,
			setStorageConfigAllValue(false),
		},
		{
			"Error case: Return an error if the value of an item in the configuration file is not a boolean",
			args{
				settingsPath + "invalidvalue_statusstate.yaml",
				yamlContent,
			},
			true,
			setStorageConfigAllValue(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loadStorageConfig(tt.args.filepath, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadStorageConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.args.settings, tt.expected); diff != "" {
					t.Errorf("Value is mismatch :\n%s", diff)
				}
			}
		})
	}
}

func TestStorageMetricsHandler(t *testing.T) {
	// Generate gin context
	w := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/dummy", nil)
	ginContext.Request = req

	hwMetric := setHwMetric("testdata/storage/hw_info/normal.json")

	type args struct {
		context  *gin.Context
		hwOutput *model.HwOutput
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"nomal",
			args{
				ginContext,
				&hwMetric,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StorageMetricsHandler(tt.args.context, tt.args.hwOutput)
		})
	}
}

func setStorageConfigAllValue(value bool) model.StorageConfig {
	settings := model.StorageConfig{}
	settings.VolumeCapacity.Data.AllocatedBytes = value
	settings.VolumeCapacity.Data.ConsumedBytes = value
	settings.VolumeCapacity.Data.GuaranteedBytes = value
	settings.VolumeCapacity.Data.ProvisionedBytes = value

	settings.VolumeCapacity.Metadata.AllocatedBytes = value
	settings.VolumeCapacity.Metadata.ConsumedBytes = value
	settings.VolumeCapacity.Metadata.GuaranteedBytes = value
	settings.VolumeCapacity.Metadata.ProvisionedBytes = value

	settings.VolumeCapacity.Snapshot.AllocatedBytes = value
	settings.VolumeCapacity.Snapshot.ConsumedBytes = value
	settings.VolumeCapacity.Snapshot.GuaranteedBytes = value
	settings.VolumeCapacity.Snapshot.ProvisionedBytes = value

	settings.VolumeRemainingCapacityPercent = value
	settings.DriveNegotiatedSpeedGbs = value

	settings.Status.State = value
	settings.Status.Health = value
	settings.PowerState = value
	settings.PowerCapability = value
	settings.LtssmState = value

	settings.Disk.AmountUsedDisk = value
	settings.Disk.UsageIO.ReadCount = value
	settings.Disk.UsageIO.WriteCount = value
	settings.Disk.UsageIO.ReadBytes = value
	settings.Disk.UsageIO.WriteBytes = value
	settings.Disk.UsageIO.ReadTime = value
	settings.Disk.UsageIO.WriteTime = value
	settings.Disk.UsageIO.ReadMergedCount = value
	settings.Disk.UsageIO.WriteMergedCount = value
	settings.Disk.UsageIO.BusyRate = value

	settings.MetricEnergyJoules.Reading = value
	settings.MetricEnergyJoules.ReadingTime = value
	settings.MetricEnergyJoules.SensingInterval = value
	settings.MetricEnergyJoules.SensorResetTime = value
	return settings
}
