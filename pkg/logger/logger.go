// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	DEBUG = 1
	INFO  = 0
)

func SetLogLevel(logLevel string) (zap.Opts, error) {
	var level zapcore.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	default:
		return nil, fmt.Errorf("if present, log-level must be one of {error, info, debug}")
	}

	zapOpt := zap.Options{}
	zapOpt.Level = level

	return zap.UseFlagOptions(&zapOpt), nil
}
