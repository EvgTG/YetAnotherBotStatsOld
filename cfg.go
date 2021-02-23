package main

import "YetAnotherBotStatsOld/go-config"

type cfg struct {
	dir      string
	filePath string
	stage    int
}

func NewCFG() (cfg cfg) {
	configObj := config.New()

	cfg.dir = configObj.GetString("DIR")
	cfg.filePath = configObj.GetString("FILE")
	cfg.stage = configObj.GetInt("FILE")

	return
}
