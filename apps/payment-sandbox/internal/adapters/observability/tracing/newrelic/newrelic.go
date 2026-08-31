package newrelictracing

import (
	"strings"

	nr "github.com/newrelic/go-agent/v3/newrelic"
)

func Setup(appName, licenseKey string) (*nr.Application, error) {
	if strings.TrimSpace(licenseKey) == "" {
		return nil, nil
	}
	if strings.TrimSpace(appName) == "" {
		appName = "payment-sandbox"
	}
	app, err := nr.NewApplication(
		nr.ConfigAppName(appName),
		nr.ConfigLicense(licenseKey),
	)
	if err != nil {
		return nil, err
	}
	return app, nil
}
