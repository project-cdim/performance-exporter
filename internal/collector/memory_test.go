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
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/project-cdim/performance-exporter/internal/model"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

func setMemoryConfig(filepath string) model.MemoryConfig {
	yamlContent := model.MemoryConfig{}
	buf, _ := os.ReadFile(filepath)

	unmarshalErr := yaml.Unmarshal(buf, &yamlContent)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr.Error())
		os.Exit(1)
	}
	return yamlContent
}

func createMemoryMetrics(hwOutput model.HwOutput, yamlContent model.MemoryConfig) *memoryMetrics {
	reg := prometheus.NewRegistry()
	m, _ := NewMemoryMetrics(reg, &hwOutput, &yamlContent)
	return m
}

func createMemoryMetricsWithRegistry(hwOutput model.HwOutput, yamlContent model.MemoryConfig) (*memoryMetrics, *prometheus.Registry) {
	reg := prometheus.NewRegistry()
	m, _ := NewMemoryMetrics(reg, &hwOutput, &yamlContent)
	return m, reg
}

func setMemoryConfigAllValue(value bool) model.MemoryConfig {
	settings := model.MemoryConfig{}
	settings.Enabled = value
	settings.Status.State = value
	settings.Status.Health = value
	settings.PowerState = value
	settings.PowerCapability = value
	settings.DevicePortList.LTSSMState = value
	settings.CXLCapacity.Volatile = value
	settings.CXLCapacity.Persistent = value
	settings.CXLCapacity.Total = value
	settings.UsedMemory = value
	settings.Swap = value
	settings.MetricBandwidthPercent = value
	settings.MetricBlockSizeBytes = value
	settings.MetricOperatingSpeedMHz = value

	settings.MetricHealthData.DataLossDetected = value
	settings.MetricHealthData.LastShutdownSuccess = value
	settings.MetricHealthData.PerformanceDegraded = value
	settings.MetricHealthData.PredictedMediaLifeLeftPercent = value

	settings.MetricEnergyJoules.Reading = value
	settings.MetricEnergyJoules.ReadingTime = value
	settings.MetricEnergyJoules.SensingInterval = value
	settings.MetricEnergyJoules.SensorResetTime = value
	return settings
}

func TestNewMemoryMetrics(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	settings := setMemoryConfig(settingsPath + "normal.yaml")

	type args struct {
		reg      prometheus.Registerer
		hwOutput model.HwOutput
		settings *model.MemoryConfig
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
			"Error case: Return an error if the data type of CXLCapacity is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_CXLCapacity.json"),
				&settings,
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricHealthData is not map[string]any",
			args{
				prometheus.NewRegistry(),
				setHwMetric(hwInfoPath + "not_object_metricHealthData.json"),
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
			_, err := NewMemoryMetrics(tt.args.reg, &tt.args.hwOutput, tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMemoryMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setMemoryMetrics(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	type args struct {
		m        *memoryMetrics
		hwOutput model.HwOutput
		settings model.MemoryConfig
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
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: All values in the configuration file are false",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setMemoryConfig(settingsPath+"all_false.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setMemoryConfig(settingsPath + "all_false.yaml"),
			},
			false,
		},
		{
			"Normal case: When the configuration file is for status monitoring",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setMemoryConfig(settingsPath+"status_monitoring.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setMemoryConfig(settingsPath + "status_monitoring.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file is empty, treat all items as false",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setMemoryConfig(settingsPath+"empty.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setMemoryConfig(settingsPath + "empty.yaml"),
			},
			false,
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"normal.json"),
					setMemoryConfig(settingsPath+"aaa.yaml"),
				),
				setHwMetric(hwInfoPath + "normal.json"),
				setMemoryConfig(settingsPath + "aaa.yaml"),
			},
			false,
		},
		// Pattern of obtaining metric information for HW control. Confirm that panic does not occur
		{
			"Normal case: No metrics in the performance part of the HW control JSON",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"not_exist_metrics.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_exist_metrics.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The value of metrics in the performance part of the HW control JSON is null",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"is_null_metrics_value.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "is_null_metrics_value.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The liveness monitoring part of the HW control JSON. When prometheus side becomes 1",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"monitoring_1.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_1.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Normal case: The liveness monitoring part of the HW control JSON. When prometheus side becomes 0",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"monitoring_0.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "monitoring_0.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			false,
		},
		{
			"Error case: The data type of the metrics value in the HW control JSON is different from expected",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"unexpected_data_type_usagerate.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "unexpected_data_type_usagerate.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of status is not map[string]any",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"not_object_status.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_status.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of CXLCapacity is not map[string]any",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"not_object_CXLCapacity.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_CXLCapacity.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricHealthData is not map[string]any",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"not_object_metricHealthData.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_metricHealthData.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
		{
			"Error case: Return an error if the data type of metricEnergyJoules is not map[string]any",
			args{
				createMemoryMetrics(
					setHwMetric(hwInfoPath+"not_object_metricEnergyJoules.json"),
					setMemoryConfig(settingsPath+"normal.yaml"),
				),
				setHwMetric(hwInfoPath + "not_object_metricEnergyJoules.json"),
				setMemoryConfig(settingsPath + "normal.yaml"),
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setMemoryMetrics(tt.args.m, &tt.args.hwOutput, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("setMemoryMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadMemoryConfig(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	yamlContent := model.MemoryConfig{}

	type args struct {
		filepath string
		settings model.MemoryConfig
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected model.MemoryConfig
	}{
		{
			"normal",
			args{
				settingsPath + "normal.yaml",
				yamlContent,
			},
			false,
			setMemoryConfigAllValue(true),
		},
		{
			"Normal case: If the configuration file does not exist, treat all items as false",
			args{
				settingsPath + "aaa.yaml",
				yamlContent,
			},
			false,
			setMemoryConfigAllValue(false),
		},
		{
			"Error case: If the content of the configuration file is not in YAML format, return an error",
			args{
				settingsPath + "textonly.yaml",
				yamlContent,
			},
			true,
			setMemoryConfigAllValue(false),
		},
		{
			"Error case: If the value of an item in the configuration file is not a boolean, return an error",
			args{
				settingsPath + "invalidvalue_statusstate.yaml",
				yamlContent,
			},
			true,
			setMemoryConfigAllValue(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loadMemoryConfig(tt.args.filepath, &tt.args.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadMemoryConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.args.settings, tt.expected); diff != "" {
					t.Errorf("Value is mismatch :\n%s", diff)
				}
			}
		})
	}
}

func TestMemoryMetricsHandler(t *testing.T) {
	// Create gin context
	w := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/dummy", nil)
	ginContext.Request = req

	hwMetric := setHwMetric("testdata/memory/hw_info/normal.json")

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
			MemoryMetricsHandler(tt.args.context, tt.args.hwOutput)
		})
	}
}

