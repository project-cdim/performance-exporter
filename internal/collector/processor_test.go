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

package collector

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/project-cdim/performance-exporter/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func setHwMetric(filepath string) model.HwOutput {
	hwOutput := model.HwOutput{}
	buf, readErr := os.ReadFile(filepath)
	if readErr != nil {
		fmt.Println(readErr.Error())
		os.Exit(1)
	}
	unmarshalErr := json.Unmarshal(buf, &hwOutput.HwMetrics)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr.Error())
		os.Exit(1)
	}
	return hwOutput
}

func setProcessorConfig(filepath string) model.ProcessorConfig {
	yamlContent := model.ProcessorConfig{}
	buf, _ := os.ReadFile(filepath)

	unmarshalErr := yaml.Unmarshal(buf, &yamlContent)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr.Error())
		os.Exit(1)
	}
	return yamlContent
}

func createProcessorMetrics(hwOutput model.HwOutput, yamlContent model.ProcessorConfig) *processorMetrics {
	reg := prometheus.NewRegistry()
	m, _ := NewProcessorMetrics(reg, &hwOutput, &yamlContent)
	return m
}

func createProcessorMetricsWithRegistry(hwOutput model.HwOutput, yamlContent model.ProcessorConfig) (*processorMetrics, *prometheus.Registry) {
	reg := prometheus.NewRegistry()
	m, _ := NewProcessorMetrics(reg, &hwOutput, &yamlContent)
	return m, reg
}

func setProcessorConfigAllValue(value bool) model.ProcessorConfig {
	settings := model.ProcessorConfig{}
	settings.Status.State = value
	settings.Status.Health = value
	settings.PowerState = value
	settings.PowerCapability = value
	settings.DevicePortList.LTSSMState = value
	settings.UsageRate = value
	settings.User = value
	settings.System = value
	settings.Wait = value
	settings.Idle = value
	settings.CPUWatts = value
	settings.MetricBandwidthPercent = value
	settings.MetricOperatingSpeedMHz = value
	settings.MetricLocalMemoryBandwidthBytes = value
	settings.MetricRemoteMemoryBandwidthBytes = value
	settings.MetricEnergyJoules.Reading = value
	settings.MetricEnergyJoules.ReadingTime = value
	settings.MetricEnergyJoules.SensingInterval = value
	settings.MetricEnergyJoules.SensorResetTime = value
	return settings
}

func TestNewProcessorMetrics(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	settings := setProcessorConfig(settingsPath + "normal.yaml")

	type args struct {
		reg      prometheus.Registerer
		hwOutput model.HwOutput
		settings *model.ProcessorConfig
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProcessorMetrics(tt.args.reg, &tt.args.hwOutput, tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProcessorMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setProcessorMetrics(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	type args struct {
		m        *processorMetrics
		hwOutput model.HwOutput
		settings model.ProcessorConfig
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
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: All values in the configuration file are false",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setProcessorConfig(settingsPath+"all_false.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setProcessorConfig(settingsPath + "all_false.yaml"),
			},
			false,
		},
		{
			"Normal case: When the configuration file is for status monitoring",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setProcessorConfig(settingsPath+"status_monitoring.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setProcessorConfig(settingsPath + "status_monitoring.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file is empty, treat all items as false",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setProcessorConfig(settingsPath+"empty.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setProcessorConfig(settingsPath + "empty.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setProcessorConfig(settingsPath+"aaa.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setProcessorConfig(settingsPath + "aaa.yaml"),
			},
			false,
		},
		// Pattern of obtaining metric information for HW control. Confirm that panic does not occur
		{
			"Normal case: No metrics in the performance part of the HW control JSON",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"not_exist_metrics.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_metrics.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The value of metrics in the performance part of the HW control JSON is null",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"is_null_metrics_value.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "is_null_metrics_value.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The liveness monitoring part of the HW control JSON. When prometheus side becomes 1",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"monitoring_1.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_1.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The liveness monitoring part of the HW control JSON. When prometheus side becomes 0",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"monitoring_0.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_0.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Error case: Return an error if the data type of metrics value in the HW control JSON is different from expected",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"unexpected_data_type_usagerate.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "unexpected_data_type_usagerate.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of status is not map[string]any",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"not_object_metricEnergyJoules.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_metricEnergyJoules.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricEnergyJoules is not map[string]any",
			args{
				createProcessorMetrics(
					setHwMetric(hwInfoPath+"not_object_metricEnergyJoules.json"),
					setProcessorConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_metricEnergyJoules.json"),
				setProcessorConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setProcessorMetrics(tt.args.m, &tt.args.hwOutput, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("setProcessorMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadProcessorConfig(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	processorConfig := model.ProcessorConfig{}

	type args struct {
		filepath string
		settings model.ProcessorConfig
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected model.ProcessorConfig
	}{
		{
			"normal",
			args{
				settingsPath + "normal.yaml",
				processorConfig,
			},
			false,
			setProcessorConfigAllValue(true),
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				settingsPath + "aaa.yaml",
				processorConfig,
			},
			false,
			setProcessorConfigAllValue(false),
		},
		{
			"Error case: Return an error if the content of the configuration file is not in YAML format",
			args{
				settingsPath + "textonly.yaml",
				processorConfig,
			},
			true,
			setProcessorConfigAllValue(false),
		},
		{
			"Error case: Return an error if the value of an item in the configuration file is not a boolean",
			args{
				settingsPath + "invalidvalue_statusstate.yaml",
				processorConfig,
			},
			true,
			setProcessorConfigAllValue(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loadProcessorConfig(tt.args.filepath, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadProcessorConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.args.settings, tt.expected); diff != "" {
					t.Errorf("Value is mismatch :\n%s", diff)
				}
			}
		})
	}
}

func TestProcessorMetricsHandler(t *testing.T) {
	// Generate gin context
	w := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/dummy", nil)
	ginContext.Request = req

	hwMetric := setHwMetric("testdata/processor/hw_info/normal.json")

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
			ProcessorMetricsHandler(tt.args.context, tt.args.hwOutput)
		})
	}
}

func Test_setProcessorMetrics_Values(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	t.Run("LTSSMState L0 should set value 1", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "normal.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				for _, metric := range gather.GetMetric() {
					labelPairs := metric.GetLabel()
					labels := make(map[string]string)
					for _, label := range labelPairs {
						labels[label.GetName()] = label.GetValue()
					}
					fabricId := labels["fabric_id"]
					switchId := labels["switch_id"]
					// switchPortNumber is optional, so not validated in this test
					assert.NotEmpty(t, fabricId, "fabric_id label should not be empty")
					assert.NotEmpty(t, switchId, "switch_id label should not be empty")
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 1.0, actualValue, "LTSSMState should be 1")
				}
			}
		}
	})
}

