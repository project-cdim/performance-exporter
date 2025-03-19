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
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCdimError_Error(t *testing.T) {
	tests := []struct {
		name string
		gce  *Error
		want string
	}{
		{
			name: "nomal",
			gce:  &Error{StatusCode: 400, Code: "400", Message: "Bad Request"},
			want: "http status code = 400, code = 400 message = Bad Request",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.gce.Error(); got != tt.want {
				t.Errorf("CdimError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCdimErrorNew(t *testing.T) {
	type args struct {
		statusCode int
		code       string
		message    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nomal",
			args: args{
				statusCode: 400,
				code:       "400",
				message:    "Bad Request",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrorNew(tt.args.statusCode, tt.args.code, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("CdimErrorNew() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "nomal",
			args: args{
				err: &Error{StatusCode: 400, Code: "400", Message: "Bad Request"},
			},
			want: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStatusCode(tt.args.err); got != tt.want {
				t.Errorf("GetStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToJson(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want gin.H
	}{
		{
			name: "nomal",
			args: args{
				err: &Error{StatusCode: 400, Code: "400", Message: "Bad Request"},
			},
			want: gin.H{"code": "400", "message": "Bad Request"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToJson(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToJson() = %v, want %v", got, tt.want)
			}
		})
	}
}
