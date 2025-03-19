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
	enabledTrueValue              float64 = 1
	enabledFalseValue             float64 = 0
	dataLossDetectedTrueValue     float64 = 1
	dataLossDetectedFalseValue    float64 = 0
	lastShutdownSuccessTrueValue  float64 = 1
	lastShutdownSuccessFalseValue float64 = 0
	performanceDegradedTrueValue  float64 = 1
	performanceDegradedFalseValue float64 = 0
	memoryYamlFilePath            string  = "configs/memory.yaml"
)

// Definition of Metrics
type memoryMetrics struct {
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	enabledTrue         prometheus.Gauge
	statusStateEnabled  prometheus.Gauge
	statusHealthOk      prometheus.Gauge
	powerStateOn        prometheus.Gauge
	powerCapabilityTrue prometheus.Gauge
	ltssmStateL0        prometheus.Gauge

	cxlCapacityVolatile     prometheus.Gauge
	cxlCapacityPersistent   prometheus.Gauge
	cxlCapacityTotal        prometheus.Gauge
	usedMemory              prometheus.Gauge
	swap                    prometheus.Gauge
	metricBandwidthPercent  prometheus.Gauge
	metricBlockSizeBytes    prometheus.Gauge
	metricOperatingSpeedMHz prometheus.Gauge

	metricHealthDataDataLossDetectedTrue          prometheus.Gauge
	metricHealthDataLastShutdownSuccessTrue       prometheus.Gauge
	metricHealthDataPerformanceDegradedTrue       prometheus.Gauge
	metricHealthDataPredictedMediaLifeLeftPercent prometheus.Gauge

	metricEnergyJoulesSensorResetTime prometheus.Gauge
	metricEnergyJoulesSensingInterval prometheus.Gauge
	metricEnergyJoulesReadingTime     prometheus.Gauge
	metricEnergyJoulesReading         prometheus.Gauge
}

