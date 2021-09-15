// Package cmd contains an entrypoint for running an ion-sfu instance.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	log "github.com/pion/ion-log"
	biz "github.com/pion/ion/apps/biz/server"
	"github.com/spf13/viper"
)

var (
	buildstamp = ""
	githash    = ""
	goversion  = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	conf = biz.Config{}
	file string
)

func version() {
	fmt.Printf("Git Commit Hash: %s\n", githash)
	fmt.Printf("UTC Build Time:  %s\n", buildstamp)
	fmt.Printf("Golang Version:  %s\n", goversion)
}

func unmarshal(rawVal interface{}) bool {
	if err := viper.Unmarshal(rawVal); err != nil {
		fmt.Printf("config file %s loaded failed. %v\n", file, err)
		return false
	}
	return true
}

func load() bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}

	viper.SetConfigFile(file)
	viper.SetConfigType("toml")

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("config file %s read failed. %v\n", file, err)
		return false
	}

	if !unmarshal(&conf) {
		return false
	}
	if err != nil {
		fmt.Printf("config file %s loaded failed. %v\n", file, err)
		return false
	}

	fmt.Printf("config %s load ok!\n", file)

	return true
}

func parse() bool {
	v := flag.Bool("v", false, "show version info")
	flag.StringVar(&file, "c", "configs/biz.toml", "config file")

	flag.Parse()
	if *v {
		version()
		return false
	}

	if !load() {
		return false
	}

	return true
}

func main() {
	if !parse() {
		os.Exit(-1)
	}
	log.Init(conf.Log.Level)
	log.Infof("--- Starting Biz Node ---")
	node := biz.NewBIZ(conf.Node.NID)
	if err := node.Start(conf); err != nil {
		log.Errorf("biz init start: %v", err)
		os.Exit(-1)
	}
	defer node.Close()
	select {}
}