func Test_setProcessorMetrics_SwitchPortNumber(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	t.Run("multiple switchPortNumber normal values should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "normal.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 2, "Should have two metrics")
				for _, metric := range gather.GetMetric() {
					labelPairs := metric.GetLabel()
					labels := make(map[string]string)
					for _, label := range labelPairs {
						labels[label.GetName()] = label.GetValue()
					}
					fabricId := labels["fabric_id"]
					switchId := labels["switch_id"]
					switchPortNumber := labels["switch_port_number"]
					assert.NotEmpty(t, fabricId, "fabric_id label should not be empty")
					assert.NotEmpty(t, switchId, "switch_id label should not be empty")
					assert.NotEmpty(t, switchPortNumber, "switch_port_number label should not be empty")
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 1.0, actualValue, "LTSSMState should be 1")
				}
			}
		}
	})

	t.Run("switchPortNumber empty string should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "switchport_empty.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 1, "Should have one metric")
				for _, metric := range gather.GetMetric() {
					labelPairs := metric.GetLabel()
					labels := make(map[string]string)
					for _, label := range labelPairs {
						labels[label.GetName()] = label.GetValue()
					}
					fabricId := labels["fabric_id"]
					switchId := labels["switch_id"]
					switchPortNumber := labels["switch_port_number"]
					assert.Equal(t, "1", fabricId, "fabric_id should be 1")
					assert.Equal(t, "fmsw-01", switchId, "switch_id should be fmsw-01")
					assert.Equal(t, "", switchPortNumber, "switch_port_number should be empty string")
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 1.0, actualValue, "LTSSMState should be 1")
				}
			}
		}
	})

	t.Run("switchPortNumber missing field should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "switchport_missing.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 1, "Should have one metric")
				for _, metric := range gather.GetMetric() {
					labelPairs := metric.GetLabel()
					labels := make(map[string]string)
					for _, label := range labelPairs {
						labels[label.GetName()] = label.GetValue()
					}
					fabricId := labels["fabric_id"]
					switchId := labels["switch_id"]
					switchPortNumber := labels["switch_port_number"]
					assert.Equal(t, "1", fabricId, "fabric_id should be 1")
					assert.Equal(t, "fmsw-01", switchId, "switch_id should be fmsw-01")
					assert.Equal(t, "", switchPortNumber, "switch_port_number should be empty string when field is missing")
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 1.0, actualValue, "LTSSMState should be 1")
				}
			}
		}
	})
}

func Test_setProcessorMetrics_ErrorHandling(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	t.Run("switchID empty should skip processing and return error", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "switchid_empty.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.Error(t, err, "Should return error when switchID is empty")
		assert.Contains(t, err.Error(), "devicePortList_switchID", "Error message should contain devicePortList_switchID")

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics should be generated
		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics when switchID is empty")
			}
		}
	})

	t.Run("LTSSMState empty should skip processing and return error", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "ltssmstate_empty.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.Error(t, err, "Should return error when LTSSMState is empty")
		assert.Contains(t, err.Error(), "devicePortList_ltssmState", "Error message should contain devicePortList_ltssmState")

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics should be generated
		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics when LTSSMState is empty")
			}
		}
	})

	t.Run("fabricID empty should skip processing and return error", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "fabricid_empty.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.Error(t, err, "Should return error when fabricID is empty")
		assert.Contains(t, err.Error(), "devicePortList_fabricID", "Error message should contain devicePortList_fabricID")

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics should be generated
		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics when fabricID is empty")
			}
		}
	})
}

func Test_setProcessorMetrics_LTSSMStateValues(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	t.Run("LTSSMState non-L0 should set value 0", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "ltssmstate_l1.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 1, "Should have one metric")
				for _, metric := range gather.GetMetric() {
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 0.0, actualValue, "LTSSMState non-L0 should be 0")
				}
			}
		}
	})
}

func Test_setProcessorMetrics_BoundaryValues(t *testing.T) {
	const settingsPath = "testdata/processor/settings/"
	const hwInfoPath = "testdata/processor/hw_info/"

	t.Run("empty devicePortList should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "deviceportlist_empty.json")
		settings := setProcessorConfig(settingsPath + "normal.yaml")
		m, gatherer := createProcessorMetricsWithRegistry(hwOutput, settings)

		err := setProcessorMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics are generated for empty array
		for _, gather := range gathers {
			if gather.GetName() == "CPU_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics for empty devicePortList")
			}
		}
	})
}