// Define memory metrics.
// Descriptor of metrics "One of the pieces of information embedded in metrics (name, information to be placed on #HELP), and non-numeric information to be displayed later in the graph"
func NewMemoryMetrics(reg prometheus.Registerer, hwOutput *model.HwOutput, settings *model.MemoryConfig) (*memoryMetrics, error) {
	m := &memoryMetrics{
		// deviceID is not registered as a metric because it uses the job name
		// type is not registered as a metric because it uses the namespace
		// attribute is not registered as a metric because it is a fixed value

		enabledTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "enabled",
			Help:        "Whether memory is enabled",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		statusStateEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "status_state",
			Help:        "Memory status information. Resource status",
			ConstLabels: prometheus.Labels{"value": "Enabled"},
		}),
		statusHealthOk: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "status_health",
			Help:        "Memory status information. Resource health status",
			ConstLabels: prometheus.Labels{"value": "OK"},
		}),
		powerStateOn: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "powerState",
			Help:        "Current power state of the memory.",
			ConstLabels: prometheus.Labels{"value": "On"},
		}),
		powerCapabilityTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "powerCapability",
			Help:        "Whether power control function is enabled.",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		ltssmStateL0: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "LTSSMState",
			Help:        "CXL device link state",
			ConstLabels: prometheus.Labels{"value": "L0"},
		}),

		cxlCapacityVolatile: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "CXLCapacity_volatile",
			Help:      "CXL memory volatile capacity (byte)",
		}),
		cxlCapacityPersistent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "CXLCapacity_persistent",
			Help:      "CXL memory non-volatile capacity (byte)",
		}),
		cxlCapacityTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "CXLCapacity_total",
			Help:      "CXL memory total capacity (byte)",
		}),
		usedMemory: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "usedMemory",
			Help:      "Used memory (kilobyte)",
		}),
		swap: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "swap",
			Help:      "Swap usage (kilobyte)",
		}),
		metricBandwidthPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricBandwidthPercent",
			Help:      "Memory bandwidth usage (%)",
		}),
		metricBlockSizeBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricBlockSizeBytes",
			Help:      "Block size in bytes",
		}),
		metricOperatingSpeedMHz: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricOperatingSpeedMHz",
			Help:      "Memory operating speed in MHz or MT/s as needed",
		}),

		metricHealthDataDataLossDetectedTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "metricHealthData_dataLossDetected",
			Help:        "Memory health information. Whether data loss has been detected",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		metricHealthDataLastShutdownSuccessTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "metricHealthData_lastShutdownSuccess",
			Help:        "Memory health information. Whether the last shutdown was successful",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		metricHealthDataPerformanceDegradedTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   model.Memory,
			Name:        "metricHealthData_performanceDegraded",
			Help:        "Memory health information. Whether performance has degraded",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		metricHealthDataPredictedMediaLifeLeftPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricHealthData_predictedMediaLifeLeftPercent",
			Help:      "Memory health information. The percentage of media expected to be available for read/write",
		}),

		metricEnergyJoulesSensorResetTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricEnergyJoules_sensorResetTime",
			Help:      "The date and time when the property was last reset. (UTC)",
		}),
		metricEnergyJoulesSensingInterval: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricEnergyJoules_sensingInterval",
			Help:      "Time interval between sensor readings (seconds)",
		}),
		metricEnergyJoulesReadingTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricEnergyJoules_readingTime",
			Help:      "The date and time when the measurement was obtained from the sensor. (UTC)",
		}),
		metricEnergyJoulesReading: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: model.Memory,
			Name:      "metricEnergyJoules_reading",
			Help:      "Measurement of energy consumption (J)",
		}),
	}

	errorItems := []string{}
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	if settings.Enabled && service.IsExistValue(hwOutput.HwMetrics, "enabled") {
		reg.MustRegister(m.enabledTrue)
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

	if service.IsExistValue(hwOutput.HwMetrics, "CXLCapacity") {
		if cxlCapacity, ok := hwOutput.HwMetrics["CXLCapacity"].(map[string]any); ok {
			if settings.CXLCapacity.Volatile && service.IsExistValue(cxlCapacity, "volatile") {
				reg.MustRegister(m.cxlCapacityVolatile)
			}
			if settings.CXLCapacity.Persistent && service.IsExistValue(cxlCapacity, "persistent") {
				reg.MustRegister(m.cxlCapacityPersistent)
			}
			if settings.CXLCapacity.Total && service.IsExistValue(cxlCapacity, "total") {
				reg.MustRegister(m.cxlCapacityTotal)
			}
		} else {
			errorItems = append(errorItems, "CXLCapacity")
		}
	}
	if settings.UsedMemory && service.IsExistValue(hwOutput.HwMetrics, "usedMemory") {
		reg.MustRegister(m.usedMemory)
	}
	if settings.Swap && service.IsExistValue(hwOutput.HwMetrics, "swap") {
		reg.MustRegister(m.swap)
	}
	if settings.MetricBandwidthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricBandwidthPercent") {
		reg.MustRegister(m.metricBandwidthPercent)
	}
	if settings.MetricBlockSizeBytes && service.IsExistValue(hwOutput.HwMetrics, "metricBlockSizeBytes") {
		reg.MustRegister(m.metricBlockSizeBytes)
	}
	if settings.MetricOperatingSpeedMHz && service.IsExistValue(hwOutput.HwMetrics, "metricOperatingSpeedMHz") {
		reg.MustRegister(m.metricOperatingSpeedMHz)
	}

	if service.IsExistValue(hwOutput.HwMetrics, "metricHealthData") {
		if metricHealthData, ok := hwOutput.HwMetrics["metricHealthData"].(map[string]any); ok {
			if settings.MetricHealthData.DataLossDetected && service.IsExistValue(metricHealthData, "dataLossDetected") {
				reg.MustRegister(m.metricHealthDataDataLossDetectedTrue)
			}
			if settings.MetricHealthData.LastShutdownSuccess && service.IsExistValue(metricHealthData, "lastShutdownSuccess") {
				reg.MustRegister(m.metricHealthDataLastShutdownSuccessTrue)
			}
			if settings.MetricHealthData.PerformanceDegraded && service.IsExistValue(metricHealthData, "performanceDegraded") {
				reg.MustRegister(m.metricHealthDataPerformanceDegradedTrue)
			}
			if settings.MetricHealthData.PredictedMediaLifeLeftPercent && service.IsExistValue(metricHealthData, "predictedMediaLifeLeftPercent") {
				reg.MustRegister(m.metricHealthDataPredictedMediaLifeLeftPercent)
			}
		} else {
			errorItems = append(errorItems, "metricHealthData")
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

// Set the memory metrics value.
func setMemoryMetrics(m *memoryMetrics, hwOutput *model.HwOutput, settings *model.MemoryConfig) error {
	errorItems := []string{}

	// The attribute is a fixed value, so it is not registered.

	if settings.Enabled && service.IsExistValue(hwOutput.HwMetrics, "enabled") {
		if enabled, ok := hwOutput.HwMetrics["enabled"].(bool); ok {
			enabledValue := enabledFalseValue
			if enabled {
				enabledValue = enabledTrueValue
			}
			m.enabledTrue.Set(enabledValue)
		} else {
			errorItems = append(errorItems, "enabled")
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

	if service.IsExistValue(hwOutput.HwMetrics, "CXLCapacity") {
		if cxlCapacity, ok := hwOutput.HwMetrics["CXLCapacity"].(map[string]any); ok {
			if settings.CXLCapacity.Volatile && service.IsExistValue(cxlCapacity, "volatile") {
				if volatile, ok := cxlCapacity["volatile"].(float64); ok {
					m.cxlCapacityVolatile.Set(volatile)
				} else {
					errorItems = append(errorItems, "CXLCapacity_volatile")
				}
			}
			if settings.CXLCapacity.Persistent && service.IsExistValue(cxlCapacity, "persistent") {
				if persistent, ok := cxlCapacity["persistent"].(float64); ok {
					m.cxlCapacityPersistent.Set(persistent)
				} else {
					errorItems = append(errorItems, "CXLCapacity_persistent")
				}
			}
			if settings.CXLCapacity.Total && service.IsExistValue(cxlCapacity, "total") {
				if total, ok := cxlCapacity["total"].(float64); ok {
					m.cxlCapacityTotal.Set(total)
				} else {
					errorItems = append(errorItems, "CXLCapacity_total")
				}
			}
		} else {
			errorItems = append(errorItems, "CXLCapacity")
		}
	}
	if settings.UsedMemory && service.IsExistValue(hwOutput.HwMetrics, "usedMemory") {
		if usedMemory, ok := hwOutput.HwMetrics["usedMemory"].(float64); ok {
			m.usedMemory.Set(usedMemory)
		} else {
			errorItems = append(errorItems, "usedMemory")
		}
	}
	if settings.Swap && service.IsExistValue(hwOutput.HwMetrics, "swap") {
		if swap, ok := hwOutput.HwMetrics["swap"].(float64); ok {
			m.swap.Set(swap)
		} else {
			errorItems = append(errorItems, "swap")
		}
	}
	if settings.MetricBandwidthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricBandwidthPercent") {
		if metricBandwidthPercent, ok := hwOutput.HwMetrics["metricBandwidthPercent"].(float64); ok {
			m.metricBandwidthPercent.Set(metricBandwidthPercent)
		} else {
			errorItems = append(errorItems, "metricBandwidthPercent")
		}
	}
	if settings.MetricBlockSizeBytes && service.IsExistValue(hwOutput.HwMetrics, "metricBlockSizeBytes") {
		if metricBlockSizeBytes, ok := hwOutput.HwMetrics["metricBlockSizeBytes"].(float64); ok {
			m.metricBlockSizeBytes.Set(metricBlockSizeBytes)
		} else {
			errorItems = append(errorItems, "metricBlockSizeBytes")
		}
	}
	if settings.MetricOperatingSpeedMHz && service.IsExistValue(hwOutput.HwMetrics, "metricOperatingSpeedMHz") {
		if metricOperatingSpeedMHz, ok := hwOutput.HwMetrics["metricOperatingSpeedMHz"].(float64); ok {
			m.metricOperatingSpeedMHz.Set(metricOperatingSpeedMHz)
		} else {
			errorItems = append(errorItems, "metricOperatingSpeedMHz")
		}
	}

	if service.IsExistValue(hwOutput.HwMetrics, "metricHealthData") {
		if metricHealthData, ok := hwOutput.HwMetrics["metricHealthData"].(map[string]any); ok {
			if settings.MetricHealthData.DataLossDetected && service.IsExistValue(metricHealthData, "dataLossDetected") {
				if dataLossDetected, ok := metricHealthData["dataLossDetected"].(bool); ok {
					dataLossDetectedValue := dataLossDetectedFalseValue
					if dataLossDetected {
						dataLossDetectedValue = dataLossDetectedTrueValue
					}
					m.metricHealthDataDataLossDetectedTrue.Set(dataLossDetectedValue)
				} else {
					errorItems = append(errorItems, "dataLossDetected")
				}
			}
			if settings.MetricHealthData.LastShutdownSuccess && service.IsExistValue(metricHealthData, "lastShutdownSuccess") {
				if lastShutdownSuccess, ok := metricHealthData["lastShutdownSuccess"].(bool); ok {
					lastShutdownSuccessValue := lastShutdownSuccessFalseValue
					if lastShutdownSuccess {
						lastShutdownSuccessValue = lastShutdownSuccessTrueValue
					}
					m.metricHealthDataLastShutdownSuccessTrue.Set(lastShutdownSuccessValue)
				} else {
					errorItems = append(errorItems, "lastShutdownSuccess")
				}
			}
			if settings.MetricHealthData.PerformanceDegraded && service.IsExistValue(metricHealthData, "performanceDegraded") {
				if performanceDegraded, ok := metricHealthData["performanceDegraded"].(bool); ok {
					performanceDegradedValue := performanceDegradedFalseValue
					if performanceDegraded {
						performanceDegradedValue = performanceDegradedTrueValue
					}
					m.metricHealthDataPerformanceDegradedTrue.Set(performanceDegradedValue)
				} else {
					errorItems = append(errorItems, "performanceDegraded")
				}
			}
			if settings.MetricHealthData.PredictedMediaLifeLeftPercent && service.IsExistValue(metricHealthData, "predictedMediaLifeLeftPercent") {
				if predictedMediaLifeLeftPercent, ok := metricHealthData["predictedMediaLifeLeftPercent"].(float64); ok {
					m.metricHealthDataPredictedMediaLifeLeftPercent.Set(predictedMediaLifeLeftPercent)
				} else {
					errorItems = append(errorItems, "predictedMediaLifeLeftPercent")
				}
			}
		} else {
			errorItems = append(errorItems, "metricHealthData")
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

// Load the memory configuration file and store it in a struct
func loadMemoryConfig(filepath string, settings *model.MemoryConfig) error {
	buf, readErr := os.ReadFile(filepath)
	if readErr != nil {
		// If the file does not exist, treat all items as false and return nil.
		return nil
	}

	unmarshalErr := yaml.Unmarshal(buf, &settings)
	if unmarshalErr != nil {
		service.Log.Error(unmarshalErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0021", "Failed to unmarshal yaml.")
	}

	return nil
}

// Create Prometheus metrics from the HW control metric information and Memory metrics acquisition settings, and return them to Prometheus
func MemoryMetricsHandler(context *gin.Context, hwOutput *model.HwOutput) {
	settings := model.MemoryConfig{}

	readErr := loadMemoryConfig(memoryYamlFilePath, &settings)
	if readErr != nil {
		service.Log.Error(readErr.Error())
		context.JSON(service.GetStatusCode(readErr), service.ToJson(readErr))
		return
	}

	// Create a non-global registry.
	reg := prometheus.NewRegistry()
	// Create new metrics and register them using the custom registry.
	m, err := NewMemoryMetrics(reg, hwOutput, &settings)
	if err != nil {
		service.Log.Error(err.Error())
		context.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}
	// Set values for the new created metrics.
	setErr := setMemoryMetrics(m, hwOutput, &settings)
	if setErr != nil {
		service.Log.Error(setErr.Error())
		context.JSON(service.GetStatusCode(setErr), service.ToJson(setErr))
		return
	}

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(context.Writer, context.Request)
	service.Log.Info(context.Request.URL.Path + "[" + context.Request.Method + "] completed successfully.")
}
