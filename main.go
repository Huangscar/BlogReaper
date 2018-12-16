package main

import (
	"flag"
	"github.com/XMatrixStudio/BlogReaper/app"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func main() {
	// 加载配置文件
	configFile := flag.String("c", "config/config.yaml", "Where is your config file?")
	flag.Parse()
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Printf("Can't find the config file in %v", *configFile)
		return
	}
	log.Printf("Load the config file in %v", *configFile)
	conf := app.Config{}
	yaml.Unmarshal(data, &conf)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(conf.Server.Port), app.App(conf)))
}
