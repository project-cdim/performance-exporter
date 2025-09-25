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
	"testing"
)

func Test_str2unix(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			"normal",
			args{
				"2006-01-02T15:04:05Z",
			},
			1136214245,
			false,
		},
		{
			"Error case: Input is in an unexpected format",
			args{
				"aaa",
			},
			0,
			true,
		},
		{
			"Error case: Input is an empty string",
			args{
				"",
			},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := str2unix(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("Str2unix() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Str2unix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExistValue(t *testing.T) {
	notNull := map[string]any{"test": "test"}
	valueIsNull := map[string]any{"test": nil}

	type args struct {
		target map[string]any
		item   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Normal case: Key exists and value is not null",
			args{
				notNull,
				"test",
			},
			true,
		},
		{
			"Normal case: Key exists and value is null",
			args{
				valueIsNull,
				"test",
			},
			false,
		},
		{
			"Normal case: Key does not exist",
			args{
				notNull,
				"other",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExistValue(tt.args.target, tt.args.item); got != tt.want {
				t.Errorf("IsExistValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeStr2Float(t *testing.T) {
	normal := map[string]any{
		"timeString": "2006-01-02T15:04:05Z",
	}
	invalid := map[string]any{
		"timeString": 55,
	}

	type args struct {
		target map[string]any
		item   string
	}
	tests := []struct {
		name  string
		args  args
		want  float64
		want1 bool
	}{
		{
			"normal",
			args{
				normal,
				"timeString",
			},
			float64(1136214245),
			true,
		},
		{
			"Abnormal case: When the data type is not as expected",
			args{
				invalid,
				"timeString",
			},
			float64(0),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := TimeStr2Float(tt.args.target, tt.args.item)
			if got != tt.want {
				t.Errorf("GetTime2FloatValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetTime2FloatValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
