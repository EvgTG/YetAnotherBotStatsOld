package main

import (
	"YetAnotherBotStatsOld/go-config"
	"regexp"
)

type rgx struct {
	rgxVoteClose *regexp.Regexp
}

func NewRegexp() *rgx {
	return &rgx{
		rgxVoteClose: regexp.MustCompile(`^\[Bot\] Результаты голосования за вопрос`),
	}
}

type cfg struct {
	dir      string
	filePath string
	stage    int
}

func NewCFG() (cfg cfg) {
	configObj := config.New()

	cfg.dir = configObj.GetString("DIR")
	cfg.filePath = configObj.GetString("FILE")
	cfg.stage = configObj.GetInt("STAGE")

	return
}

func pnc(err error) {
	if err != nil {
		panic(err)
	}
}
