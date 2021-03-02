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
	file := a.createFileNTrunc("Polls/VotingResults.txt")
	defer file.Close()

	for msg := range a.unmarshalChan() {
		if a.rgx.rgxPollClose.MatchString(msg.Text) {
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
		raw      string
	}

	type pollsStat struct {
		total                              int
		totalVotes, maxVotes, averageVotes int
		creatorsInc                        int
		rocketInc                          int
	}

	rgxDate := regexp.MustCompile(`^[0-9]{4}\.[0-9]{2}\.[0-9]{2} [0-9]{2}\:[0-9]{2}`)
	rgxCreater := regexp.MustCompile(`^\n\[Bot\] Результаты голосования за вопрос #.+:\n`)
	rgxQuestion := regexp.MustCompile(`\n\n(.+голос\(а\), .+|.+%\(\d+\) - .+)\n`)
	rgxResults := regexp.MustCompile(`\n\n`)
	rgxUsersN := regexp.MustCompile(`^Всего проголосовало(|:) [0-9]{1,3}( юзер\(ов\)\.|)\n$`)
	rgxRocket := regexp.MustCompile(`(Р|р)акет`)
	//rgx := regexp.MustCompile(``)

	nickCut := strings.NewReplacer("\n[Bot] Результаты голосования за вопрос ", "", "`", "", ":\n", "")
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

	var (
		strFile       string
		strSplit      []string
		polls         []poll
		pl            poll
		plStat        pollsStat
		creatorsMap   = make(map[string]int, 0)
		creatorsSlice = make([]kv, 0)
		file          *os.File
		chatWords     = []string{" чат", "YetAnotherBot"}
		chatWordsStr  string
	)

	//чтение файла
	strFile = a.readFile("Polls/VotingResults.txt")
	strSplit = strings.Split(strFile, "--------------------------------------------------\n")
	strSplit = strSplit[1:]
	polls = make([]poll, 0, len(strSplit))

	//заполнение списка опросов и статистики
	for i := range strSplit {
		pl = unmarshalPoll(strSplit[i])

		plStat.totalVotes += pl.usersN
		creatorsMap[pl.creator]++

		if inArray(pl.question, chatWords) {
			chatWordsStr += fmt.Sprintf("--------------------------------------------------\n%v", pl.raw)
		}

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

	//графики
	{
		dumper := a.NewDumper("Polls/Charts", 1, 3, 2000, 500)
		defer dumper.Close()

		var (
			dataC1      = xy{make([]float64, 0), make([]float64, 0)}
			dataC2      = xy{make([]float64, 0), make([]float64, 0)}
			dataC3      = xy{make([]float64, 0), make([]float64, 0)}
			dataC2map   = make(map[string]int, 0)
			dataC2slice = make([]kv, 0)
			dataC3map   = make(map[string]int, 0)
			dataC3slice = make([]kv, 0)
			charts      = make([]chart.Chart, 0, dumper.N*dumper.M)
		)

		for _, pl := range polls {
			dataC1.x = append(dataC1.x, float64(pl.date.Unix()))
			dataC1.y = append(dataC1.y, float64(pl.usersN))

			dataC2map[pl.date.Format("2006-01")]++
			dataC3map[pl.date.Format("2006-01")] += pl.usersN
		}

		//время к колву проголосовавших (точки)
		c1 := &chart.ScatterChart{}
		c1.Title = "Все опросы"
		c1.Key.Hide = true
		c1.XRange.Time = true
		c1.YRange.Label = "Кол-во голосов"
		c1.AddDataPair("polls", dataC1.x, dataC1.y, chart.PlotStylePoints,
			chart.Style{FillColor: ParseHexColor("#FF5733"), Symbol: 'o', LineWidth: 2})
		charts = append(charts, c1)

		//время к колву опросов (по месяцам)
		dataC2slice = mapSortByTime(dataC2map, "2006-01")
		for _, v := range dataC2slice {
			tm, _ := time.Parse("2006-01", v.Key)
			dataC2.x = append(dataC2.x, float64(tm.Unix()))
			dataC2.y = append(dataC2.y, float64(v.Value))
		}
		c2 := &chart.BarChart{}
		c2.Title = "Кол-во опросов в месяц"
		c2.Key.Hide = true
		c2.XRange.Time = true
		c2.YRange.Label = "Кол-во опросов"
		c2.AddDataPair("polls", dataC2.x, dataC2.y, chart.Style{LineColor: ParseHexColor("#FF5733"), LineWidth: 3, FillColor: ParseHexColor("#FF5733")})
		charts = append(charts, c2)

		//время к колву голосов (по месяцам)
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
		c3.AddDataPair("polls", dataC3.x, dataC3.y, chart.Style{LineColor: ParseHexColor("#FF5733"), LineWidth: 3, FillColor: ParseHexColor("#FF5733")})
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
	plStat.total = len(polls)
	plStat.averageVotes = plStat.totalVotes / plStat.total
	plStat.creatorsInc = len(creatorsSlice)
	plStat.rocketInc = len(rgxRocket.FindAllString(strFile, -1))
	plStat.maxVotes = polls[0].usersN

	fmt.Printf(
		""+
			"Всего опросов:%26v\n"+
			"Всего голосов:%26v\n"+
			"Максимум голосов в опросе:%14v\n"+
			"Голосов на опрос в среднем:%13v\n"+
			"Создателей опросов:%21v\n"+
			"Упоминаний Ракеты:%22v",
		plStat.total, plStat.totalVotes, plStat.maxVotes, plStat.averageVotes, plStat.creatorsInc, plStat.rocketInc,
	)
}
