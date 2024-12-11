package config

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testValidConfig = `
logger:
  level: debug
  out:
    - stdout
  type: console
  stacktrace: true
dispatcher:
  host: localhost:8080
`
	testInvalidConfig = `
\logger:level: debug
  out:
    - stdout
  type: console
  stacktrace: true
`
)

var (
	testExpectedConfig = CommonConfig{
		LoggerConfig: LoggerConfig{
			Level:      "debug",
			Out:        []string{"stdout"},
			Type:       "console",
			Stacktrace: true,
		},
	}
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Context("ReadConfigFromYAML", func() {
		var (
			configFile *os.File
		)

		BeforeEach(func() {
			tempDir := GinkgoT().TempDir()
			var err error
			configFile, err = os.Create(tempDir + "/config.yaml")
			Expect(err).To(Succeed())

			DeferCleanup(func() {
				configFile.Close()
			})
		})

		It("Sunny", func() {
			_, err := configFile.Write([]byte(testValidConfig))
			Expect(err).To(Succeed())

			conf, err := ReadConfigFromYAML[CommonConfig](configFile.Name())
			Expect(err).To(Succeed())
			Expect(*conf).To(Equal(testExpectedConfig))
		})

		It("Rainy", func() {
			_, err := configFile.Write([]byte(testInvalidConfig))
			Expect(err).To(Succeed())

			_, err = ReadConfigFromYAML[CommonConfig](configFile.Name())
			Expect(err).NotTo(Succeed())
		})
	})

	Context("ValidateConfig", func() {
		It("Sunny", func() {
			Expect(ValidateConfig(&testExpectedConfig)).To(Succeed())
		})

		It("Rainy", func() {
			testExpectedConfig.LoggerConfig.Level = "unkown"
			Expect(ValidateConfig(&testExpectedConfig)).NotTo(Succeed())
		})
	})
})
