package main

import (
	"YetAnotherBotStatsOld/go-config"
	"os"
	"regexp"
	"sort"
	"strings"
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

func (a app) createFileNTrunc(path string) *os.File {
	file, err := os.OpenFile(a.cfg.dir+path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	pnc(err)
	stat, err := file.Stat()
	pnc(err)

	if stat.Size() > 0 {
		err = file.Truncate(0)
		pnc(err)
	}

	return file
}

func (a app) readFile(path string) (str string) {
	file, err := os.OpenFile(a.cfg.dir+path, os.O_RDONLY, 0666)
	pnc(err)
	defer file.Close()
	stat, err := file.Stat()
	pnc(err)

	strBytes := make([]byte, stat.Size())
	_, err = file.Read(strBytes)

	return string(strBytes)
}

type kv struct {
	Key   string
	Value int
}

func mapSort(mp map[string]int) (ss []kv) {
	for k, v := range mp {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return
}

// Объединяет ники по id
// На вход стринги с "#ADFL 🍕Pizza".
func mapNickTransformation(mp map[string]int) map[string]int {
	var id1, id2, id3, nick1, nick2 string
	bl := true

	for bl == true {
		bl = false
		mp2 := make(map[string]int, 0)

		for k1, v1 := range mp {
			id1 = k1[:strings.Index(k1, " ")]
			nick1 = k1[strings.Index(k1, " ")+1:]

			for k2, v2 := range mp {
				id2 = k2[:strings.Index(k2, " ")]
				nick2 = k2[strings.Index(k2, " ")+1:]

				if id1 == id2 && nick1 != nick2 {
					cont := false
					for k3, _ := range mp2 {
						id3 = k3[:strings.Index(k3, " ")]
						if id1 == id3 {
							cont = true
							break
						}
					}
					if cont {
						continue
					}

					mp2[id1+" "+nick1+"/"+nick2] = v1 + v2
					delete(mp, k1)
					delete(mp, k2)
					bl = true
				}
			}
		}

		if bl {
			for k, v := range mp2 {
				mp[k] = v
			}
		}
	}

	return mp
}

func inArray(str string, array []string) bool {
	for _, word := range array {
		if strings.Contains(str, word) {
			return true
		}
	}
	return false
}
