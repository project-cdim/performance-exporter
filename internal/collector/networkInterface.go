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
	deviceEnabledTrueValue       float64 = 1
	deviceEnabledFalseValue      float64 = 0
	networkInterfaceYamlFilePath string  = "configs/networkInterface.yaml"
)

var namespaceNetworkInterface = model.NetworkInterface

// Definition of Metrics
type networkInterfaceMetrics struct {
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	deviceEnabled      prometheus.Gauge
	statusStateEnabled prometheus.Gauge
	statusHealthOk     prometheus.Gauge

	powerStateOn               prometheus.Gauge
	powerCapabilityTrue        prometheus.Gauge
	devicePortListLtssmStateL0 *prometheus.GaugeVec

	networkInterfaceInformationNetworkTrafficReceivePackets  prometheus.Gauge
	networkInterfaceInformationNetworkTrafficTransmitPackets prometheus.Gauge
	networkInterfaceInformationNetworkTrafficBytesSent       prometheus.Gauge
	networkInterfaceInformationNetworkTrafficBytesRecv       prometheus.Gauge
	networkInterfaceInformationNetworkTrafficErrin           prometheus.Gauge
	networkInterfaceInformationNetworkTrafficErrout          prometheus.Gauge
	networkInterfaceInformationNetworkTrafficDropin          prometheus.Gauge
	networkInterfaceInformationNetworkTrafficDropout         prometheus.Gauge
	networkInterfaceInformationNetworkSpeed                  prometheus.Gauge

	metricsCPUCorePercent         prometheus.Gauge
	metricsHostBusRXPercent       prometheus.Gauge
	metricsHostBusTXPercent       prometheus.Gauge
	metricsRXAvgQueueDepthPercent prometheus.Gauge
	metricsTXAvgQueueDepthPercent prometheus.Gauge
	metricsRXBytes                prometheus.Gauge
	metricsRXFrames               prometheus.Gauge
	metricsTXBytes                prometheus.Gauge
	metricsTXFrames               prometheus.Gauge

	metricEnergyJoulesSensorResetTime prometheus.Gauge
	metricEnergyJoulesSensingInterval prometheus.Gauge
	metricEnergyJoulesReadingTime     prometheus.Gauge
	metricEnergyJoulesReading         prometheus.Gauge
}

