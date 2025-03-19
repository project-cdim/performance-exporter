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
        
package service

import (
	"net/http"
	"time"
)

// Convert time string (e.g., "2006-01-02T15:04:05Z") to UNIX time
func str2unix(t string) (int64, error) {
	// String -> time.Time type
	parsedTime, parseErr := time.Parse(time.RFC3339, t)
	if parseErr != nil {
		return int64(0), parseErr
	}
	// time.Time type -> UNIX time
	unixTime := parsedTime.Unix()
	return unixTime, nil
}

// Existence check. Check that the key of map[string]any{} exists and the value is not null.
func IsExistValue(target map[string]any, item string) bool {
	if targetValue, ok := target[item]; ok && targetValue != nil {
		return true
	}
	return false
}

// Convert a string in the format "2023-10-01T00:00:00Z" to UNIX time and then to float64
func TimeStr2Float(target map[string]any, item string) (float64, bool) {
	if stringValue, ok := target[item].(string); ok {
		value, err := str2unix(stringValue)
		if err != nil {
			return float64(0), false
		}
		return float64(value), true
	}
	return float64(0), false
}

func CreateErroeMessage(item string) error {
	msg := "Data type differs from expected. [items : " + item + "]"
	return ErrorNew(http.StatusInternalServerError, "0050", msg)
}
