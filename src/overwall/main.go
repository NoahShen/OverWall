package main

import (
	l4g "code.google.com/p/log4go"
	"encoding/json"
	"flag"
	"news"
	"os"
	"runtime"
	"time"
	"web"
)

var (
	//logConfig  = flag.String("log-config", "", "path of log config file")
	logConfig = "../../config/logConfig.xml"
	//configPath = flag.String("config-path", "", "path of config file")
	configPath = "../../config/overwall.config"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	l4g.LoadConfiguration(logConfig)
	l4g.Debug("MAXPROCS: %d", runtime.GOMAXPROCS(0))
	defer time.Sleep(2 * time.Second) // make sure log4go output the log

	configFile, readConfigFileErr := os.Open(configPath)
	if readConfigFileErr != nil {
		l4g.Error("Read config error: %s", readConfigFileErr.Error())
		return
	}
	decoder := json.NewDecoder(configFile)
	config := &web.Config{}
	decoder.Decode(&config)

	l4g.Info("Config info: %+v", config)

	opt := news.Option{}
	opt.SpeechFileDir = config.SpeechFileDir
	opt.SpeechCache = config.CacheSize
	opt.SentenceLen = config.OneSentenceLen
	opt.MaxGenVoiceTask = config.MaxGenVoiceTask
	opt.EasyreadUsername = config.EasyreadUsername
	opt.EasyreadPwd = config.EasyreadPwd
	opt.GenVoiceFile = true

	newsManager := news.NewNewsManager(opt)
	webManager := web.NewWebManager(newsManager, "8787", config)
	webManager.StartServer()
}
