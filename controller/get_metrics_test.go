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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/project-cdim/performance-exporter/internal/model"
	"github.com/project-cdim/performance-exporter/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	t.Skip("not test")
}

func TestLogAndRespondWithError(t *testing.T) {
	// Define test cases
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "BadRequestError",
			err:          service.ErrorNew(http.StatusBadRequest, "TEST_ERROR", "This is a test error"),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"TEST_ERROR","message":"This is a test error"}`,
		},
		{
			name:         "InternalServerError",
			err:          service.ErrorNew(http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_ERROR","message":"An internal error occurred"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gin context and response recorder
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Execute function
			logAndRespondWithError(c, tc.err)

			// Verify status code and response body
			assert.Equal(t, tc.expectedCode, w.Code)
			assert.JSONEq(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestErrorInfoExists(t *testing.T) {
	tests := []struct {
		name      string
		hwOutput  *model.HwOutput
		id        string
		expectErr bool
		errorCode string
		errorMsg  string
	}{
		{
			name: "ErrorInfoExists",
			hwOutput: &model.HwOutput{
				HwMetrics: map[string]interface{}{
					"errorInfo": "Some error info",
				},
			},
			id:        "123",
			expectErr: true,
			errorCode: "0014",
			errorMsg:  "Error occurred in the HW control system. [id : 123]",
		},
		{
			name: "ErrorInfoNotExists",
			hwOutput: &model.HwOutput{
				HwMetrics: map[string]interface{}{},
			},
			id:        "456",
			expectErr: false,
		},
		{
			name: "ErrorInfoIsNull",
			hwOutput: &model.HwOutput{
				HwMetrics: map[string]interface{}{
					"errorInfo": nil,
				},
			},
			id:        "789",
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := errorInfoExists(tc.hwOutput, tc.id)

			if tc.expectErr {
				assert.Error(t, err, "No error occurred")
				if err != nil {
					cdimErr, ok := err.(*service.Error) // CdimError is assumed to be an error type defined in the service package
					assert.True(t, ok)
					assert.Equal(t, tc.errorCode, cdimErr.Code)
					assert.Equal(t, tc.errorMsg, cdimErr.Message)
				}
			} else {
				assert.NoError(t, err, "An error occurred")
			}
		})
	}
}

func Test_loadConfig(t *testing.T) {
	const settingsPath = "testdata/settings/"
	const deviceID = "b9f9b678-bce8-49a1-bca3-dbbc32259abf"

	yamlContent := ExporterConfigs{}
	yamlContent.Configs.TimeOut = defaultTimeout

	type args struct {
		filepath string
		settings ExporterConfigs
		id       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"nomal",
			args{
				settingsPath + "normal.yaml",
				yamlContent,
				deviceID,
			},
			false,
		},
		{
			"Error: yaml file does not exist",
			args{
				settingsPath + "aaa.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Error: host is required and cannot be omitted",
			args{
				settingsPath + "host_empty.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Error: host value is null",
			args{
				settingsPath + "host_is_null.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Error: host value is not in host format",
			args{
				settingsPath + "host_format_invalid.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Error: file is not in yaml format",
			args{
				settingsPath + "textonly.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Error: Boundary test. timeout < lower limit",
			args{
				settingsPath + "timeout_0.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Normal: Boundary test. timeout = lower limit",
			args{
				settingsPath + "timeout_1.yaml",
				yamlContent,
				deviceID,
			},
			false,
		},
		{
			"Normal: Boundary test. timeout = upper limit",
			args{
				settingsPath + "timeout_36000.yaml",
				yamlContent,
				deviceID,
			},
			false,
		},
		{
			"Error: Boundary test. timeout > upper limit",
			args{
				settingsPath + "timeout_36001.yaml",
				yamlContent,
				deviceID,
			},
			true,
		},
		{
			"Normal: Use default value when timeout is omitted",
			args{
				settingsPath + "timeout_empty.yaml",
				yamlContent,
				deviceID,
			},
			false,
		},
		{
			"Normal: Use default value when timeout is null",
			args{
				settingsPath + "timeout_is_null.yaml",
				yamlContent,
				deviceID,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadConfig(tt.args.filepath, &tt.args.settings, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_requestMetrics(t *testing.T) {
	t.Skip("not test")
}

func Test_dispatchMetricsHandler(t *testing.T) {
	const hwInfoPath = "testdata/hw_info/"

	// Create gin context
	w := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/dummy", nil)
	ginContext.Request = req

	type args struct {
		c        *gin.Context
		hwOutput *model.HwOutput
	}
	tests := []struct {
		name string
		args args
	}{
		// Pattern for obtaining HW control metric information. Confirm that panic does not occur
		{
			"nomal(type:CPU)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "CPU.json"),
			},
		},
		{
			"nomal(type:Accelerator)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "Accelerator.json"),
			},
		},
		{
			"nomal(type:DSP)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "DSP.json"),
			},
		},
		{
			"nomal(type:FPGA)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "FPGA.json"),
			},
		},
		{
			"nomal(type:GPU)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "GPU.json"),
			},
		},
		{
			"nomal(type:UnknownProcessor)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "UnknownProcessor.json"),
			},
		},
		{
			"nomal(type:memory)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "memory.json"),
			},
		},
		{
			"nomal(type:storage)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "storage.json"),
			},
		},
		{
			"nomal(type:networkInterface)",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "networkInterface.json"),
			},
		},
		{
			"Error: type element does not exist",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "not_exist_type.json"),
			},
		},
		{
			"Error: type value is an unexpected string",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "type_aaa.json"),
			},
		},
		{
			"Error: type value is null",
			args{
				ginContext,
				setHwMetric(hwInfoPath + "is_null_type.json"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatchMetricsHandler(tt.args.c, tt.args.hwOutput)
		})
	}
}

func setHwMetric(filepath string) *model.HwOutput {
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
	return &hwOutput
}
