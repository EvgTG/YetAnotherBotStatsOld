package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/vdobler/chart"
	"image/color"
	"io"
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

		if inArrayContains(pl.question, chatWords) {
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
		dumper := a.NewDumper("Polls/Charts", 1, 5, 2000, 500)
		defer dumper.Close()

		var (
			chartsSlice = make([]chart.Chart, 0, dumper.N*dumper.M)
			colors      = []color.RGBA{ParseHexColor("#FF0000"), ParseHexColor("#0000FF")}

			dataP  = &xy{make([]float64, 0, 0), make([]float64, 0, 0)}
			dataIP = &xy{make([]float64, 0, 0), make([]float64, 0, 0)}
			//Не забывать обнулять - resetData()

			dataC2mapP  = make(map[string]int, 0)
			dataC2mapIP = make(map[string]int, 0)

			dataC3map = make(map[string]int, 0)

			dataC4mapP  = make(map[int]int, 0)
			dataC4mapIP = make(map[int]int, 0)

			dataC5mapP  = make(map[time.Weekday]int, 0)
			dataC5mapIP = make(map[time.Weekday]int, 0)
		)
		resetData := func() {
			dataP = &xy{make([]float64, 0, 0), make([]float64, 0, 0)}
			dataIP = &xy{make([]float64, 0, 0), make([]float64, 0, 0)}
		}

		for _, pl := range polls {
			dataP.x = append(dataP.x, float64(pl.date.Unix()))
			dataP.y = append(dataP.y, float64(pl.usersN))

			dataC2mapP[pl.date.Format("2006-01")]++

			dataC3map[pl.date.Format("2006-01")] += pl.usersN

			_, week := pl.date.ISOWeek()
			dataC4mapP[week]++

			dataC5mapP[pl.date.Weekday()+1]++
		}
		for _, ipl := range ipolls {
			dataC2mapIP[ipl.date.Format("2006-01")]++

			_, week := ipl.date.ISOWeek()
			dataC4mapIP[week]++

			dataC5mapIP[ipl.date.Weekday()+1]++
		}

		//время к колву проголосовавших (точки)
		c1 := &chart.ScatterChart{}
		c1.Title = "Все опросы"
		c1.Key.Hide = true
		c1.XRange.Time = true
		c1.YRange.Label = "Кол-во голосов"
		c1.AddDataPair("polls", dataP.x, dataP.y, chart.PlotStylePoints,
			chart.Style{FillColor: colors[0], Symbol: 'o', LineWidth: 2})
		chartsSlice = append(chartsSlice, c1)
		resetData()

		//время к колву опросов в месяц (столбцы)
		for k, v := range dataC2mapP {
			tm, _ := time.Parse("2006-01", k)
			dataP.x = append(dataP.x, float64(tm.Unix()))
			dataP.y = append(dataP.y, float64(v))
		}
		for k, v := range dataC2mapIP {
			tm, _ := time.Parse("2006-01", k)
			dataIP.x = append(dataIP.x, float64(tm.Unix()))
			dataIP.y = append(dataIP.y, float64(v))
		}
		c2 := &chart.BarChart{}
		c2.Title = "Кол-во опросов в месяц"
		c2.Key.Pos = "itl"
		c2.XRange.Time = true
		c2.YRange.Label = "Кол-во опросов"
		c2.AddDataPair("Опросы", dataP.x, dataP.y, chart.Style{LineColor: colors[0], LineWidth: 2, FillColor: colors[0]})
		c2.AddDataPair("Инлайн опросы", dataIP.x, dataIP.y, chart.Style{LineColor: colors[1], LineWidth: 2, FillColor: colors[1]})
		chartsSlice = append(chartsSlice, c2)
		resetData()

		//время к колву голосов в месяц (столбцы)
		for k, v := range dataC3map {
			tm, _ := time.Parse("2006-01", k)
			dataP.x = append(dataP.x, float64(tm.Unix()))
			dataP.y = append(dataP.y, float64(v))
		}
		c3 := &chart.BarChart{}
		c3.Title = "Кол-во голосов в месяц"
		c3.Key.Hide = true
		c3.XRange.Time = true
		c3.YRange.Label = "Кол-во голосов"
		c3.AddDataPair("polls", dataP.x, dataP.y, chart.Style{LineColor: colors[0], LineWidth: 3, FillColor: colors[0]})
		chartsSlice = append(chartsSlice, c3)
		resetData()

		//активность в год (столбцы)
		for k, v := range dataC4mapP {
			dataP.x = append(dataP.x, float64(k))
			dataP.y = append(dataP.y, float64(v))
		}
		for k, v := range dataC4mapIP {
			dataIP.x = append(dataIP.x, float64(k))
			dataIP.y = append(dataIP.y, float64(v))
		}
		c4 := &chart.BarChart{}
		c4.Title = "Годовая активность (не средняя, а суммарная), столбик - неделя"
		c4.Key.Pos = "itr"
		c4.YRange.Label = "Кол-во опросов"
		c4.AddDataPair("Опросы", dataP.x, dataP.y, chart.Style{LineColor: colors[0], LineWidth: 2, FillColor: colors[0]})
		c4.AddDataPair("Инлайн опросы", dataIP.x, dataIP.y, chart.Style{LineColor: colors[1], LineWidth: 2, FillColor: colors[1]})
		chartsSlice = append(chartsSlice, c4)
		resetData()

		//активность в год (столбцы)
		for k, v := range dataC5mapP {
			dataP.x = append(dataP.x, float64(k))
			dataP.y = append(dataP.y, float64(v))
		}
		for k, v := range dataC5mapIP {
			dataIP.x = append(dataIP.x, float64(k))
			dataIP.y = append(dataIP.y, float64(v))
		}
		c5 := &chart.BarChart{}
		c5.Title = "Недельная активность (не средняя, а суммарная), столбик - день"
		c5.Key.Pos = "itl"
		c5.YRange.Label = "Кол-во опросов"
		c5.AddDataPair("Опросы", dataP.x, dataP.y, chart.Style{LineColor: colors[0], LineWidth: 2, FillColor: colors[0]})
		c5.AddDataPair("Инлайн опросы", dataIP.x, dataIP.y, chart.Style{LineColor: colors[1], LineWidth: 2, FillColor: colors[1]})
		chartsSlice = append(chartsSlice, c5)
		resetData()

		//рисовка/ген
		for _, c := range chartsSlice {
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
			"Всего опросов:%18v\n"+
			"Всего голосов:%18v\n"+
			"Максимум голосов в опросе:%6v\n"+
			"Голосов на опрос в среднем:%5v\n"+
			"Создателей опросов:%13v\n\n"+
			"Всего инлайн опросов:%11v\n"+
			"Создателей инлайн опросов:%6v\n\n"+
			"Упоминаний Ракеты:%14v",
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

// Круговая диаграмма создателей опросов
func (a app) stage3() {
	type poll struct {
		idNick string
		n      int
	}

	var (
		err                   error
		strFileP, strFileIP   string
		strSplitP, strSplitIP []string
		polls                 []poll
		ipolls                []poll

		colors2 = []string{"#F44336", "#E91E63", "#9C27B0", "#673AB7", "#3F51B5", "#2196F3", "#00BCD4", "#009688", "#4CAF50", "#8BC34A", "#CDDC39", "#FFEB3B", "#FFC107", "#FF9800", "#FF5722", "#9E9E9E"}

		sumTop, sumAll int
		dataC6         = make([]opts.PieData, 0)
		dataC7         = make([]opts.PieData, 0)
	)

	//чтение файлов и преобразование
	page := components.NewPage()
	f2, err := os.Create(a.cfg.dir + "Polls/CreatorsChart.html")
	pnc(err)

	strFileP = a.readFile("Polls/PollsCreators.txt")
	strSplitP = strings.Split(strFileP, "\n")
	strSplitP = strSplitP[:len(strSplitP)-1]
	polls = make([]poll, 0, len(strSplitP))

	strFileIP = a.readFile("Polls/IPollsCreators.txt")
	strSplitIP = strings.Split(strFileIP, "\n")
	strSplitIP = strSplitIP[:len(strSplitIP)-1]
	ipolls = make([]poll, 0, len(strSplitIP))

	for _, str := range strSplitP {
		poll := poll{idNick: str[4:]}

		poll.n, err = strconv.Atoi(strings.Replace(str[:3], " ", "", -1))
		pnc(err, str)

		polls = append(polls, poll)
	}

	for _, str := range strSplitIP {
		poll := poll{idNick: str[4:]}

		poll.n, err = strconv.Atoi(strings.Replace(str[:3], " ", "", -1))
		pnc(err, str)

		ipolls = append(ipolls, poll)
	}

	//топ создателей опросов
	c6 := charts.NewPie()
	c6.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: "Топ 10 создателей опросов",
	}))
	sumTop, sumAll = 0, 0
	for i, v := range polls {
		if i < 10 {
			dataC6 = append(dataC6, opts.PieData{
				Name:  v.idNick,
				Value: v.n,
				ItemStyle: &opts.ItemStyle{
					Color: colors2[i],
				},
			})
			sumTop += v.n
		}
		sumAll += v.n
	}
	dataC6 = append(dataC6, opts.PieData{
		Name:  "Остальные",
		Value: sumAll - sumTop,
		ItemStyle: &opts.ItemStyle{
			Color: colors2[15],
		},
	})
	c6.AddSeries("Создатели", dataC6).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius: []string{"40%", "75%"},
			}),
		)

	//топ создателей инлайн опросов
	c7 := charts.NewPie()
	c7.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: "Топ 15 создателей инлайн опросов",
	}))
	sumTop, sumAll = 0, 0
	for i, v := range ipolls {
		if i < 15 {
			dataC7 = append(dataC7, opts.PieData{
				Name:  v.idNick,
				Value: v.n,
				ItemStyle: &opts.ItemStyle{
					Color: colors2[i],
				},
			})
			sumTop += v.n
		}
		sumAll += v.n
	}
	dataC7 = append(dataC7, opts.PieData{
		Name:  "Остальные",
		Value: sumAll - sumTop,
		ItemStyle: &opts.ItemStyle{
			Color: colors2[15],
		},
	})
	c7.AddSeries("Создатели", dataC7).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius: []string{"40%", "75%"},
			}),
		)

	//создание страницы
	page.AddCharts(c6, c7)
	page.Render(io.MultiWriter(f2))
}

