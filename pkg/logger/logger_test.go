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
		err error
		logOpt zap.Opts
		controllerOutput io.Writer
		logger logr.Logger
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
			Entry("info", 0, true),
			Entry("debug", 1, false),
			Entry("trace", 2, false),
		)

		It("outputs info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.V(0).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("does not output debug logs", func() {
			aNicePhrase := "Let's make the most of this beautiful day"
			logger.V(1).Info(aNicePhrase)
			Expect(controllerBuffer).NotTo(gbytes.Say(aNicePhrase))
		})

		It("does not output trace logs", func() {
			aNicePhrase := "Since we're together, might as well say"
			logger.V(2).Info(aNicePhrase)
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
			Entry("info", 0, true),
			Entry("debug", 1, true),
			Entry("trace", 2, false),
		)

		It("outputs info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.V(0).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs debug logs", func() {
			aNicePhrase := "Let's make the most of this beautiful day"
			logger.V(1).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("does not output trace logs", func() {
			aNicePhrase := "Since we're together, might as well say"
			logger.V(2).Info(aNicePhrase)
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

	When("log output level is trace", func() {
		BeforeEach(func() {
			logOpt, err = SetLogLevel("trace")
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("log enabled is correct",
			func(loglevel int, expected bool) {
				Expect(logger.V(loglevel).Enabled()).To(Equal(expected))
			},
			Entry("info", 0, true),
			Entry("debug", 1, true),
			Entry("trace", 2, true),
		)

		It("outputs info logs", func() {
			aNicePhrase := "It's a beautiful day in the neighborhood"
			logger.V(0).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs debug logs", func() {
			aNicePhrase := "Let's make the most of this beautiful day"
			logger.V(1).Info(aNicePhrase)
			Expect(controllerBuffer).To(gbytes.Say(aNicePhrase))
		})

		It("outputs trace logs", func() {
			aNicePhrase := "Since we're together, might as well say"
			logger.V(2).Info(aNicePhrase)
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