package client

import (
	"net/http"

	"github.com/dcos/dcos-go/dcos/http/transport"
)

// NewClient returns a new http client ready to handle DC/OS security.
// A caller can optionally pass a configured *http.Client (with the right timeout for instance), in this case
// NewClient will replace Transport.
func NewClient(client *http.Client, iamConfig, caCert string) (*http.Client, error) {
	transportOptions := []transport.OptionTransportFunc{}

	// add appropriate options based on config files
	if iamConfig != "" {
		transportOptions = append(transportOptions, transport.OptionIAMConfigPath(iamConfig))
	}

	if caCert != "" {
		transportOptions = append(transportOptions, transport.OptionCaCertificatePath(caCert))
	}

	tr, err := transport.NewTransport(transportOptions...)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = &http.Client{}
	}

	client.Transport = tr

	return client, nil
}
