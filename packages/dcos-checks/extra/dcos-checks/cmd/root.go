// Copyright © 2017 Mesosphere Inc. <http://mesosphere.com>
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

package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/dcos/dcos-go/dcos"
	"github.com/dcos/dcos/packages/dcos-checks/extra/dcos-checks/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	statusOK      = 0
	statusWarning = 1
	statusFailure = 2
	statusUnknown = 3
)

var (
	// DCOSConfig is a global variable contains CLI options.
	DCOSConfig = new(CLIConfigFlags)

	cfgFile    string
	httpClient *http.Client
)

// ErrInvalidConfig is a generic error structure thrown by CLIConfigFlags receiver functions
// on config error.
type ErrInvalidConfig string

// Error is used to implement error interface.
func (e ErrInvalidConfig) Error() string {
	return string(e)
}

// Msg is a helper method that return a custom error message passed as "s".
func (e ErrInvalidConfig) Msg(s string) error {
	e = ErrInvalidConfig(s)
	return e
}

// CLIConfigFlags consolidates CLI cobra flags
type CLIConfigFlags struct {
	// CACert is a path to DC/OS CA authority file.
	CACert string

	// Verbose enabled debugging output with logrus.Debug(...)
	Verbose bool

	// ForceTLS forces to use HTTPS over HTTP schema.
	ForceTLS bool

	// IAMConfig is a path to identity and access managment config.
	IAMConfig string

	// Role defines DC/OS node's role. Valid roles are: master, agent, agent_public
	// defined in "github.com/dcos/dcos-go/dcos" package.
	Role string

	// DetectIP is a path to detect_ip script. Usually must be /opt/mesosphere/bin/detect_ip
	DetectIP string

	// NodeIPStr describes an IP address. This option will override the output of DetectIP.
	NodeIPStr string
}

// IP returns a valid IP address. If NodeIPStr is set, it will be used. Otherwise DetectIP will be executed
// and output will be returned.
func (cli *CLIConfigFlags) IP(c *http.Client) (net.IP, error) {
	if cli.NodeIPStr != "" {
		ip := net.ParseIP(cli.NodeIPStr)
		if ip == nil {
			var err ErrInvalidConfig
			err.Msg("NodeIPStr has invalid IP address: " + cli.NodeIPStr)
			return nil, err
		}
		return ip, nil
	}

	// NodeIPStr is empty at this point. Now execute a command DetectIP variable.
	nodeInfo, err := client.NewNodeInfo(c, cli.Role, cli.ForceTLS)
	if err != nil {
		return nil, err
	}

	return nodeInfo.DetectIP()
}

func (cli *CLIConfigFlags) validateRequiredFlags(c *http.Client) error {
	if cli.Role != dcos.RoleMaster && cli.Role != dcos.RoleAgent && cli.Role != dcos.RoleAgentPublic {
		var err ErrInvalidConfig
		return err.Msg("--role must be one of master, agent, agent_public: " + cli.Role)
	}

	ip, err := cli.IP(c)
	if err != nil {
		var e ErrInvalidConfig
		return e.Msg("unable to get node's IP. Make sure to use --detect-ip or --node-ip option. " + err.Error())
	}

	logrus.Debugf("using node's IP address %s", ip)
	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "checks <check name> [parameters]",
	Short: "DC/OS health checks",
	Long: `DC/OS checks provides an easy interface to check the DC/OS components health

The checks could be executed against a signle node, or a whole cluster.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if DCOSConfig.Verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// init http client
		var err error
		httpClient, err = client.NewClient(nil, DCOSConfig.IAMConfig, DCOSConfig.CACert)
		if err != nil {
			logrus.Fatalf("Unable to initialize http client: %s", err)
		}

		if err := DCOSConfig.validateRequiredFlags(httpClient); err != nil {
			logrus.Fatal(err)
		}

	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// run the commands parser
	if err := RootCmd.Execute(); err != nil {
		logrus.Fatalf("Error parsing subcommands: %s", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.checks.yaml)")

	RootCmd.PersistentFlags().BoolVar(&DCOSConfig.ForceTLS, "force-tls", false, "use HTTPS for GET/POST requests")
	RootCmd.PersistentFlags().BoolVar(&DCOSConfig.Verbose, "verbose", false, "enable verbose output")
	RootCmd.PersistentFlags().StringVar(&DCOSConfig.Role, "role", "", "set DC/OS role. (valid roles: master, agent, public-agent)")
	RootCmd.PersistentFlags().StringVar(&DCOSConfig.IAMConfig, "iam-config", "", "a path to identity and access managment config")
	RootCmd.PersistentFlags().StringVar(&DCOSConfig.CACert, "ca-cert", "", "a path to certificate authority file")
	RootCmd.PersistentFlags().StringVar(&DCOSConfig.DetectIP, "detect-ip", "/opt/mesosphere/bin/detect_ip", "a path to detect ip script")
	RootCmd.PersistentFlags().StringVar(&DCOSConfig.NodeIPStr, "node-ip", "", "set node IP address overriding detect_ip output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".checks") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// DCOSChecker defines an interface for a generic DC/OS check.
// ID() returns a check unique ID and RunCheck(...) returns a combined stdout/stderr, exit code and error.
type DCOSChecker interface {
	ID() string
	Run(context.Context, *CLIConfigFlags) (string, int, error)
}

// RunCheck is a helper function that takes a list of DC/OS checks and runs checks one by one.
func RunCheck(checks []DCOSChecker) {

	exitCode := statusOK

	for _, check := range checks {
		output, retCode, err := check.Run(nil, DCOSConfig)
		if err != nil {
			logrus.Fatalf("Error executing %s: %s", check.ID(), err)
		}
		if retCode != statusOK {
			exitCode = statusFailure
		}
		if output != "" {
			fmt.Printf("[%s]: %s\n", check.ID(), output)
		}
	}
	os.Exit(exitCode)
}
