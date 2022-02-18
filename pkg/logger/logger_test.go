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

package logger_test

import (
	"fmt"
	"io"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/vmware-tanzu/cartographer/pkg/logger"
)

var _ = Describe("Logger", func() {
	var (
		err              error
		logOpt           zap.Opts
		controllerOutput io.Writer
		logger           logr.Logger
		controllerBuffer *gbytes.Buffer
	)

	BeforeEach(func() {
		controllerBuffer = gbytes.NewBuffer()
		controllerOutput = io.MultiWriter(controllerBuffer, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		logger = zap.New(logOpt, zap.WriteTo(controllerOutput))
	})

	When("log output level is info", func() {
		BeforeEach(func() {
			logOpt, err = SetLogLevel("info")
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("log enabled is correct",
			func(loglevel int, expected bool) {
				Expect(logger.V(loglevel).Enabled()).To(Equal(expected))
			},
			Entry("info", INFO, true),
			Entry("debug", DEBUG, false),
		)

		It("outputs info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.V(INFO).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs default info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("does not output debug logs", func() {
			aNicePhrase := "Let's make the most of this beautiful day"
			logger.V(DEBUG).Info(aNicePhrase)
			Expect(controllerBuffer).NotTo(gbytes.Say(aNicePhrase))
		})

		It("outputs error logs", func() {
			aNiceErrorMessage := "would you be my, could you be my"
			aNiceErrorLogMessage := "Won't you be my neighbor?"

			logger.Error(fmt.Errorf(aNiceErrorMessage), aNiceErrorLogMessage)
			Expect(controllerBuffer).To(gbytes.Say(aNiceErrorMessage))

			logger.Error(fmt.Errorf(aNiceErrorMessage), aNiceErrorLogMessage)
			Expect(controllerBuffer).To(gbytes.Say(aNiceErrorLogMessage))
		})
	})

	When("log output level is error", func() {
		BeforeEach(func() {
			logOpt, err = SetLogLevel("error")
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("log enabled is correct",
			func(loglevel int, expected bool) {
				Expect(logger.V(loglevel).Enabled()).To(Equal(expected))

			},
			Entry("info", INFO, false),
			Entry("debug", DEBUG, false),
		)

		It("does not output info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.V(INFO).Info(aNicePhrase)
			Expect(controllerBuffer).NotTo(gbytes.Say(aNicePhrase))
		})

		It("does not output default info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.Info(aNicePhrase)
			Expect(controllerBuffer).NotTo(gbytes.Say(aNicePhrase))
		})

		It("does not output debug logs", func() {
			aNicePhrase := "Let's make the most of this beautiful day"
			logger.V(DEBUG).Info(aNicePhrase)
			Expect(controllerBuffer).NotTo(gbytes.Say(aNicePhrase))
		})

		It("outputs error logs", func() {
			aNiceErrorMessage := "would you be my, could you be my"
			aNiceErrorLogMessage := "Won't you be my neighbor?"

			logger.Error(fmt.Errorf(aNiceErrorMessage), aNiceErrorLogMessage)
			Expect(controllerBuffer).To(gbytes.Say(aNiceErrorMessage))

			logger.Error(fmt.Errorf(aNiceErrorMessage), aNiceErrorLogMessage)
			Expect(controllerBuffer).To(gbytes.Say(aNiceErrorLogMessage))
		})
	})

	When("log output level is debug", func() {
		BeforeEach(func() {
			logOpt, err = SetLogLevel("debug")
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("log enabled is correct",
			func(loglevel int, expected bool) {
				Expect(logger.V(loglevel).Enabled()).To(Equal(expected))
			},
			Entry("info", INFO, true),
			Entry("debug", DEBUG, true),
		)

		It("outputs info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.V(INFO).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs default info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs debug logs", func() {
			aNicePhrase := "Let's make the most of this beautiful day"
			logger.V(DEBUG).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs error logs", func() {
			aNiceErrorMessage := "would you be my, could you be my"
			aNiceErrorLogMessage := "Won't you be my neighbor?"

			logger.Error(fmt.Errorf(aNiceErrorMessage), aNiceErrorLogMessage)
			Expect(controllerBuffer).To(gbytes.Say(aNiceErrorMessage))

			logger.Error(fmt.Errorf(aNiceErrorMessage), aNiceErrorLogMessage)
			Expect(controllerBuffer).To(gbytes.Say(aNiceErrorLogMessage))
		})
	})

})
