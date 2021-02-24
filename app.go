package main

import (
	"YetAnotherBotStatsOld/go-config"
	"os"
)

type app struct {
	cfg  cfg
	file *os.File
}

type cfg struct {
	dir      string
	filePath string
	stage    int
}

func (a app) Start() {
	var err error

	a.file, err = os.Open(a.cfg.filePath)
	pnc(err)
	defer a.file.Close()

	switch a.cfg.stage {
	case 0: // тесты
		a.stageTest()
	}
}

func pnc(err error) {
	if err != nil {
		panic(err)
	}
}

func NewCFG() (cfg cfg) {
	configObj := config.New()

	cfg.dir = configObj.GetString("DIR")
	cfg.filePath = configObj.GetString("FILE")
	cfg.stage = configObj.GetInt("STAGE")

	return
}
