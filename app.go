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
	case 2: // Обработка голосований, только после case1
		a.stage2()
	case 3: // Круговая диаграмма создателей опросов, только после удаления ненужных ников из файла посредника
		a.stage3()
	}

	fmt.Println("\n\nОбработано за ", time.Since(tm).String())
}
