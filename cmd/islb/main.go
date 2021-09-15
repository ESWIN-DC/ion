package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/pion/ion-log"
	"github.com/pion/ion/pkg/node/islb"
	"github.com/spf13/viper"
)

var (
	buildstamp = ""
	githash    = ""
	goversion  = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	conf = islb.Config{}
	file string
)

func version() {
	fmt.Printf("Git Commit Hash: %s\n", githash)
	fmt.Printf("UTC Build Time:  %s\n", buildstamp)
	fmt.Printf("Golang Version:  %s\n", goversion)
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
	err = viper.UnmarshalExact(&conf)
	if err != nil {
		fmt.Printf("config file %s loaded failed. %v\n", file, err)
		return false
	}
	fmt.Printf("config %s load ok!\n", file)
	return true
}

func parse() bool {
	v := flag.Bool("v", false, "show version info")
	flag.StringVar(&file, "c", "configs/islb.toml", "config file")
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

	log.Infof("--- starting islb node ---")

	node := islb.NewISLB(conf.Node.NID)
	if err := node.Start(conf); err != nil {
		log.Errorf("islb start error: %v", err)
		os.Exit(-1)
	}
	defer node.Close()

	// Press Ctrl+C to exit the process
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
}
