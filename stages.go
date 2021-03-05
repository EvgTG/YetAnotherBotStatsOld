package main

import (
	"fmt"
	"github.com/vdobler/chart"
	"image/color"
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
		dumper2 := a.NewDumper("Polls/CreatorsChart", 1, 2, 700, 500)
		defer dumper2.Close()

		var (
			charts  = make([]chart.Chart, 0, dumper.N*dumper.M)
			charts2 = make([]chart.Chart, 0, dumper2.N*dumper2.M)
			colors  = []color.RGBA{ParseHexColor("#FF0000"), ParseHexColor("#0000FF")}
			colors2 = []color.RGBA{
				ParseHexColor("#000000"),
				ParseHexColor("#E53935"),
				ParseHexColor("#D81B60"),
				ParseHexColor("#8E24AA"),
				ParseHexColor("#5E35B1"),
				ParseHexColor("#3949AB"),
				ParseHexColor("#1E88E5"),
				ParseHexColor("#039BE5"),
				ParseHexColor("#00ACC1"),
				ParseHexColor("#00897B"),
				ParseHexColor("#43A047"),
				ParseHexColor("#7CB342"),
				ParseHexColor("#C0CA33"),
				ParseHexColor("#FDD835"),
				ParseHexColor("#FFB300"),
				ParseHexColor("#FB8C00"),
				ParseHexColor("#F4511E"),
				ParseHexColor("#6D4C41"),
				ParseHexColor("#757575"),
				ParseHexColor("#546E7A"),
				ParseHexColor("#D0D3D4"),
			}

			dataC1 = xy{make([]float64, 0), make([]float64, 0)}

			dataC2P     = xy{make([]float64, 0), make([]float64, 0)}
			dataC2IP    = xy{make([]float64, 0), make([]float64, 0)}
			dataC2mapP  = make(map[string]int, 0)
			dataC2mapIP = make(map[string]int, 0)

			dataC3    = xy{make([]float64, 0), make([]float64, 0)}
			dataC3map = make(map[string]int, 0)

			dataC4P     = xy{make([]float64, 0), make([]float64, 0)}
			dataC4IP    = xy{make([]float64, 0), make([]float64, 0)}
			dataC4mapP  = make(map[int]int, 0)
			dataC4mapIP = make(map[int]int, 0)

			dataC5P     = xy{make([]float64, 0), make([]float64, 0)}
			dataC5IP    = xy{make([]float64, 0), make([]float64, 0)}
			dataC5mapP  = make(map[time.Weekday]int, 0)
			dataC5mapIP = make(map[time.Weekday]int, 0)

			sum         int
			dataC6      = make([]chart.CatValue, 0)
			dataC6Style = make([]chart.Style, 0)

			dataC7      = make([]chart.CatValue, 0)
			dataC7Style = make([]chart.Style, 0)
		)

		for _, pl := range polls {
			dataC1.x = append(dataC1.x, float64(pl.date.Unix()))
			dataC1.y = append(dataC1.y, float64(pl.usersN))

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
		c1.AddDataPair("polls", dataC1.x, dataC1.y, chart.PlotStylePoints,
			chart.Style{FillColor: colors[0], Symbol: 'o', LineWidth: 2})
		charts = append(charts, c1)

		//время к колву опросов в месяц (столбцы)
		for k, v := range dataC2mapP {
			tm, _ := time.Parse("2006-01", k)
			dataC2P.x = append(dataC2P.x, float64(tm.Unix()))
			dataC2P.y = append(dataC2P.y, float64(v))
		}
		for k, v := range dataC2mapIP {
			tm, _ := time.Parse("2006-01", k)
			dataC2IP.x = append(dataC2IP.x, float64(tm.Unix()))
			dataC2IP.y = append(dataC2IP.y, float64(v))
		}
		c2 := &chart.BarChart{}
		c2.Title = "Кол-во опросов в месяц"
		c2.Key.Pos = "itl"
		c2.XRange.Time = true
		c2.YRange.Label = "Кол-во опросов"
		c2.AddDataPair("Опросы", dataC2P.x, dataC2P.y, chart.Style{LineColor: colors[0], LineWidth: 2, FillColor: colors[0]})
		c2.AddDataPair("Инлайн опросы", dataC2IP.x, dataC2IP.y, chart.Style{LineColor: colors[1], LineWidth: 2, FillColor: colors[1]})
		charts = append(charts, c2)

		//время к колву голосов в месяц (столбцы)
		for k, v := range dataC3map {
			tm, _ := time.Parse("2006-01", k)
			dataC3.x = append(dataC3.x, float64(tm.Unix()))
			dataC3.y = append(dataC3.y, float64(v))
		}
		c3 := &chart.BarChart{}
		c3.Title = "Кол-во голосов в месяц"
		c3.Key.Hide = true
		c3.XRange.Time = true
		c3.YRange.Label = "Кол-во голосов"
		c3.AddDataPair("polls", dataC3.x, dataC3.y, chart.Style{LineColor: colors[0], LineWidth: 3, FillColor: colors[0]})
		charts = append(charts, c3)

		//активность в год (столбцы)
		for k, v := range dataC4mapP {
			dataC4P.x = append(dataC4P.x, float64(k))
			dataC4P.y = append(dataC4P.y, float64(v))
		}
		for k, v := range dataC4mapIP {
			dataC4IP.x = append(dataC4IP.x, float64(k))
			dataC4IP.y = append(dataC4IP.y, float64(v))
		}
		c4 := &chart.BarChart{}
		c4.Title = "Годовая активность (не средняя, а суммарная), столбик - неделя"
		c4.Key.Pos = "itr"
		c4.YRange.Label = "Кол-во опросов"
		c4.AddDataPair("Опросы", dataC4P.x, dataC4P.y, chart.Style{LineColor: colors[0], LineWidth: 2, FillColor: colors[0]})
		c4.AddDataPair("Инлайн опросы", dataC4IP.x, dataC4IP.y, chart.Style{LineColor: colors[1], LineWidth: 2, FillColor: colors[1]})
		charts = append(charts, c4)

		//активность в год (столбцы)
		for k, v := range dataC5mapP {
			dataC5P.x = append(dataC5P.x, float64(k))
			dataC5P.y = append(dataC5P.y, float64(v))
		}
		for k, v := range dataC5mapIP {
			dataC5IP.x = append(dataC5IP.x, float64(k))
			dataC5IP.y = append(dataC5IP.y, float64(v))
		}
		c5 := &chart.BarChart{}
		c5.Title = "Недельная активность (не средняя, а суммарная), столбик - день"
		c5.Key.Pos = "itl"
		c5.YRange.Label = "Кол-во опросов"
		c5.AddDataPair("Опросы", dataC5P.x, dataC5P.y, chart.Style{LineColor: colors[0], LineWidth: 2, FillColor: colors[0]})
		c5.AddDataPair("Инлайн опросы", dataC5IP.x, dataC5IP.y, chart.Style{LineColor: colors[1], LineWidth: 2, FillColor: colors[1]})
		charts = append(charts, c5)

		//топ создателей опросов
		c6 := &chart.PieChart{}
		sum = 0
		for i, v := range creatorsSliceP {
			if i == 10 {
				break
			}
			dataC6 = append(dataC6, chart.CatValue{
				Cat: fmt.Sprintf("%4.3v%% %v", float64(v.Value)/float64(len(polls))*100, v.Key[:strings.Index(v.Key, " ")]),
				Val: float64(v.Value),
			})
			dataC6Style = append(dataC6Style, chart.Style{LineColor: colors2[0], LineWidth: 0, FillColor: colors2[i+1]})
			sum += v.Value
		}
		dataC6 = append(dataC6, chart.CatValue{
			Cat: fmt.Sprintf("%4.3v%% %v", float64(float64(len(polls)-sum))/float64(len(polls))*100, "Остальные"),
			Val: float64(len(polls) - sum),
		})
		dataC6Style = append(dataC6Style, chart.Style{LineColor: colors2[0], LineWidth: 1, FillColor: colors2[len(colors2)-1]})
		c6.Key.Pos = "ort"
		c6.Key.Border = -1
		c6.Title = "Топ 10 создателей опросов"
		c6.AddData("Создатели", dataC6, dataC6Style)
		charts2 = append(charts2, c6)

		//топ создателей инлайн опросов
		c7 := &chart.PieChart{}
		sum = 0
		for i, v := range creatorsSliceIP {
			if i == 15 {
				break
			}
			dataC7 = append(dataC7, chart.CatValue{
				Cat: fmt.Sprintf("%4.3v%% %v", float64(v.Value)/float64(len(ipolls))*100, v.Key[:strings.Index(v.Key, " ")]),
				Val: float64(v.Value),
			})
			dataC7Style = append(dataC7Style, chart.Style{LineColor: colors2[0], LineWidth: 0, FillColor: colors2[i+1]})
			sum += v.Value
		}
		dataC7 = append(dataC7, chart.CatValue{
			Cat: fmt.Sprintf("%4.3v%% %v", float64(float64(len(ipolls)-sum))/float64(len(ipolls))*100, "Остальные"),
			Val: float64(len(ipolls) - sum),
		})
		dataC7Style = append(dataC7Style, chart.Style{LineColor: colors2[0], LineWidth: 1, FillColor: colors2[len(colors2)-1]})
		c7.Key.Pos = "ort"
		c7.Key.Border = -1
		c7.Title = "Топ 15 создателей инлайн опросов"
		c7.AddData("Создатели", dataC7, dataC7Style)
		charts2 = append(charts2, c7)

		//рисовка
		for _, c := range charts {
			dumper.Plot(c)
			c.Reset()
		}
		for _, c := range charts2 {
			dumper2.Plot(c)
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
