package main

import (
	"fmt"
	"os"
	"strconv"

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
	file := a.createFileNTrunc("Polls/VotingResults.txt")
	defer file.Close()

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

	type pollsStat struct {
		total                                   int
		totalVotes, averageVotes, creatorsVotes int
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
		strSplit      []string
		polls         []poll
		pl            poll
		plStat        pollsStat
		creatorsMap   = make(map[string]int, 0)
		creatorsSlice = make([]kv, 0)
		file          *os.File
	)

	//чтение файла
	strSplit = strings.Split(a.readFile("Polls/VotingResults.txt"), "--------------------------------------------------\n")
	strSplit = strSplit[1:]
	polls = make([]poll, 0, len(strSplit))

	//заполнение списка опросов и статистики
	for i := range strSplit {
		pl = unmarshalPoll(strSplit[i])

		plStat.totalVotes += pl.usersN
		creatorsMap[pl.creator]++

		polls = append(polls, pl)
	}

	//топ создателей
	creatorsMap = mapNickTransformation(creatorsMap)
	file = a.createFileNTrunc("Polls/PollsCreators.txt")
	creatorsSlice = mapSort(creatorsMap)
	for _, v := range creatorsSlice {
		file.WriteString(fmt.Sprintf("%3v %v\n", v.Value, v.Key))
	}
	file.Close()

	//общая статистика
	plStat.total = len(polls)
	plStat.averageVotes = plStat.totalVotes / plStat.total
	plStat.creatorsVotes = len(creatorsSlice)

	fmt.Printf(
		""+
			"Всего опросов:%16v\n"+
			"Всего голосов:%16v\n"+
			"Голосов на опрос в среднем:%3v\n"+
			"Создателей опросов:%11v\n\n",
		plStat.total, plStat.totalVotes, plStat.averageVotes, plStat.creatorsVotes,
	)
}