// Define the metrics for networkInterface.
// Descriptor of metrics "One of the pieces of information embedded in metrics (name, information to be placed on #HELP), and non-numeric information to be displayed later in the graph"
func NewNetworkInterfaceMetrics(reg prometheus.Registerer, hwOutput *model.HwOutput, settings *model.NetworkInterfaceConfig) (*networkInterfaceMetrics, error) {
	m := &networkInterfaceMetrics{
		// deviceID is not registered as a metric because it uses the job name
		// type is not registered as a metric because it uses the namespace
		// attribute is not registered as a metric because it is a fixed value

		deviceEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceNetworkInterface,
			Name:        "deviceEnabled",
			Help:        "Indicates whether this network device function is enabled",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),
		statusStateEnabled: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceNetworkInterface,
			Name:        "status_state",
			Help:        "Network status information. Resource status",
			ConstLabels: prometheus.Labels{"value": "Enabled"},
		}),
		statusHealthOk: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceNetworkInterface,
			Name:        "status_health",
			Help:        "Network status information. Resource health status",
			ConstLabels: prometheus.Labels{"value": "OK"},
		}),
		powerStateOn: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceNetworkInterface,
			Name:        "powerState",
			Help:        "Current power state of the network function.",
			ConstLabels: prometheus.Labels{"value": "On"},
		}),
		powerCapabilityTrue: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespaceNetworkInterface,
			Name:        "powerCapability",
			Help:        "Whether power control function is enabled",
			ConstLabels: prometheus.Labels{"value": "true"},
		}),

		devicePortListLtssmStateL0: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespaceNetworkInterface,
			Name:        "devicePortList_LTSSMState",
			Help:        "CXL device link state",
			ConstLabels: prometheus.Labels{"value": "L0"},
		}, []string{"fabric_id", "switch_id", "switch_port_number"}),

		networkInterfaceInformationNetworkSpeed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkSpeed",
			Help:      "Network speed of network statistics",
		}),
		networkInterfaceInformationNetworkTrafficReceivePackets: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_receivePackets",
			Help:      "Number of received packets in network statistics",
		}),
		networkInterfaceInformationNetworkTrafficTransmitPackets: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_transmitPackets",
			Help:      "Number of transmitted packets in network statistics",
		}),
		networkInterfaceInformationNetworkTrafficBytesSent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_bytesSent",
			Help:      "Number of bytes sent in network statistics (unit: bytes)",
		}),
		networkInterfaceInformationNetworkTrafficBytesRecv: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_bytesRecv",
			Help:      "Number of bytes received in network statistics (unit: bytes)",
		}),
		networkInterfaceInformationNetworkTrafficErrin: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_errin",
			Help:      "Total number of errors during reception in network statistics",
		}),
		networkInterfaceInformationNetworkTrafficErrout: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_errout",
			Help:      "Total number of errors during transmission in network statistics",
		}),
		networkInterfaceInformationNetworkTrafficDropin: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_dropin",
			Help:      "Total number of dropped received packets in network statistics",
		}),
		networkInterfaceInformationNetworkTrafficDropout: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "networkInterfaceInformation_networkTraffic_dropout",
			Help:      "Total number of dropped transmitted packets in network statistics",
		}),

		metricsCPUCorePercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsCPUCorePercent",
			Help:      "CPU core usage rate of the device (percentage)",
		}),
		metricsHostBusRXPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsHostBusRXPercent",
			Help:      "Host bus RX usage rate, such as PCIe (percentage)",
		}),
		metricsHostBusTXPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsHostBusTXPercent",
			Help:      "Host bus TX usage rate, such as PCIe (percentage)",
		}),
		metricsRXAvgQueueDepthPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsRXAvgQueueDepthPercent",
			Help:      "RX average queue depth.",
		}),
		metricsTXAvgQueueDepthPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsTXAvgQueueDepthPercent",
			Help:      "TX average queue depth.",
		}),
		metricsRXBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsRXBytes",
			Help:      "Total number of bytes received by the network function.",
		}),
		metricsRXFrames: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsRXFrames",
			Help:      "Total number of frames received by the network function.",
		}),
		metricsTXBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsTXBytes",
			Help:      "Total number of bytes sent by the network function.",
		}),
		metricsTXFrames: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricsTXFrames",
			Help:      "Total number of frames sent by the network function.",
		}),

		metricEnergyJoulesSensorResetTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricEnergyJoules_sensorResetTime",
			Help:      "The date and time when the time property was last reset. (UTC)",
		}),
		metricEnergyJoulesSensingInterval: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricEnergyJoules_sensingInterval",
			Help:      "Time interval between sensor readings (seconds)",
		}),
		metricEnergyJoulesReadingTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricEnergyJoules_readingTime",
			Help:      "The date and time when the measurement was obtained from the sensor. (UTC)",
		}),
		metricEnergyJoulesReading: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespaceNetworkInterface,
			Name:      "metricEnergyJoules_reading",
			Help:      "Measurement of energy consumption (J)",
		}),
	}

	errorItems := []string{}
	// deviceID is not registered as a metric because it uses the job name
	// type is not registered as a metric because it uses the namespace
	// attribute is not registered as a metric because it is a fixed value

	if settings.DeviceEnabled && service.IsExistValue(hwOutput.HwMetrics, "deviceEnabled") {
		reg.MustRegister(m.deviceEnabled)
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

	if settings.DevicePortList.LTSSMState && service.IsExistValue(hwOutput.HwMetrics, "devicePortList") {
		reg.MustRegister(m.devicePortListLtssmStateL0)
	}

	if service.IsExistValue(hwOutput.HwMetrics, "networkInterfaceInformation") {
		if networkInterfaceInformation, ok := hwOutput.HwMetrics["networkInterfaceInformation"].(map[string]any); ok {
			if settings.NetworkInterfaceInformation.NetworkSpeed && service.IsExistValue(networkInterfaceInformation, "networkSpeed") {
				reg.MustRegister(m.networkInterfaceInformationNetworkSpeed)
			}
			if service.IsExistValue(networkInterfaceInformation, "networkTraffic") {
				if networkTraffic, ok := networkInterfaceInformation["networkTraffic"].(map[string]any); ok {
					if settings.NetworkInterfaceInformation.NetworkTraffic.ReceivePackets && service.IsExistValue(networkTraffic, "receivePackets") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficReceivePackets)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.TransmitPackets && service.IsExistValue(networkTraffic, "transmitPackets") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficTransmitPackets)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.BytesSent && service.IsExistValue(networkTraffic, "bytesSent") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficBytesSent)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.BytesRecv && service.IsExistValue(networkTraffic, "bytesRecv") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficBytesRecv)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Errin && service.IsExistValue(networkTraffic, "errin") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficErrin)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Errout && service.IsExistValue(networkTraffic, "errout") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficErrout)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Dropin && service.IsExistValue(networkTraffic, "dropin") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficDropin)
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Dropout && service.IsExistValue(networkTraffic, "dropout") {
						reg.MustRegister(m.networkInterfaceInformationNetworkTrafficDropout)
					}
				} else {
					errorItems = append(errorItems, "networkTraffic")
				}
			}
		} else {
			errorItems = append(errorItems, "networkInterfaceInformation")
		}
	}

	if settings.MetricsCPUCorePercent && service.IsExistValue(hwOutput.HwMetrics, "metricsCPUCorePercent") {
		reg.MustRegister(m.metricsCPUCorePercent)
	}
	if settings.MetricsHostBusRXPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsHostBusRXPercent") {
		reg.MustRegister(m.metricsHostBusRXPercent)
	}
	if settings.MetricsHostBusTXPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsHostBusTXPercent") {
		reg.MustRegister(m.metricsHostBusTXPercent)
	}
	if settings.MetricsRXAvgQueueDepthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsRXAvgQueueDepthPercent") {
		reg.MustRegister(m.metricsRXAvgQueueDepthPercent)
	}
	if settings.MetricsTXAvgQueueDepthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsTXAvgQueueDepthPercent") {
		reg.MustRegister(m.metricsTXAvgQueueDepthPercent)
	}
	if settings.MetricsRXBytes && service.IsExistValue(hwOutput.HwMetrics, "metricsRXBytes") {
		reg.MustRegister(m.metricsRXBytes)
	}
	if settings.MetricsRXFrames && service.IsExistValue(hwOutput.HwMetrics, "metricsRXFrames") {
		reg.MustRegister(m.metricsRXFrames)
	}
	if settings.MetricsTXBytes && service.IsExistValue(hwOutput.HwMetrics, "metricsTXBytes") {
		reg.MustRegister(m.metricsTXBytes)
	}
	if settings.MetricsTXFrames && service.IsExistValue(hwOutput.HwMetrics, "metricsTXFrames") {
		reg.MustRegister(m.metricsTXFrames)
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

// Set the metrics value for NetworkInterface.
func setNetworkInterfaceMetrics(m *networkInterfaceMetrics, hwOutput *model.HwOutput, settings *model.NetworkInterfaceConfig) error {
	errorItems := []string{}

	// attribute is not registered because it is a fixed value.

	if settings.DeviceEnabled && service.IsExistValue(hwOutput.HwMetrics, "deviceEnabled") {
		if deviceEnabled, ok := hwOutput.HwMetrics["deviceEnabled"].(bool); ok {
			deviceEnabledValue := deviceEnabledFalseValue
			if deviceEnabled {
				deviceEnabledValue = deviceEnabledTrueValue
			}
			m.deviceEnabled.Set(deviceEnabledValue)
		} else {
			errorItems = append(errorItems, "deviceEnabled")
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

	if service.IsExistValue(hwOutput.HwMetrics, "networkInterfaceInformation") {
		if networkInterfaceInformation, ok := hwOutput.HwMetrics["networkInterfaceInformation"].(map[string]any); ok {
			if settings.NetworkInterfaceInformation.NetworkSpeed && service.IsExistValue(networkInterfaceInformation, "networkSpeed") {
				if networkSpeed, ok := networkInterfaceInformation["networkSpeed"].(float64); ok {
					m.networkInterfaceInformationNetworkSpeed.Set(networkSpeed)
				} else {
					errorItems = append(errorItems, "networkSpeed")
				}
			}
			if service.IsExistValue(networkInterfaceInformation, "networkTraffic") {
				if networkTraffic, ok := networkInterfaceInformation["networkTraffic"].(map[string]any); ok {
					if settings.NetworkInterfaceInformation.NetworkTraffic.ReceivePackets && service.IsExistValue(networkTraffic, "receivePackets") {
						if receivePackets, ok := networkTraffic["receivePackets"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficReceivePackets.Set(receivePackets)
						} else {
							errorItems = append(errorItems, "receivePackets")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.TransmitPackets && service.IsExistValue(networkTraffic, "transmitPackets") {
						if transmitPackets, ok := networkTraffic["transmitPackets"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficTransmitPackets.Set(transmitPackets)
						} else {
							errorItems = append(errorItems, "transmitPackets")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.BytesSent && service.IsExistValue(networkTraffic, "bytesSent") {
						if bytesSent, ok := networkTraffic["bytesSent"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficBytesSent.Set(bytesSent)
						} else {
							errorItems = append(errorItems, "bytesSent")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.BytesRecv && service.IsExistValue(networkTraffic, "bytesRecv") {
						if bytesRecv, ok := networkTraffic["bytesRecv"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficBytesRecv.Set(bytesRecv)
						} else {
							errorItems = append(errorItems, "bytesRecv")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Errin && service.IsExistValue(networkTraffic, "errin") {
						if errin, ok := networkTraffic["errin"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficErrin.Set(errin)
						} else {
							errorItems = append(errorItems, "errin")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Errout && service.IsExistValue(networkTraffic, "errout") {
						if errout, ok := networkTraffic["errout"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficErrout.Set(errout)
						} else {
							errorItems = append(errorItems, "errout")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Dropin && service.IsExistValue(networkTraffic, "dropin") {
						if dropin, ok := networkTraffic["dropin"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficDropin.Set(dropin)
						} else {
							errorItems = append(errorItems, "dropin")
						}
					}
					if settings.NetworkInterfaceInformation.NetworkTraffic.Dropout && service.IsExistValue(networkTraffic, "dropout") {
						if dropout, ok := networkTraffic["dropout"].(float64); ok {
							m.networkInterfaceInformationNetworkTrafficDropout.Set(dropout)
						} else {
							errorItems = append(errorItems, "dropout")
						}
					}
				} else {
					errorItems = append(errorItems, "networkTraffic")
				}
			}
		} else {
			errorItems = append(errorItems, "networkInterfaceInformation")
		}
	}

	if settings.MetricsCPUCorePercent && service.IsExistValue(hwOutput.HwMetrics, "metricsCPUCorePercent") {
		if metricsCPUCorePercent, ok := hwOutput.HwMetrics["metricsCPUCorePercent"].(float64); ok {
			m.metricsCPUCorePercent.Set(metricsCPUCorePercent)
		} else {
			errorItems = append(errorItems, "metricsCPUCorePercent")
		}
	}
	if settings.MetricsHostBusRXPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsHostBusRXPercent") {
		if metricsHostBusRXPercent, ok := hwOutput.HwMetrics["metricsHostBusRXPercent"].(float64); ok {
			m.metricsHostBusRXPercent.Set(metricsHostBusRXPercent)
		} else {
			errorItems = append(errorItems, "metricsHostBusRXPercent")
		}
	}
	if settings.MetricsHostBusTXPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsHostBusTXPercent") {
		if metricsHostBusTXPercent, ok := hwOutput.HwMetrics["metricsHostBusTXPercent"].(float64); ok {
			m.metricsHostBusTXPercent.Set(metricsHostBusTXPercent)
		} else {
			errorItems = append(errorItems, "metricsHostBusTXPercent")
		}
	}
	if settings.MetricsRXAvgQueueDepthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsRXAvgQueueDepthPercent") {
		if metricsRXAvgQueueDepthPercent, ok := hwOutput.HwMetrics["metricsRXAvgQueueDepthPercent"].(float64); ok {
			m.metricsRXAvgQueueDepthPercent.Set(metricsRXAvgQueueDepthPercent)
		} else {
			errorItems = append(errorItems, "metricsRXAvgQueueDepthPercent")
		}
	}
	if settings.MetricsTXAvgQueueDepthPercent && service.IsExistValue(hwOutput.HwMetrics, "metricsTXAvgQueueDepthPercent") {
		if metricsTXAvgQueueDepthPercent, ok := hwOutput.HwMetrics["metricsTXAvgQueueDepthPercent"].(float64); ok {
			m.metricsTXAvgQueueDepthPercent.Set(metricsTXAvgQueueDepthPercent)
		} else {
			errorItems = append(errorItems, "metricsTXAvgQueueDepthPercent")
		}
	}
	if settings.MetricsRXBytes && service.IsExistValue(hwOutput.HwMetrics, "metricsRXBytes") {
		if metricsRXBytes, ok := hwOutput.HwMetrics["metricsRXBytes"].(float64); ok {
			m.metricsRXBytes.Set(metricsRXBytes)
		} else {
			errorItems = append(errorItems, "metricsRXBytes")
		}
	}
	if settings.MetricsRXFrames && service.IsExistValue(hwOutput.HwMetrics, "metricsRXFrames") {
		if metricsRXFrames, ok := hwOutput.HwMetrics["metricsRXFrames"].(float64); ok {
			m.metricsRXFrames.Set(metricsRXFrames)
		} else {
			errorItems = append(errorItems, "metricsRXFrames")
		}
	}
	if settings.MetricsTXBytes && service.IsExistValue(hwOutput.HwMetrics, "metricsTXBytes") {
		if metricsTXBytes, ok := hwOutput.HwMetrics["metricsTXBytes"].(float64); ok {
			m.metricsTXBytes.Set(metricsTXBytes)
		} else {
			errorItems = append(errorItems, "metricsTXBytes")
		}
	}
	if settings.MetricsTXFrames && service.IsExistValue(hwOutput.HwMetrics, "metricsTXFrames") {
		if metricsTXFrames, ok := hwOutput.HwMetrics["metricsTXFrames"].(float64); ok {
			m.metricsTXFrames.Set(metricsTXFrames)
		} else {
			errorItems = append(errorItems, "metricsTXFrames")
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

// Load the NetworkInterface configuration file and store it in a struct
func loadNetworkInterfaceConfig(filepath string, settings *model.NetworkInterfaceConfig) error {
	buf, readErr := os.ReadFile(filepath)
	if readErr != nil {
		// If the file does not exist, treat all items as false and return nil.
		return nil
	}

	unmarshalErr := yaml.Unmarshal(buf, &settings)
	if unmarshalErr != nil {
		service.Log.Error(unmarshalErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0041", "Failed to unmarshal yaml.")
	}

	return nil
}

// Create Prometheus metrics from the HW control metric information and networkInterface metric acquisition settings, and return them to Prometheus
func NetworkInterfaceMetricsHandler(context *gin.Context, hwOutput *model.HwOutput) {
	namespaceNetworkInterface = hwOutput.HwMetrics["type"].(string)
	settings := model.NetworkInterfaceConfig{}

	readErr := loadNetworkInterfaceConfig(networkInterfaceYamlFilePath, &settings)
	if readErr != nil {
		service.Log.Error(readErr.Error())
		context.JSON(service.GetStatusCode(readErr), service.ToJson(readErr))
		return
	}

	// Create a non-global registry.
	reg := prometheus.NewRegistry()
	// Create new metrics and register them using the custom registry.
	m, err := NewNetworkInterfaceMetrics(reg, hwOutput, &settings)
	if err != nil {
		service.Log.Error(err.Error())
		context.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}
	// Set values for the new created metrics.
	setErr := setNetworkInterfaceMetrics(m, hwOutput, &settings)
	if setErr != nil {
		service.Log.Error(setErr.Error())
		context.JSON(service.GetStatusCode(setErr), service.ToJson(setErr))
		return
	}

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(context.Writer, context.Request)
	service.Log.Info(context.Request.URL.Path + "[" + context.Request.Method + "] completed successfully.")
}