//анализ слов
/*
//анализ слов
	textCut := strings.NewReplacer(" голос(а), ", "", " - ", "", "️⃣ ", "", "⃣", "", "ё", "е")
	textCut2 := strings.NewReplacer("\n", " ")
	textCut3 := strings.NewReplacer("  ", " ", "   ", " ", "    ", " ", "     ", " ")
	rgxSymbol := regexp.MustCompile(`(\p{P}|\d)`)
	ignoreWords := strings.Split("не и ли а бы или без из в вокруг в у с до над при для за к между на о об около от перед под по про с из-за я меня мне мной обо мне ты тебя тебе тобой он оно его ему его им нем она ее ей ее ею ней мы нас нам нами вы вас вам вами они их им ими них", " ")

	for _, pl := range polls {
		text += pl.question + "\n" + pl.results + "\n"
	}
	for _, ipl := range ipolls {
		text += ipl.question + "\n" + ipl.variants + "\n"
	}

	text = textCut.Replace(text)
	text = rgxSymbol.ReplaceAllLiteralString(text, "")
	text = textCut2.Replace(text)
	text = textCut3.Replace(text)
	text = strings.ToLower(text)
	for _, s := range strings.Split(text, " ") {
		if s != "" {
			textWords[s]++
		}
	}
	i := 0
	for _, v := range mapSort(textWords) {
		if inArray(v.Key, ignoreWords) {
			continue
		}
		if i == 50 {
			break
		}
		i++
		textOut += fmt.Sprintf("%2v %3v %v\n", i+1, v.Value, v.Key)
	}

	file = a.createFileNTrunc("Polls/Words.txt")
	file.WriteString(textOut)
	file.Close()
*/
