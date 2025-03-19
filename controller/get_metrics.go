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
        
package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/project-cdim/performance-exporter/internal/collector"
	"github.com/project-cdim/performance-exporter/internal/model"
	"github.com/project-cdim/performance-exporter/internal/service"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const (
	defaultTimeout  int    = 10
	maxTimeout      int    = 36000
	minTimeout      int    = 1
	yamlFilePath    string = "configs/hw-performance-exporter.yaml"
	hwGetMetricsUrl string = "http://%s/cdim/api/v1/devices/%s/metrics"
)

type ExporterConfigs struct {
	Configs Config `yaml:"collect_configs"`
}

type Config struct {
	TargetHost string `yaml:"target_host"`
	TimeOut    int    `yaml:"timeout"`
}

// Execute HW control metric information, process it into Prometheus metrics, and return it to Prometheus
func GetMetrics(c *gin.Context) {
	service.Log.Info(c.Request.URL.Path + "[" + c.Request.Method + "] start.")
	settings := ExporterConfigs{}
	settings.Configs.TimeOut = defaultTimeout

	if err := loadConfig(yamlFilePath, &settings, c.Param("id")); err != nil {
		logAndRespondWithError(c, err)
		return
	}

	hwOutput := model.HwOutput{}
	if err := requestMetrics(&settings, &hwOutput, c.Param("id")); err != nil {
		logAndRespondWithError(c, err)
		return
	}

	if err := errorInfoExists(&hwOutput, c.Param("id")); err != nil {
		logAndRespondWithError(c, err)
		return
	}

	dispatchMetricsHandler(c, &hwOutput)
}

// Log error information and return an error message
func logAndRespondWithError(c *gin.Context, err error) {
	service.Log.Error(err.Error())
	c.JSON(service.GetStatusCode(err), service.ToJson(err))
}

// Return an error message if error information exists in the HW control system
func errorInfoExists(hwOutput *model.HwOutput, id string) error {
	if hwOutput.HwMetrics["errorInfo"] != nil {
		errorMessage := fmt.Sprintf("Error occurred in the HW control system. [id : %s]", id)
		service.Log.Error(errorMessage)
		return service.ErrorNew(http.StatusInternalServerError, "0014", errorMessage)
	}
	return nil
}

// Load the exporter configuration file
func loadConfig(filepath string, settings *ExporterConfigs, id string) error {
	buf, readErr := os.ReadFile(filepath)
	if readErr != nil {
		service.Log.Error(readErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0001", "Failed to read file.")
	}

	unmarshalErr := yaml.Unmarshal(buf, &settings)
	if unmarshalErr != nil {
		service.Log.Error(unmarshalErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0002", "Failed to unmarshal yaml.")
	}

	targetHost := settings.Configs.TargetHost
	if targetHost == "" {
		return service.ErrorNew(http.StatusInternalServerError, "0003", "target_host setting is required.")
	}
	// url.Parse allows relative paths. url.ParseRequestURI only allows absolute URIs or absolute paths.
	// Here, we want an absolute URI, so we parse it with url.ParseRequestURI.
	hw_metrics_url := fmt.Sprintf(hwGetMetricsUrl, settings.Configs.TargetHost, id)
	_, parseErr := url.ParseRequestURI(hw_metrics_url)
	if parseErr != nil {
		service.Log.Error(parseErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0004", "Format of the target_host or id is invalid.")
	}

	timeout := settings.Configs.TimeOut
	if timeout < minTimeout || timeout > maxTimeout {
		timeoutValue := strconv.Itoa(timeout)
		service.Log.Error("timeout value= " + timeoutValue)
		return service.ErrorNew(http.StatusInternalServerError, "0005", "Timeout value is out of range.")
	}
	return nil
}

// Get metric information from the HW control system
func requestMetrics(settings *ExporterConfigs, output *model.HwOutput, id string) error {
	// Since http.Client does not have a timeout set by default, set it
	httpClient := http.Client{Timeout: time.Duration(settings.Configs.TimeOut) * time.Second}

	hw_metrics_url := fmt.Sprintf(hwGetMetricsUrl, settings.Configs.TargetHost, id)

	resp, requestErr := httpClient.Get(hw_metrics_url)
	if requestErr != nil {
		service.Log.Error(requestErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0006", "Get request failure.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		statusCode := strconv.Itoa(resp.StatusCode)
		service.Log.Debug("status code= " + statusCode)

		if resp.StatusCode == http.StatusNotFound {
			msg := "No search results. [id : " + id + "]"
			return service.ErrorNew(http.StatusNotFound, "0007", msg)
		}
		return service.ErrorNew(http.StatusInternalServerError, "0008", "Collect target failure.")
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		service.Log.Error(readErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0009", "Failed to read response.")
	}
	unmarshalErr := json.Unmarshal(body, &output.HwMetrics)
	if unmarshalErr != nil {
		service.Log.Error(unmarshalErr.Error())
		return service.ErrorNew(http.StatusInternalServerError, "0010", "Failed to unmarshal response.")
	}

	return nil
}

// Branch the process of obtaining metric information according to the type of metric information
func dispatchMetricsHandler(c *gin.Context, hwOutput *model.HwOutput) {
	typeValue, ok := hwOutput.HwMetrics["type"]
	if !ok || typeValue == nil {
		service.Log.Error("Type value does not exist.")
		err := service.ErrorNew(http.StatusInternalServerError, "0011", "Type value does not exist.")
		c.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}

	deviceType, ok := typeValue.(string)
	if !ok {
		service.Log.Error("Invalid type value.")
		err := service.ErrorNew(http.StatusInternalServerError, "0012", "Invalid type value.")
		c.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}

	switch deviceType {
	case model.CPU, model.Accelerator, model.DSP, model.FPGA, model.GPU, model.UnknownProcessor:
		collector.ProcessorMetricsHandler(c, hwOutput)
	case model.Memory:
		collector.MemoryMetricsHandler(c, hwOutput)
	case model.Storage:
		collector.StorageMetricsHandler(c, hwOutput)
	case model.NetworkInterface:
		collector.NetworkInterfaceMetricsHandler(c, hwOutput)
	default:
		service.Log.Error("Invalid type value.")
		err := service.ErrorNew(http.StatusInternalServerError, "0013", "Invalid type value.")
		c.JSON(service.GetStatusCode(err), service.ToJson(err))
		return
	}
}
