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

func SetLogLevel(logLevel string) (zap.Opts, error) {
	var level zapcore.Level
	switch strings.ToUpper(logLevel) {
	case "TRACE":
		level = zapcore.Level(-2)
	case "DEBUG":
		level = zapcore.Level(-1)
	case "INFO":
		level = zapcore.Level(0)
	case "ERROR":
		level = zapcore.Level(1)
	default:
		return nil, fmt.Errorf("if present, log-level must be one of {error, info, debug, trace}")
	}

	zapOpt := zap.Options{}
	zapOpt.Level = level

	return zap.UseFlagOptions(&zapOpt), nil
}
