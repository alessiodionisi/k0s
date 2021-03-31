/*
Copyright 2021 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"github.com/spf13/pflag"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/k0sproject/k0s/pkg/apis/v1beta1"
	"github.com/k0sproject/k0s/pkg/constant"
)

var (
	CfgFile        string
	DataDir        string
	Debug          bool
	DebugListenOn  string
	K0sVars        constant.CfgVars
	workerOpts     WorkerOptions
	controllerOpts ControllerOptions
)

// This struct holds all the CLI options & settings required by the
// different k0s sub-commands
type CLIOptions struct {
	WorkerOptions
	ControllerOptions
	CfgFile          string
	ClusterConfig    *v1beta1.ClusterConfig
	Debug            bool
	DebugListenOn    string
	DefaultLogLevels map[string]string
	K0sVars          constant.CfgVars
	KubeClient       k8s.Interface
	Logging          map[string]string // merged outcome of default log levels and cmdLoglevels
}

// Shared controller cli flags
type ControllerOptions struct {
	EnableWorker bool
	SingleNode   bool
}

// Shared worker cli flags
type WorkerOptions struct {
	APIServer        string
	CIDRRange        string
	CloudProvider    bool
	ClusterDNS       string
	CmdLogLevels     map[string]string
	CriSocket        string
	KubeletExtraArgs string
	Labels           []string
	TokenFile        string
	TokenArg         string
	WorkerProfile    string
}

func DefaultLogLevels() map[string]string {
	return map[string]string{
		"etcd":                    "info",
		"containerd":              "info",
		"konnectivity-server":     "1",
		"kube-apiserver":          "1",
		"kube-controller-manager": "1",
		"kube-scheduler":          "1",
		"kubelet":                 "1",
		"kube-proxy":              "1",
	}
}

func GetPersistentFlagSet() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}
	flagset.StringVarP(&CfgFile, "config", "c", "", "config file (default: ./k0s.yaml)")
	flagset.BoolVarP(&Debug, "debug", "d", false, "Debug logging (default: false)")
	flagset.StringVar(&DataDir, "data-dir", "", "Data Directory for k0s (default: /var/lib/k0s). DO NOT CHANGE for an existing setup, things will break!")
	flagset.StringVar(&DebugListenOn, "debugListenOn", ":6060", "Http listenOn for Debug pprof handler")
	return flagset
}

func GetWorkerFlags() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}

	flagset.StringVar(&workerOpts.WorkerProfile, "profile", "default", "worker profile to use on the node")
	flagset.StringVar(&workerOpts.CriSocket, "cri-socket", "", "contrainer runtime socket to use, default to internal containerd. Format: [remote|docker]:[path-to-socket]")
	flagset.StringVar(&workerOpts.APIServer, "api-server", "", "HACK: api-server for the windows worker node")
	flagset.StringVar(&workerOpts.CIDRRange, "cidr-range", "10.96.0.0/12", "HACK: cidr range for the windows worker node")
	flagset.StringVar(&workerOpts.ClusterDNS, "cluster-dns", "10.96.0.10", "HACK: cluster dns for the windows worker node")
	flagset.BoolVar(&workerOpts.CloudProvider, "enable-cloud-provider", false, "Whether or not to enable cloud provider support in kubelet")
	flagset.StringVar(&workerOpts.TokenFile, "token-file", "", "Path to the file containing token.")
	flagset.StringToStringVarP(&workerOpts.CmdLogLevels, "logging", "l", DefaultLogLevels(), "Logging Levels for the different components")
	flagset.StringSliceVarP(&workerOpts.Labels, "labels", "", []string{}, "Node labels, list of key=value pairs")
	flagset.StringVar(&workerOpts.KubeletExtraArgs, "kubelet-extra-args", "", "extra args for kubelet")
	return flagset
}

func GetControllerFlags() *pflag.FlagSet {
	flagset := &pflag.FlagSet{}

	flagset.StringVar(&workerOpts.WorkerProfile, "profile", "default", "worker profile to use on the node")
	flagset.BoolVar(&controllerOpts.EnableWorker, "enable-worker", false, "enable worker (default false)")
	flagset.StringVar(&workerOpts.TokenFile, "token-file", "", "Path to the file containing join-token.")
	flagset.StringVar(&workerOpts.CriSocket, "cri-socket", "", "contrainer runtime socket to use, default to internal containerd. Format: [remote|docker]:[path-to-socket]")
	flagset.StringToStringVarP(&workerOpts.CmdLogLevels, "logging", "l", DefaultLogLevels(), "Logging Levels for the different components")
	flagset.BoolVar(&controllerOpts.SingleNode, "single", false, "enable single node (implies --enable-worker, default false)")

	return flagset
}

func GetCmdOpts() CLIOptions {
	K0sVars = constant.GetConfig(DataDir)

	opts := CLIOptions{
		ControllerOptions: controllerOpts,
		WorkerOptions:     workerOpts,

		CfgFile:          CfgFile,
		Debug:            Debug,
		DefaultLogLevels: DefaultLogLevels(),
		K0sVars:          K0sVars,
		DebugListenOn:    DebugListenOn,
	}
	return opts
}
