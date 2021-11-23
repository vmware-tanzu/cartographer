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

package main

import (
	"flag"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/root"
)

var devMode bool
var port int
var certDir string
var verbosity string

func init() {
	flag.IntVar(&port, "Port", 9443, "Webhook server Port")
	flag.StringVar(&certDir, "cert-dir", "", "Webhook server tls dir")
	flag.BoolVar(&devMode, "dev", false, "Human readable logs")
	flag.StringVar(&verbosity, "log-level", "info", "Log levels")
	flag.Parse()
}

func main() {
	loggerOpt, err := logger.SetLogLevel(verbosity)
	if err != nil {
		panic(err)
	}

	cmd := root.Command{
		Port:    port,
		CertDir: certDir,
		Logger:  zap.New(zap.UseDevMode(devMode), loggerOpt),
	}

	if err = cmd.Execute(ctrl.SetupSignalHandler()); err != nil {
		panic(err)
	}
}
