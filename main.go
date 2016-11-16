package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"gopkg.in/telegram-bot-api.v4"
	"gopkg.in/yaml.v2"
)

const WEATHER = "Погода в %s на %s\nТемпература: %.1f°C (%.1f°C..%.1f°C)\nДавление: %d мм/рт.с\nВлажность: %d%%\nВетер: %s %.1f м/с"

type c struct {
	Bot struct {
		Username string `yaml:"username"`
		Token    string `yaml:"token"`
	}
	Weather struct {
		City  string `yaml:"city"`
		Token string `yaml:"token"`
	}
}

type w struct {
	Coord struct {
		Lon float64
		Lat float64
	}
	Weather []struct {
		Id          int64
		Main        string
		Description string
		Icon        string
	}
	Base string
	Main struct {
		Temp     float64
		Pressure int64
		Humidity int64
		Temp_min float64
		Temp_max float64
	}
	Wind struct {
		Speed float64
		Deg   int64
	}
	Clouds struct {
		All int64
	}
	Rain struct {
		_3h int64
	}
	Dt  int64
	Sys struct {
		Type    int64
		Id      int64
		Message float64
		Country string
		Sunrise int64
		Sunset  int64
	}
	Id   int64
	Name string
	Cod  int64
}

func readconfig(data string) c {
	out := c{}
	err := yaml.Unmarshal([]byte(data), &out)
	if err != nil {
		log.Fatalf("Wrong config data.")
	}
	return out
}

func main() {
	cdata, err := ioutil.ReadFile(".config.yml")
	conf := readconfig(string(cdata))
	bot, err := tgbotapi.NewBotAPI(conf.Bot.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	ucfg := tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	updates, err := bot.GetUpdatesChan(ucfg)
	for {
		select {
		case update := <-updates:
			UserName := update.Message.From.UserName
			UserID := update.Message.From.ID
			ChatID := update.Message.Chat.ID
			Text := update.Message.Text

			log.Printf("%s %d %d <- %s", UserName, UserID, ChatID, Text)

			var reply string
			var quote bool

			/*			if update.Message.NewChatMember.UserName != "" {
						reply = fmt.Sprintf(`Привет @%s! Ты чего сюда припёрся?`, update.Message.NewChatMember.UserName)
						quote = false
					}*/

			switch Text {
			case "/weather":
				{
					var weather w
					resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?units=metric&id=%s&APPID=%s", conf.Weather.City, conf.Weather.Token))
					if err != nil {
						log.Println(err)
						reply = "No data."
					}
					body, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
						log.Println(err)
						reply = "No data."
					}
					err = json.Unmarshal([]byte(body), &weather)
					if err != nil {
						log.Println(err)
						reply = "No data."
					}
					//					fmt.Println(string(body))
					//					fmt.Println(weather)
					reply = fmt.Sprintf(
						WEATHER,
						weather.Name,
						time.Unix(weather.Dt, 0),
						weather.Main.Temp,
						weather.Main.Temp_min,
						weather.Main.Temp_max,
						int64(float64(weather.Main.Pressure)/1.333224),
						weather.Main.Humidity,
						winddirection(weather.Wind.Deg),
						weather.Wind.Speed,
					)
				}
			case "/stocks":
				{
					resp, err := http.Get("http://finance.yahoo.com/d/quotes.csv?s=MAIL.IL&f=l1p2")
					if err != nil {
						reply = "No data."
					}
					body, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
						reply = "No data."
					}
					reply = fmt.Sprintf("%s", body)
				}
			case "/photo":
				{

				}
			default:
				{

				}
			}

			if reply != "" {
				msg := tgbotapi.NewMessage(ChatID, reply)
				if quote {
					msg.ReplyToMessageID = update.Message.MessageID
				}
				log.Printf("%s %d %d -> %s", UserName, UserID, ChatID, reply)
				bot.Send(msg)
			}
		}
	}

	/*	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}*/
}

func winddirection(i int64) string {
	winds := map[int]string{
		0:   "С",
		22:  "ССВ",
		45:  "СВ",
		67:  "ВСВ",
		90:  "В",
		112: "ВЮВ",
		135: "ЮВ",
		157: "ЮЮВ",
		180: "Ю",
		202: "ЮЮЗ",
		225: "ЮЗ",
		247: "ЗЮЗ",
		270: "З",
		292: "ЗСЗ",
		315: "СЗ",
		337: "ССЗ",
		360: "С",
	}
	var keys []int
	for k := range winds {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		if k >= int(i) {
			return winds[k]
		}
	}
	return ""
}
