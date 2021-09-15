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
	"github.com/pion/ion/pkg/node/sfu"
	"github.com/spf13/viper"
)

const (
	portRangeLimit = 100
)

var (
	buildstamp = ""
	githash    = ""
	goversion  = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	conf = sfu.Config{}
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

	if !unmarshal(&conf) || !unmarshal(&conf.Config) {
		return false
	}

	if len(conf.WebRTC.ICEPortRange) > 2 {
		fmt.Printf("config file %s loaded failed. range port must be [min,max]\n", file)
		return false
	}

	if len(conf.WebRTC.ICEPortRange) != 0 && conf.WebRTC.ICEPortRange[1]-conf.WebRTC.ICEPortRange[0] < portRangeLimit {
		fmt.Printf("config file %s loaded failed. range port must be [min, max] and max - min >= %d\n", file, portRangeLimit)
		return false
	}

	fmt.Printf("config %s load ok!\n", file)
	return true
}

func parse() bool {
	v := flag.Bool("v", false, "show version info")
	flag.StringVar(&file, "c", "configs/sfu.toml", "config file")

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

	log.Init("info")

	log.Infof("--- starting sfu node ---")

	node := sfu.NewSFU(conf.Node.NID)
	if err := node.Start(conf); err != nil {
		log.Errorf("sfu init start: %v", err)
		os.Exit(-1)
	}
	defer node.Close()

	// Press Ctrl+C to exit the process
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
}
