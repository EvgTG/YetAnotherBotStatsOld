package main

import (
	"fmt"
	"os"
)

func (a app) stageTest() {
	for msg := range a.unmarshalChan() {
		fmt.Println(msg)
	}
}

// Голосования
func (a app) stage1() {
	var err error

	file, _ := os.OpenFile(a.cfg.dir+"VotingResults.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	stat, err := file.Stat()
	pnc(err)
	defer file.Close()

	if stat.Size() > 0 {
		err = file.Truncate(0)
		pnc(err)
	}

	for msg := range a.unmarshalChan() {
		if a.rgx.rgxVoteClose.MatchString(msg.Text) {
			file.WriteString(fmt.Sprintf("--------------------------------------------------\n%v\n%v\n", msg.Date.Format("2006.01.02 15:04"), msg.Text))
		}
	}
}
