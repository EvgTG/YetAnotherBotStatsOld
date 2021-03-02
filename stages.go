package main

import (
	"fmt"
	"github.com/vdobler/chart"
	"os"
	"sort"
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
	file1 := a.createFileNTrunc("Polls/PollsResults.txt")
	defer file1.Close()
	file2 := a.createFileNTrunc("Polls/IPollsResults.txt")
	defer file2.Close()

	for msg := range a.unmarshalChan() {
		if a.rgx.rgxPollClose.MatchString(msg.Text) {
			file1.WriteString(fmt.Sprintf("--------------------------------------------------\n%v\n%v\n", msg.Date.Format("2006.01.02 15:04"), msg.Text))
		}
		if a.rgx.rgxIPoll.MatchString(msg.Text) {
			file2.WriteString(fmt.Sprintf("--------------------------------------------------\n%v\n%v\n", msg.Date.Format("2006.01.02 15:04"),
				strings.Replace(msg.Text, "✔", "", 1)))
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
		raw      string
	}

	type ipoll struct {
		date     time.Time
		creator  string
		question string
		variants string
		raw      string
	}

	type pollsStat struct {
		totalP, totalIP                    int
		totalVotes, maxVotes, averageVotes int
		creatorsIncP, creatorsIncIP        int
		rocketInc                          int
	}

	//rgx poll
	rgxDate := regexp.MustCompile(`^[0-9]{4}\.[0-9]{2}\.[0-9]{2} [0-9]{2}\:[0-9]{2}`)
	rgxCreater := regexp.MustCompile(`^\n\[Bot\] Результаты голосования за вопрос #.+:\n`)
	rgxQuestion := regexp.MustCompile(`\n\n(.+голос\(а\), .+|.+%\(\d+\) - .+)\n`)
	rgxResults := regexp.MustCompile(`\n\n`)
	rgxUsersN := regexp.MustCompile(`^Всего проголосовало(|:) [0-9]{1,3}( юзер\(ов\)\.|)\n$`)
	rgxRocket := regexp.MustCompile(`(Р|р)акет`)
	//rgx ipoll
	rgxCreaterIP := regexp.MustCompile(`^\n\[Bot\] #.+ поставил\(а\) вопрос:\n`)
	rgxQuestionIP := regexp.MustCompile(`\n\n1️⃣ - .*\n`)

	nickCut := strings.NewReplacer("\n[Bot] Результаты голосования за вопрос ", "", "`", "", ":\n", "")
	nickCutIP := strings.NewReplacer("\n[Bot] ", "", " поставил(а) вопрос", "", "`", "", ":\n", "")
	usersNCut := strings.NewReplacer("Всего проголосовало", "", ":", "", " ", "", "юзер(ов)", "", ".", "", "\n", "")

	unmarshalPoll := func(str string) (p poll) {
		p.raw = str
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

	unmarshalIPoll := func(str string) (p ipoll) {
		p.raw = str
		var err error
		var index int

		p.date, err = time.Parse("2006.01.02 15:04", rgxDate.FindString(str))
		pnc(err)
		str = rgxDate.ReplaceAllLiteralString(str, "")

		p.creator = nickCutIP.Replace(rgxCreaterIP.FindString(str))
		str = rgxCreaterIP.ReplaceAllLiteralString(str, "")

		index = rgxQuestionIP.FindIndex([]byte(str))[0]
		p.question = str[:index]
		str = str[index+2:]

		p.variants = str

		return
	}

	var (
		strFileP, strFileIP   string
		strSplitP, strSplitIP []string
		polls                 []poll
		ipolls                []ipoll
		pl                    poll
		ipl                   ipoll
		plStat                pollsStat
		creatorsMapP          = make(map[string]int, 0)
		creatorsMapIP         = make(map[string]int, 0)
		creatorsSliceP        = make([]kv, 0)
		creatorsSliceIP       = make([]kv, 0)
		file                  *os.File
		chatWords             = []string{" чат", "YetAnotherBot"}
		chatWordsStr          string
	)

	//чтение файлов
	strFileP = a.readFile("Polls/PollsResults.txt")
	strSplitP = strings.Split(strFileP, "--------------------------------------------------\n")
	strSplitP = strSplitP[1:]
	polls = make([]poll, 0, len(strSplitP))

	strFileIP = a.readFile("Polls/IPollsResults.txt")
	strSplitIP = strings.Split(strFileIP, "--------------------------------------------------\n")
	strSplitIP = strSplitIP[1:]
	ipolls = make([]ipoll, 0, len(strSplitIP))

	//заполнение списка опросов и статистики
	for i := range strSplitP {
		pl = unmarshalPoll(strSplitP[i])

		plStat.totalVotes += pl.usersN
		creatorsMapP[pl.creator]++

		if inArray(pl.question, chatWords) {
			chatWordsStr += fmt.Sprintf("--------------------------------------------------\n%v", pl.raw)
		}

		polls = append(polls, pl)
	}

	for i := range strSplitIP {
		ipl = unmarshalIPoll(strSplitIP[i])

		creatorsMapIP[ipl.creator]++

		ipolls = append(ipolls, ipl)
	}

	//топ создателей
	creatorsMapP = mapNickTransformation(creatorsMapP)
	file = a.createFileNTrunc("Polls/PollsCreators.txt")
	creatorsSliceP = mapSort(creatorsMapP)
	for _, v := range creatorsSliceP {
		file.WriteString(fmt.Sprintf("%3v %v\n", v.Value, v.Key))
	}
	file.Close()

	creatorsMapIP = mapNickTransformation(creatorsMapIP)
	file = a.createFileNTrunc("Polls/IPollsCreators.txt")
	creatorsSliceIP = mapSort(creatorsMapIP)
	for _, v := range creatorsSliceIP {
		file.WriteString(fmt.Sprintf("%3v %v\n", v.Value, v.Key))
	}
	file.Close()

	//графики
	{
		dumper := a.NewDumper("Polls/Charts", 1, 3, 2000, 500)
		defer dumper.Close()

		var (
			dataC1 = xy{make([]float64, 0), make([]float64, 0)}

			dataC2P       = xy{make([]float64, 0), make([]float64, 0)}
			dataC2IP      = xy{make([]float64, 0), make([]float64, 0)}
			dataC2mapP    = make(map[string]int, 0)
			dataC2mapIP   = make(map[string]int, 0)
			dataC2sliceP  = make([]kv, 0)
			dataC2sliceIP = make([]kv, 0)

			dataC3      = xy{make([]float64, 0), make([]float64, 0)}
			dataC3map   = make(map[string]int, 0)
			dataC3slice = make([]kv, 0)
			charts      = make([]chart.Chart, 0, dumper.N*dumper.M)
		)

		for _, pl := range polls {
			dataC1.x = append(dataC1.x, float64(pl.date.Unix()))
			dataC1.y = append(dataC1.y, float64(pl.usersN))

			dataC2mapP[pl.date.Format("2006-01")]++

			dataC3map[pl.date.Format("2006-01")] += pl.usersN
		}
		for _, ipl := range ipolls {
			dataC2mapIP[ipl.date.Format("2006-01")]++
		}

		//время к колву проголосовавших (точки)
		c1 := &chart.ScatterChart{}
		c1.Title = "Все опросы"
		c1.Key.Hide = true
		c1.XRange.Time = true
		c1.YRange.Label = "Кол-во голосов"
		c1.AddDataPair("polls", dataC1.x, dataC1.y, chart.PlotStylePoints,
			chart.Style{FillColor: ParseHexColor("#FF0000"), Symbol: 'o', LineWidth: 2})
		charts = append(charts, c1)

		//время к колву опросов в месяц (столбцы)
		dataC2sliceP = mapSortByTime(dataC2mapP, "2006-01")
		dataC2sliceIP = mapSortByTime(dataC2mapIP, "2006-01")
		for _, v := range dataC2sliceP {
			tm, _ := time.Parse("2006-01", v.Key)
			dataC2P.x = append(dataC2P.x, float64(tm.Unix()))
			dataC2P.y = append(dataC2P.y, float64(v.Value))
		}
		for _, v := range dataC2sliceIP {
			tm, _ := time.Parse("2006-01", v.Key)
			dataC2IP.x = append(dataC2IP.x, float64(tm.Unix()))
			dataC2IP.y = append(dataC2IP.y, float64(v.Value))
		}
		c2 := &chart.BarChart{}
		c2.Title = "Кол-во опросов в месяц"
		c2.Key.Pos = "itl"
		c2.XRange.Time = true
		c2.YRange.Label = "Кол-во опросов"
		c2.AddDataPair("Опросы", dataC2P.x, dataC2P.y, chart.Style{LineColor: ParseHexColor("#FF0000"), LineWidth: 2, FillColor: ParseHexColor("#FF0000")})
		c2.AddDataPair("Инлайн опросы", dataC2IP.x, dataC2IP.y, chart.Style{LineColor: ParseHexColor("#0000FF"), LineWidth: 2, FillColor: ParseHexColor("#0000FF")})
		charts = append(charts, c2)

		//время к колву голосов в месяц (столбцы)
		dataC3slice = mapSortByTime(dataC3map, "2006-01")
		for _, v := range dataC3slice {
			tm, _ := time.Parse("2006-01", v.Key)
			dataC3.x = append(dataC3.x, float64(tm.Unix()))
			dataC3.y = append(dataC3.y, float64(v.Value))
		}
		c3 := &chart.BarChart{}
		c3.Title = "Кол-во голосов в месяц"
		c3.Key.Hide = true
		c3.XRange.Time = true
		c3.YRange.Label = "Кол-во голосов"
		c3.AddDataPair("polls", dataC3.x, dataC3.y, chart.Style{LineColor: ParseHexColor("#FF0000"), LineWidth: 3, FillColor: ParseHexColor("#FF0000")})
		charts = append(charts, c3)

		//рисовка
		for _, c := range charts {
			dumper.Plot(c)
			c.Reset()
		}
	}

	//топ30 голосований по количеству проголосовавших
	file = a.createFileNTrunc("Polls/PollsTop30.txt")
	sort.Slice(polls, func(i, j int) bool {
		return polls[i].usersN > polls[j].usersN
	})
	for i := 0; i < 30; i++ {
		file.WriteString(fmt.Sprintf("--------------------------------------------------\n%v. %v\n", i+1, polls[i].raw))
	}
	file.Close()

	//вопросы про чатик
	file = a.createFileNTrunc("Polls/Chat.txt")
	file.WriteString(chatWordsStr)
	file.Close()

	//общая статистика
	plStat.totalP = len(polls)
	plStat.totalIP = len(ipolls)
	plStat.averageVotes = plStat.totalVotes / plStat.totalP
	plStat.creatorsIncP = len(creatorsSliceP)
	plStat.creatorsIncIP = len(creatorsSliceIP)
	plStat.rocketInc = len(rgxRocket.FindAllString(strFileP, -1)) + len(rgxRocket.FindAllString(strFileIP, -1))
	plStat.maxVotes = polls[0].usersN

	fmt.Printf(
		""+
			"Всего опросов:%26v\n"+
			"Всего голосов:%26v\n"+
			"Максимум голосов в опросе:%14v\n"+
			"Голосов на опрос в среднем:%13v\n"+
			"Создателей опросов:%21v\n\n"+
			"Всего инлайн опросов:%19v\n"+
			"Создателей инлайн опросов:%14v\n\n"+
			"Упоминаний Ракеты:%22v",
		plStat.totalP,
		plStat.totalVotes,
		plStat.maxVotes,
		plStat.averageVotes,
		plStat.creatorsIncP,
		plStat.totalIP,
		plStat.creatorsIncIP,
		plStat.rocketInc,
	)
}
