package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type message struct {
	ID        int64
	Type      string // message animation sticker photo voice_message audio_file video_file video_message
	Date      time.Time
	FromID    int64
	ReplyToID int64
	Text      string
}

//TODO дополнить message параметрами в будущем
type messageForUnmarshal struct {
	ID           float64     `json:"id"`
	Type         string      `json:"type"`
	Date         string      `json:"date"`
	Edited       string      `json:"edited"`
	FromID       float64     `json:"from_id"`
	ReplyToID    float64     `json:"reply_to_message_id"`
	Text         interface{} `json:"text"`
	MediaType    string      `json:"media_type"`
	StickerEmoji string      `json:"sticker_emoji"`
	Photo        string      `json:"photo"`
	Performer    string      `json:"performer"`
	Title        string      `json:"title"`
}

func (a app) unmarshalChan() chan message {
	ch := make(chan message, 100)

	go func(ch chan message) {
		var lines []string

		i, n, limiter := 0, 600000, false
		scanner := bufio.NewScanner(a.file)
		for scanner.Scan() {
			i++
			if i > n && limiter {
				break
			}

			if scanner.Text() == `  {` {
				lines = []string{}
			} else if scanner.Text() == `  },` {
				lines = append(lines, `  }`)
				ch <- unmarshalJson(strings.Join(lines, ""))
				continue
			}
			lines = append(lines, scanner.Text())
		}

		close(ch)
	}(ch)

	return ch
}

func unmarshalJson(str string) message {
	var (
		err error
		ok  bool
	)

	msgRaw := messageForUnmarshal{}
	err = json.Unmarshal([]byte(str), &msgRaw)
	pnc(err)

	msg := message{
		ID:        int64(msgRaw.ID),
		Type:      msgRaw.Type,
		Date:      time.Time{},
		FromID:    int64(msgRaw.FromID),
		ReplyToID: int64(msgRaw.ReplyToID),
		Text:      "",
	}

	msg.Date, err = time.Parse("2006-01-02T15:04:05", msgRaw.Date)

	if msgRaw.MediaType != "" {
		msg.Type = msgRaw.MediaType
	}
	if msgRaw.Photo != "" {
		msg.Type = "photo"
	}

	if _, ok = msgRaw.Text.(string); ok {
		msg.Text = msgRaw.Text.(string)
	} else { //TODO перебрать какие типы могут быть ("type": "hashtag" интересен)
		for i := 0; i < len(msgRaw.Text.([]interface{})); i++ {
			if _, ok = msgRaw.Text.([]interface{})[i].(string); ok {
				msg.Text += msgRaw.Text.([]interface{})[i].(string)
			} else if _, ok = msgRaw.Text.([]interface{})[i].(map[string]interface{}); ok {
				if _, ok = msgRaw.Text.([]interface{})[i].(map[string]interface{})["text"].(string); ok {
					msg.Text += msgRaw.Text.([]interface{})[i].(map[string]interface{})["text"].(string)
				} else {
					panic(errors.New("msgRaw.Text type error"))
				}
			}
		}
	}

	return msg
}

/*

{   "id": 107435,   "type": "message",   "date": "2016-04-24T20:13:31",   "from": "Zhenya",   "from_id": секрет,   "text": [    {     "type": "bot_command",     "text": "/start"    }   ]  }

{   "id": 107437,   "type": "message",   "date": "2016-04-24T20:13:45",   "from": "Анонимный чат",   "from_id": 4389402929,   "reply_to_message_id": 107390,   "text": "🐧:  люблю такие подъеbочки =)"  }

{   "id": 107458,   "type": "message",   "date": "2016-04-25T09:06:22",   "from": "Анонимный чат",   "from_id": 4389402929,   "file": "(File not included. Change data exporting settings to download.)",   "thumbnail": "(File not included. Change data exporting settings to download.)",   "media_type": "animation",   "mime_type": "video/mp4",   "duration_seconds": 2,   "width": 480,   "height": 182,   "text": ""  }

{   "id": 107471,   "type": "message",   "date": "2016-04-25T09:26:09",   "from": "Анонимный чат",   "from_id": 4389402929,   "file": "(File not included. Change data exporting settings to download.)",   "thumbnail": "(File not included. Change data exporting settings to download.)",   "media_type": "sticker",   "sticker_emoji": "🚬",   "width": 512,   "height": 512,   "text": ""  }

{   "id": 107570,   "type": "message",   "date": "2016-04-25T10:08:41",   "from": "Анонимный чат",   "from_id": 4389402929,   "photo": "(File not included. Change data exporting settings to download.)",   "width": 948,   "height": 1280,   "text": "Дмитрий 🚨"  }

{   "id": 108297,   "type": "message",   "date": "2016-04-25T12:51:33",   "from": "Анонимный чат",   "from_id": 4389402929,   "file": "(File not included. Change data exporting settings to download.)",   "media_type": "audio_file",   "performer": "Gwen Stefani",   "title": "Cool.",   "mime_type": "audio/mpeg",   "duration_seconds": 246,   "text": ""  }

{   "id": 117330,   "type": "message",   "date": "2016-05-01T04:59:56",   "edited": "2016-05-01T05:05:42",   "from": "Анонимный чат",   "from_id": 4389402929,   "text": [    {     "type": "bold",     "text": "[Bot]"    },    " Объявлено голосование! ",    {     "type": "hashtag",     "text": "#ME"    },    " `Кот🐱` поставил(а) вопрос:\n",    {     "type": "italic",     "text": "Я завтра проблююсь?"    },    "\nВыбери вариант ответа:\n",    {     "type": "bold",     "text": "1"    },    " - ",    {     "type": "italic",     "text": "Да"    },    "\n",    {     "type": "bold",     "text": "2"    },    " - ",    {     "type": "italic",     "text": "Нет"    },    ""   ]  }

*/

//[id type date from from_id text reply_to_message_id file media_type mime_type width height thumbnail duration_seconds sticker_emoji photo performer title edited] ненужные - [via_bot forwarded_from contact_information location_information]

/*
	indexs := make([]string, 0)


	var result map[string]interface{}
	err := json.Unmarshal([]byte(strings.Join(lines, "")), &result)
	pnc(err)

	blPrint := false
	for index, _ := range result {
		blAdd := true

		for ii := range indexs {
			if indexs[ii] == index {
				blAdd = false
			}
		}

		if blAdd {
			blPrint = true
			indexs = append(indexs, index)
		}
	}
	if blPrint {
		fmt.Println(strings.Join(lines, ""), "\n")
	}
	//fmt.Println(result)


	fmt.Println("\n\n", indexs)
*/
