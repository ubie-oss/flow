package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sakajunquality/flow/flow"
	"gopkg.in/yaml.v2"
)

var f *flow.Flow

func main() {

	config := flag.String("config", "config.yaml", "config file")
	flag.Parse()
	yamlFile, err := ioutil.ReadFile(*config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ioutil.ReadFile error:%v.\n", err)
		os.Exit(1)
	}

	cfg := new(flow.Config)
	if err := yaml.Unmarshal(yamlFile, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "yaml.Unmarshal error:%v.\n", err)
		os.Exit(1)
	}

	f, err = flow.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "flow init error:%v.\n", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	ctx := context.TODO()
	f.Start(ctx, errCh)
	err = <-errCh
	f.Stop(ctx)
}
