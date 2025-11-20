/*
 * Copyright 2025 - IBM Corporation. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/IBM/power-dra-driver/internal/info"
	"github.com/IBM/power-dra-driver/pkg/flags"
)

const (
	DriverName = "nx.device.power.ibm.com"

	PluginRegistrationPath     = "/var/lib/kubelet/plugins/" + DriverName + ".sock"
	DriverPluginPath           = "/var/lib/kubelet/plugins/" + DriverName
	DriverPluginCheckpointFile = "checkpoint.json"
)

type Flags struct {
	kubeClientConfig flags.KubeClientConfig
	loggingConfig    *flags.LoggingConfig

	nodeName   string
	cdiRoot    string
	numDevices int
}

type Config struct {
	flags      *Flags
	coreclient coreclientset.Interface
}

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	flags := &Flags{
		loggingConfig: flags.NewLoggingConfig(),
	}
	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "node-name",
			Usage:       "The name of the node to be worked on.",
			Required:    true,
			Destination: &flags.nodeName,
			EnvVars:     []string{"NODE_NAME"},
		},
		&cli.StringFlag{
			Name:        "cdi-root",
			Usage:       "Absolute path to the directory where CDI files will be generated.",
			Value:       "/etc/cdi",
			Destination: &flags.cdiRoot,
			EnvVars:     []string{"CDI_ROOT"},
		},
		&cli.IntFlag{
			Name:        "num-devices",
			Usage:       "The number of pseudo devices to be generated.",
			Value:       100,
			Destination: &flags.numDevices,
			EnvVars:     []string{"NUM_DEVICES"},
		},
	}
	cliFlags = append(cliFlags, flags.kubeClientConfig.Flags()...)
	cliFlags = append(cliFlags, flags.loggingConfig.Flags()...)

	app := &cli.App{
		Name:            "power-dra-kubeletplugin",
		Usage:           "power-dra-kubeletplugin implements a DRA driver plugin.",
		ArgsUsage:       " ",
		HideHelpCommand: true,
		Flags:           cliFlags,
		Before: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				return fmt.Errorf("arguments not supported: %v", c.Args().Slice())
			}
			if err := flags.loggingConfig.Apply(); err != nil {
				return err
			}
			klog.InfoS("Starting power-dra-kubeletplugin", "version", info.GetVersion(), "gitCommit", info.GetGitCommit())
			return nil
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context
			clientSets, err := flags.kubeClientConfig.NewClientSets()
			if err != nil {
				return fmt.Errorf("create client: %v", err)
			}

			config := &Config{
				flags:      flags,
				coreclient: clientSets.Core,
			}

			return StartPlugin(ctx, config)
		},
	}

	return app
}

func StartPlugin(ctx context.Context, config *Config) error {
	err := os.MkdirAll(DriverPluginPath, 0750)
	if err != nil {
		return err
	}

	info, err := os.Stat(config.flags.cdiRoot)
	switch {
	case err != nil && os.IsNotExist(err):
		err := os.MkdirAll(config.flags.cdiRoot, 0750)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("path for cdi file generation is not a directory: '%v'", err)
	}

	driver, err := NewDriver(ctx, config)
	if err != nil {
		return err
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigc

	err = driver.Shutdown(ctx)
	if err != nil {
		klog.FromContext(ctx).Error(err, "Unable to cleanly shutdown driver")
	}

	return nil
}
