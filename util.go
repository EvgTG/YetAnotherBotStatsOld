package main

import (
	"YetAnotherBotStatsOld/go-config"
	"fmt"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type rgx struct {
	rgxPollClose *regexp.Regexp
	rgxIPoll     *regexp.Regexp
}

func NewRegexp() *rgx {
	return &rgx{
		rgxPollClose: regexp.MustCompile(`^\[Bot\] Результаты голосования за вопрос`),
		rgxIPoll:     regexp.MustCompile(`^\[Bot\] #.+ поставил\(а\) вопрос:\n`),
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

func pnc(err error, out ...string) {
	if err != nil {
		if len(out) > 0 {
			fmt.Printf("\n\n %v", strings.Join(out, "\n"))
		}
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

func mapToSlice(mp map[string]int) (ss []kv) {
	for k, v := range mp {
		ss = append(ss, kv{k, v})
	}
	return
}

func mapSortByTime(mp map[string]int, timeLayout string) (ss []kv) {
	for k, v := range mp {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		tmi, _ := time.Parse(timeLayout, ss[i].Key)
		tmj, _ := time.Parse(timeLayout, ss[j].Key)
		return tmi.Unix() < tmj.Unix()
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
		if str == word {
			return true
		}
	}
	return false
}

func inArrayContains(str string, array []string) bool {
	for _, word := range array {
		if strings.Contains(str, word) {
			return true
		}
	}
	return false
}

func kvToStr(kv []kv) (strRet []string) {
	for _, k := range kv {
		strRet = append(strRet, k.Key)
	}
	return
}

type xy struct {
	x, y []float64
}

type Dumper struct {
	N, M, W, H, Cnt int
	I               *image.RGBA
	imgFile         *os.File
}

func (a app) NewDumper(name string, n, m, w, h int) *Dumper {
	var err error
	dumper := Dumper{N: n, M: m, W: w, H: h}

	dumper.imgFile, err = os.Create(a.cfg.dir + name + ".png")
	if err != nil {
		panic(err)
	}
	dumper.I = image.NewRGBA(image.Rect(0, 0, n*w, m*h))
	bg := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	draw.Draw(dumper.I, dumper.I.Bounds(), bg, image.ZP, draw.Src)

	return &dumper
}

func (d *Dumper) Close() {
	png.Encode(d.imgFile, d.I)
	d.imgFile.Close()
}

func (d *Dumper) Plot(c chart.Chart) {
	row, col := d.Cnt/d.N, d.Cnt%d.N

	igr := imgg.AddTo(d.I, col*d.W, row*d.H, d.W, d.H, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	c.Plot(igr)

	d.Cnt++
}

func ParseHexColor(s string) (c color.RGBA) {
	var err error
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	pnc(err)
	return
}
