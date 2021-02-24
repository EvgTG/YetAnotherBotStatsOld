package main

import (
	"fmt"
	"os"
	"time"
)

type app struct {
	cfg  cfg
	file *os.File
	rgx  *rgx
}

func (a app) Start() {
	var err error

	a.file, err = os.Open(a.cfg.filePath)
	pnc(err)
	defer a.file.Close()

	tm := time.Now()

	switch a.cfg.stage {
	case 0: // тесты
		a.stageTest()
	case 1: // Голосования
		a.stage1()
	}

	fmt.Println("Обработано за ", time.Since(tm).String())
}