func Test_setMemoryMetrics_Values(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	t.Run("LTSSMState L0 should set value 1", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "normal.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
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
			} else {
				fmt.Printf("Other metric: %s\n", gather.GetName())
			}
		}
	})
}

func Test_setMemoryMetrics_SwitchPortNumber(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	t.Run("multiple switchPortNumber normal values should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "normal.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
				// normal.json has multiple devicePortList entries for testing
				assert.Greater(t, len(gather.GetMetric()), 0, "Should have at least one metric")
				for _, metric := range gather.GetMetric() {
					labelPairs := metric.GetLabel()
					labels := make(map[string]string)
					for _, label := range labelPairs {
						labels[label.GetName()] = label.GetValue()
					}
					fabricId := labels["fabric_id"]
					switchId := labels["switch_id"]
					assert.NotEmpty(t, fabricId, "fabric_id label should not be empty")
					assert.NotEmpty(t, switchId, "switch_id label should not be empty")
					// switchPortNumber is optional, so not validated in this test
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 1.0, actualValue, "LTSSMState should be 1")
				}
			}
		}
	})

	t.Run("switchPortNumber empty string should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "switchport_empty.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
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
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
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

func Test_setMemoryMetrics_ErrorHandling(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	t.Run("switchID empty should skip processing and return error", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "switchid_empty.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.Error(t, err, "Should return error when switchID is empty")
		assert.Contains(t, err.Error(), "devicePortList_switchID", "Error message should contain devicePortList_switchID")

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics should be generated
		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics when switchID is empty")
			}
		}
	})

	t.Run("LTSSMState empty should skip processing and return error", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "ltssmstate_empty.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.Error(t, err, "Should return error when LTSSMState is empty")
		assert.Contains(t, err.Error(), "devicePortList_ltssmState", "Error message should contain devicePortList_ltssmState")

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics should be generated
		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics when LTSSMState is empty")
			}
		}
	})

	t.Run("fabricID empty should skip processing and return error", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "fabricid_empty.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.Error(t, err, "Should return error when fabricID is empty")
		assert.Contains(t, err.Error(), "devicePortList_fabricID", "Error message should contain devicePortList_fabricID")

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics should be generated
		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics when fabricID is empty")
			}
		}
	})
}

func Test_setMemoryMetrics_LTSSMStateValues(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	t.Run("LTSSMState non-L0 should set value 0", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "ltssmstate_l1.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 1, "Should have one metric")
				for _, metric := range gather.GetMetric() {
					actualValue := metric.GetGauge().GetValue()
					assert.Equal(t, 0.0, actualValue, "LTSSMState non-L0 should be 0")
				}
			}
		}
	})
}

func Test_setMemoryMetrics_BoundaryValues(t *testing.T) {
	const settingsPath = "testdata/memory/settings/"
	const hwInfoPath = "testdata/memory/hw_info/"

	t.Run("empty devicePortList should work", func(t *testing.T) {
		hwOutput := setHwMetric(hwInfoPath + "deviceportlist_empty.json")
		settings := setMemoryConfig(settingsPath + "normal.yaml")
		m, gatherer := createMemoryMetricsWithRegistry(hwOutput, settings)

		err := setMemoryMetrics(m, &hwOutput, &settings)
		assert.NoError(t, err)

		gathers, err := gatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		// No metrics are generated for empty array
		for _, gather := range gathers {
			if gather.GetName() == "memory_devicePortList_LTSSMState" {
				assert.Len(t, gather.GetMetric(), 0, "Should have no metrics for empty devicePortList")
			}
		}
	})
}
