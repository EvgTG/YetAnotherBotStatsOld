package main

import (
	"fmt"
	"strconv"

	//"github.com/vdobler/chart"
	"os"
	"regexp"
	"strings"
	"time"
)

func (a app) stageTest() {
	for msg := range a.unmarshalChan() {
		fmt.Println(msg)
	}
}

// Голосования
func (a app) stage1() {
	var err error

	file, err := os.OpenFile(a.cfg.dir+"Polls/VotingResults.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	pnc(err)
	defer file.Close()
	stat, err := file.Stat()
	pnc(err)

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

// Обработка голосований
func (a app) stage2() {
	type poll struct {
		date     time.Time
		creator  string
		question string
		results  string
		usersN   int
	}

	rgxDate := regexp.MustCompile(`^[0-9]{4}\.[0-9]{2}\.[0-9]{2} [0-9]{2}\:[0-9]{2}`)
	rgxCreater := regexp.MustCompile(`^\n\[Bot\] Результаты голосования за вопрос #.+:\n`)
	rgxQuestion := regexp.MustCompile(`\n\n(.+голос\(а\), .+|.+%\(\d+\) - .+)\n`)
	rgxResults := regexp.MustCompile(`\n\n`)
	rgxUsersN := regexp.MustCompile(`^Всего проголосовало(|:) [0-9]{1,3}( юзер\(ов\)\.|)\n$`)
	//rgx := regexp.MustCompile(``)

	nickCut := strings.NewReplacer("\n[Bot] Результаты голосования за вопрос ", "", "`", "", ":\n", "")
	usersNCut := strings.NewReplacer("Всего проголосовало", "", ":", "", " ", "", "юзер(ов)", "", ".", "", "\n", "")

	unmarshalPoll := func(str string) (p poll) {
		var err error
		var index int

		p.date, err = time.Parse("2006.01.02 15:04", rgxDate.FindString(str))
		pnc(err)
		str = rgxDate.ReplaceAllLiteralString(str, "")

		p.creator = nickCut.Replace(rgxCreater.FindString(str))
		str = rgxCreater.ReplaceAllLiteralString(str, "")

		index = rgxQuestion.FindIndex([]byte(str))[0]
		p.question = str[:index]
		str = str[index+2:]

		index = rgxResults.FindIndex([]byte(str))[0]
		p.results = str[:index]
		str = str[index+2:]

		p.usersN, err = strconv.Atoi(usersNCut.Replace(rgxUsersN.FindString(str)))
		pnc(err)

		return
	}

	var (
		err      error
		strSplit []string
		strBytes []byte
		polls    []poll
	)

	file, err := os.OpenFile(a.cfg.dir+"Polls/VotingResults.txt", os.O_RDWR, 0666)
	pnc(err)
	defer file.Close()
	stat, err := file.Stat()
	pnc(err)

	stat.Size()
	strBytes = make([]byte, stat.Size())
	_, err = file.Read(strBytes)
	strSplit = strings.Split(string(strBytes), "--------------------------------------------------\n")
	strSplit = strSplit[1:]
	polls = make([]poll, 0, len(strSplit))

	for i := range strSplit {
		polls = append(polls, unmarshalPoll(strSplit[i]))
	}
}
