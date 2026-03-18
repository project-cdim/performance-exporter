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
	processorYamlFilePath string = "configs/processor.yaml"
)

var namespaceProcessor = model.CPU

// Definition of Metrics
type processorMetrics struct {
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	statusStateEnabled         prometheus.Gauge
	statusHealthOk             prometheus.Gauge
	powerStateOn               prometheus.Gauge
	powerCapabilityTrue        prometheus.Gauge
	devicePortListLtssmStateL0 *prometheus.GaugeVec

	usageRate               prometheus.Gauge
	user                    prometheus.Gauge
	system                  prometheus.Gauge
	wait                    prometheus.Gauge
	idle                    prometheus.Gauge
	cpuWatts                prometheus.Gauge
	metricBandwidthPercent  prometheus.Gauge
	metricOperatingSpeedMHz prometheus.Gauge

	metricLocalMemoryBandwidthBytes  prometheus.Gauge
	metricRemoteMemoryBandwidthBytes prometheus.Gauge

	metricEnergyJoulesSensorResetTime prometheus.Gauge
	metricEnergyJoulesSensingInterval prometheus.Gauge
	metricEnergyJoulesReadingTime     prometheus.Gauge
	metricEnergyJoulesReading         prometheus.Gauge
}

// Define Processor metrics.
// Descriptor of metrics: "One of the pieces of information embedded in metrics (name, information to be placed on #HELP), and something other than numerical values to be displayed later in the graph"
func NewProcessorMetrics(reg prometheus.Registerer, hwOutput *model.HwOutput, settings *model.ProcessorConfig) (*processorMetrics, error) {
	m := &processorMetrics{
		// deviceID is not registered as a metric because it uses the job name
		// type is not registered as a metric because it uses the namespace
		// attribute is not registered as a metric because it is a fixed value

		statusStateEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceProcessor,
			Name:        "status_state",
			Help:        "Processor status information. Resource status",
			ConstLabels: prometheus.Labels{"value": "Enabled"},
		}),
		statusHealthOk: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceProcessor,
			Name:        "status_health",
			Help:        "Processor status information. Resource health status",
			ConstLabels: prometheus.Labels{"value": "OK"},
		}),
		powerStateOn: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceProcessor,
			Name:        "powerState",
			Help:        "Current power state of the processor.",
			ConstLabels: prometheus.Labels{"value": "On"},
		}),
		powerCapabilityTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceProcessor,
			Name:        "powerCapability",
			Help:        "Whether the power control function is enabled.",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),

		devicePortListLtssmStateL0: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespaceProcessor,
			Name:        "devicePortList_LTSSMState",
			Help:        "Link state of the CXL device",
			ConstLabels: prometheus.Labels{"value": "L0"},
		}, []string{"fabric_id", "switch_id", "switch_port_number"}),

		usageRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "usageRate",
			Help:      "CPU usage of the OS (user + system + wait values)",
		}),
		user: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "user",
			Help:      "CPU usage of the OS",
		}),
		system: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "system",
			Help:      "CPU usage of the OS",
		}),
		wait: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "wait",
			Help:      "CPU usage of the OS",
		}),
		idle: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "idle",
			Help:      "CPU usage of the OS",
		}),
		cpuWatts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "CPUWatts",
			Help:      "Collecting CPU power consumption of the OS",
		}),
		metricBandwidthPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricBandwidthPercent",
			Help:      "Processor bandwidth usage (%)",
		}),
		metricOperatingSpeedMHz: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricOperatingSpeedMHz",
			Help:      "Processor operating speed (MHz)",
		}),

		// metricCache is not registered because it is an array.
		// metricCoreMetrics is not registered because it is an array.

		metricLocalMemoryBandwidthBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricLocalMemoryBandwidthBytes",
			Help:      "Local memory bandwidth usage (in bytes)",
		}),
		metricRemoteMemoryBandwidthBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricRemoteMemoryBandwidthBytes",
			Help:      "Remote memory bandwidth usage (in bytes)",
		}),

		metricEnergyJoulesSensorResetTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricEnergyJoules_sensorResetTime",
			Help:      "Time property indicating the last reset date and time (UTC)",
		}),
		metricEnergyJoulesSensingInterval: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricEnergyJoules_sensingInterval",
			Help:      "Time interval between sensor readings (seconds)",
		}),
		metricEnergyJoulesReadingTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricEnergyJoules_readingTime",
			Help:      "Date and time when the measurement was taken from the sensor (UTC)",
		}),
		metricEnergyJoulesReading: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceProcessor,
			Name:      "metricEnergyJoules_reading",
			Help:      "Measured energy consumption (J)",
		}),
	}

	errorItems := []string{}
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

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

	if settings.DevicePortList.LTSSMState && service.IsExistValue(hwOutput.HwMetrics, "devicePortList") {
		reg.MustRegister(m.devicePortListLtssmStateL0)
	}

	if settings.UsageRate && service.IsExistValue(hwOutput.HwMetrics, "usageRate") {
		reg.MustRegister(m.usageRate)
	}
	if settings.User && service.IsExistValue(hwOutput.HwMetrics, "user") {
		reg.MustRegister(m.user)
	}
	if settings.System && service.IsExistValue(hwOutput.HwMetrics, "system") {
		reg.MustRegister(m.system)
	}
	if settings.Wait && service.IsExistValue(hwOutput.HwMetrics, "wait") {
		reg.MustRegister(m.wait)
	}
	if settings.Idle && service.IsExistValue(hwOutput.HwMetrics, "idle") {
		reg.MustRegister(m.idle)
	}
	if settings.CPUWatts && service.IsExistValue(hwOutput.HwMetrics, "CPUWatts") {
		reg.MustRegister(m.cpuWatts)
	}
	if settings.MetricBandwidthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricBandwidthPercent") {
		reg.MustRegister(m.metricBandwidthPercent)
	}
	if settings.MetricOperatingSpeedMHz && service.IsExistValue(hwOutput.HwMetrics, "metricOperatingSpeedMHz") {
		reg.MustRegister(m.metricOperatingSpeedMHz)
	}

	// metricCache is not registered because it is an array.
	// metricCoreMetrics is not registered because it is an array.

	if settings.MetricLocalMemoryBandwidthBytes && service.IsExistValue(hwOutput.HwMetrics, "metricLocalMemoryBandwidthBytes") {
		reg.MustRegister(m.metricLocalMemoryBandwidthBytes)
	}
	if settings.MetricRemoteMemoryBandwidthBytes && service.IsExistValue(hwOutput.HwMetrics, "metricRemoteMemoryBandwidthBytes") {
		reg.MustRegister(m.metricRemoteMemoryBandwidthBytes)
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

// Set the metrics values for the Processor.
func setProcessorMetrics(m *processorMetrics, hwOutput *model.HwOutput, settings *model.ProcessorConfig) error {
	errorItems := []string{}

	// attribute is not registered because it is a fixed value.

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

	if settings.DevicePortList.LTSSMState && service.IsExistValue(hwOutput.HwMetrics, "devicePortList") {
		if devicePortListRaw, ok := hwOutput.HwMetrics["devicePortList"].([]any); ok {
			for _, portDataRaw := range devicePortListRaw {
				if portDataMap, ok := portDataRaw.(map[string]any); ok {
					fabricID := service.GetStringValue(portDataMap, "fabricID")
					switchID := service.GetStringValue(portDataMap, "switchID")
					switchPortNumber := service.GetStringValue(portDataMap, "switchPortNumber")
					ltssmState := service.GetStringValue(portDataMap, "LTSSMState")

					if fabricID == "" {
						errorItems = append(errorItems, "devicePortList_fabricID")
						continue
					}

					if switchID == "" {
						errorItems = append(errorItems, "devicePortList_switchID")
						continue
					}

					// switchPortNumber is processed as empty string even if it doesn't exist or is empty
					// Continue without error

					if ltssmState == "" {
						errorItems = append(errorItems, "devicePortList_ltssmState")
						continue
					}

					labels := prometheus.Labels{
						"fabric_id":          fabricID,
						"switch_id":          switchID,
						"switch_port_number": switchPortNumber,
					}

					var ltssmStateValue float64
					switch ltssmState {
					case model.LtssmStateL0:
						ltssmStateValue = model.LtssmStateL0Value
					default:
						ltssmStateValue = model.LtssmStateOtherValue
					}

					m.devicePortListLtssmStateL0.With(labels).Set(ltssmStateValue)
				}
			}
		} else {
			errorItems = append(errorItems, "devicePortList")
		}
	}

	if settings.UsageRate && service.IsExistValue(hwOutput.HwMetrics, "usageRate") {
		if usageRate, ok := hwOutput.HwMetrics["usageRate"].(float64); ok {
			m.usageRate.Set(usageRate)
		} else {
			errorItems = append(errorItems, "usageRate")
		}
	}
	if settings.User && service.IsExistValue(hwOutput.HwMetrics, "user") {
		if user, ok := hwOutput.HwMetrics["user"].(float64); ok {
			m.user.Set(user)
		} else {
			errorItems = append(errorItems, "user")
		}
	}
	if settings.System && service.IsExistValue(hwOutput.HwMetrics, "system") {
		if system, ok := hwOutput.HwMetrics["system"].(float64); ok {
			m.system.Set(system)
		} else {
			errorItems = append(errorItems, "system")
		}
	}
	if settings.Wait && service.IsExistValue(hwOutput.HwMetrics, "wait") {
		if wait, ok := hwOutput.HwMetrics["wait"].(float64); ok {
			m.wait.Set(wait)
		} else {
			errorItems = append(errorItems, "wait")
		}
	}
	if settings.Idle && service.IsExistValue(hwOutput.HwMetrics, "idle") {
		if idle, ok := hwOutput.HwMetrics["idle"].(float64); ok {
			m.idle.Set(idle)
		} else {
			errorItems = append(errorItems, "idle")
		}
	}
	if settings.CPUWatts && service.IsExistValue(hwOutput.HwMetrics, "CPUWatts") {
		if cpuWatts, ok := hwOutput.HwMetrics["CPUWatts"].(float64); ok {
			m.cpuWatts.Set(cpuWatts)
		} else {
			errorItems = append(errorItems, "CPUWatts")
		}
	}
	if settings.MetricBandwidthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricBandwidthPercent") {
		if metricBandwidthPercent, ok := hwOutput.HwMetrics["metricBandwidthPercent"].(float64); ok {
			m.metricBandwidthPercent.Set(metricBandwidthPercent)
		} else {
			errorItems = append(errorItems, "metricBandwidthPercent")
		}
	}
	if settings.MetricOperatingSpeedMHz && service.IsExistValue(hwOutput.HwMetrics, "metricOperatingSpeedMHz") {
		if metricOperatingSpeedMHz, ok := hwOutput.HwMetrics["metricOperatingSpeedMHz"].(float64); ok {
			m.metricOperatingSpeedMHz.Set(metricOperatingSpeedMHz)
		} else {
			errorItems = append(errorItems, "metricOperatingSpeedMHz")
		}
	}

	// metricCache is not included because it is an array.
	// metricCoreMetrics is not included because it is an array.

	if settings.MetricLocalMemoryBandwidthBytes && service.IsExistValue(hwOutput.HwMetrics, "metricLocalMemoryBandwidthBytes") {
		if metricLocalMemoryBandwidthBytes, ok := hwOutput.HwMetrics["metricLocalMemoryBandwidthBytes"].(float64); ok {
			m.metricLocalMemoryBandwidthBytes.Set(metricLocalMemoryBandwidthBytes)
		} else {
			errorItems = append(errorItems, "metricLocalMemoryBandwidthBytes")
		}
	}
	if settings.MetricRemoteMemoryBandwidthBytes && service.IsExistValue(hwOutput.HwMetrics, "metricRemoteMemoryBandwidthBytes") {
		if metricRemoteMemoryBandwidthBytes, ok := hwOutput.HwMetrics["metricRemoteMemoryBandwidthBytes"].(float64); ok {
			m.metricRemoteMemoryBandwidthBytes.Set(metricRemoteMemoryBandwidthBytes)
		} else {
			errorItems = append(errorItems, "metricRemoteMemoryBandwidthBytes")
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

// Read the Processor configuration file and store it in a Struct
func loadProcessorConfig(filepath string, settings *model.ProcessorConfig) error {
	buf, readErr := os.ReadFile(filepath)
	if readErr != nil {
		// If the file does not exist, treat all items as false and return nil.
		return nil
	}

	unmarshalErr := yaml.Unmarshal(buf, &settings)
	if unmarshalErr != nil {
		service.Log.Error(unmarshalErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0011", "Failed to unmarshal yaml.")
	}

	return nil
}

// Create prometheus metrics from the HW control metric information and Processor metric acquisition settings, and return them to prometheus
func ProcessorMetricsHandler(context *gin.Context, hwOutput *model.HwOutput) {
	namespaceProcessor = hwOutput.HwMetrics["type"].(string)
	settings := model.ProcessorConfig{}

	readErr := loadProcessorConfig(processorYamlFilePath, &settings)
	if readErr != nil {
		service.Log.Error(readErr.Error())
		context.JSON(service.GetStatusCode(readErr), service.ToJson(readErr))
		return
	}

	// Create a non-global registry.
	reg := prometheus.NewRegistry()
	// Create new metrics and register them using the custom registry.
	m, err := NewProcessorMetrics(reg, hwOutput, &settings)
	if err != nil {
		service.Log.Error(err.Error())
		context.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}
	// Set values for the new created metrics.
	setErr := setProcessorMetrics(m, hwOutput, &settings)
	if setErr != nil {
		service.Log.Error(setErr.Error())
		context.JSON(service.GetStatusCode(setErr), service.ToJson(setErr))
		return
	}

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(context.Writer, context.Request)
	service.Log.Info(context.Request.URL.Path + "[" + context.Request.Method + "] completed successfully.")
}
