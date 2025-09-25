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

func setNetworkInterfaceConfig(filepath string) model.NetworkInterfaceConfig {
	yamlContent := model.NetworkInterfaceConfig{}
	buf, _ := os.ReadFile(filepath)

	unmarshalErr := yaml.Unmarshal(buf, &yamlContent)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr.Error())
		os.Exit(1)
	}
	return yamlContent
}

func createNetworkInterfaceMetrics(hwOutput model.HwOutput, yamlContent model.NetworkInterfaceConfig) *networkInterfaceMetrics {
	reg := prometheus.NewRegistry()
	m, _ := NewNetworkInterfaceMetrics(reg, &hwOutput, &yamlContent)
	return m
}

func TestNewNetworkInterfaceMetrics(t *testing.T) {
	const settingsPath = "testdata/networkInterface/settings/"
	const hwInfoPath = "testdata/networkInterface/hw_info/"

	settings := setNetworkInterfaceConfig(settingsPath + "normal.yaml")

	type args struct {
		reg      prometheus.Registerer
		hwOutput model.HwOutput
		settings *model.NetworkInterfaceConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"normal",
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
			"Error case: Return an error if the data type of networkInterfaceInformation is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_networkInterfaceInformation.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of networkTraffic is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_networkTraffic.json"),
				&settings,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNetworkInterfaceMetrics(tt.args.reg, &tt.args.hwOutput, tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNetworkInterfaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setNetworkInterfaceMetrics(t *testing.T) {
	const settingsPath = "testdata/networkInterface/settings/"
	const hwInfoPath = "testdata/networkInterface/hw_info/"

	type args struct {
		m        *networkInterfaceMetrics
		hwOutput model.HwOutput
		settings model.NetworkInterfaceConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Pattern of configuration files. Confirm that panic does not occur
		{
			"normal",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: All values in the configuration file are false",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setNetworkInterfaceConfig(settingsPath+"all_false.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setNetworkInterfaceConfig(settingsPath + "all_false.yaml"),
			},
			false,
		},
		{
			"Normal case: When the configuration file is for status monitoring",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setNetworkInterfaceConfig(settingsPath+"status_monitoring.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setNetworkInterfaceConfig(settingsPath + "status_monitoring.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file is empty, treat all items as false",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setNetworkInterfaceConfig(settingsPath+"empty.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setNetworkInterfaceConfig(settingsPath + "empty.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setNetworkInterfaceConfig(settingsPath+"aaa.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setNetworkInterfaceConfig(settingsPath + "aaa.yaml"),
			},
			false,
		},
		// Pattern of obtaining metric information for HW control. Confirm that panic does not occur
		{
			"Normal case: No metrics in the performance part of the HW control JSON",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"not_exist_metrics.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_metrics.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The value of metrics in the performance part of the HW control JSON is null",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"is_null_metrics_value.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "is_null_metrics_value.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The liveness monitoring part of the HW control JSON. When prometheus side becomes 1",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"monitoring_1.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_1.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The liveness monitoring part of the HW control JSON. When prometheus side becomes 0",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"monitoring_0.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_0.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Error case: The data type of the metrics value in the HW control JSON is different from expected",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"unexpected_data_type_usagerate.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "unexpected_data_type_usagerate.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of status is not map[string]any",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"not_object_status.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_status.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricEnergyJoules is not map[string]any",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"not_object_metricEnergyJoules.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_metricEnergyJoules.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of networkInterfaceInformation is not map[string]any",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"not_object_networkInterfaceInformation.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_networkInterfaceInformation.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of networkTraffic is not map[string]any",
			args{
				createNetworkInterfaceMetrics(
					setHwMetric(hwInfoPath+"not_object_networkTraffic.json"),
					setNetworkInterfaceConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_networkTraffic.json"),
				setNetworkInterfaceConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setNetworkInterfaceMetrics(tt.args.m, &tt.args.hwOutput, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("setNetworkInterfaceMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadNetworkInterfaceConfig(t *testing.T) {
	const settingsPath = "testdata/networkInterface/settings/"
	yamlContent := model.NetworkInterfaceConfig{}

	type args struct {
		filepath string
		settings model.NetworkInterfaceConfig
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected model.NetworkInterfaceConfig
	}{
		{
			"normal",
			args{
				settingsPath + "normal.yaml",
				yamlContent,
			},
			false,
			setNetworkInterfaceConfigAllValue(true),
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				settingsPath + "aaa.yaml",
				yamlContent,
			},
			false,
			setNetworkInterfaceConfigAllValue(false),
		},
		{
			"Error case: Return an error if the content of the configuration file is not in YAML format",
			args{
				settingsPath + "textonly.yaml",
				yamlContent,
			},
			true,
			setNetworkInterfaceConfigAllValue(false),
		},
		{
			"Error case: Return an error if the value of an item in the configuration file is not a boolean",
			args{
				settingsPath + "invalidvalue_statusstate.yaml",
				yamlContent,
			},
			true,
			setNetworkInterfaceConfigAllValue(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loadNetworkInterfaceConfig(tt.args.filepath, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadNetworkInterfaceConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.args.settings, tt.expected); diff != "" {
					t.Errorf("Value is mismatch :\n%s", diff)
				}
			}
		})
	}
}

func TestNetworkInterfaceMetricsHandler(t *testing.T) {
	// Generate gin context
	w := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/dummy", nil)
	ginContext.Request = req

	hwMetric := setHwMetric("testdata/networkInterface/hw_info/normal.json")

	type args struct {
		context  *gin.Context
		hwOutput *model.HwOutput
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"normal",
			args{
				ginContext,
				&hwMetric,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NetworkInterfaceMetricsHandler(tt.args.context, tt.args.hwOutput)
		})
	}
}

func setNetworkInterfaceConfigAllValue(value bool) model.NetworkInterfaceConfig {
	settings := model.NetworkInterfaceConfig{}
	settings.DeviceEnabled = value
	settings.Status.State = value
	settings.Status.Health = value
	settings.PowerState = value
	settings.PowerCapability = value
	settings.LtssmState = value

	settings.NetworkInterfaceInformation.NetworkSpeed = value
	settings.NetworkInterfaceInformation.NetworkTraffic.ReceivePackets = value
	settings.NetworkInterfaceInformation.NetworkTraffic.TransmitPackets = value
	settings.NetworkInterfaceInformation.NetworkTraffic.BytesSent = value
	settings.NetworkInterfaceInformation.NetworkTraffic.BytesRecv = value
	settings.NetworkInterfaceInformation.NetworkTraffic.Errin = value
	settings.NetworkInterfaceInformation.NetworkTraffic.Errout = value
	settings.NetworkInterfaceInformation.NetworkTraffic.Dropin = value
	settings.NetworkInterfaceInformation.NetworkTraffic.Dropout = value

	settings.MetricsCPUCorePercent = value
	settings.MetricsHostBusRXPercent = value
	settings.MetricsHostBusTXPercent = value
	settings.MetricsRXAvgQueueDepthPercent = value
	settings.MetricsTXAvgQueueDepthPercent = value
	settings.MetricsRXBytes = value
	settings.MetricsRXFrames = value
	settings.MetricsTXBytes = value
	settings.MetricsTXFrames = value

	settings.MetricEnergyJoules.Reading = value
	settings.MetricEnergyJoules.ReadingTime = value
	settings.MetricEnergyJoules.SensingInterval = value
	settings.MetricEnergyJoules.SensorResetTime = value
	return settings
}
