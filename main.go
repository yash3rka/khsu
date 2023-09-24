package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/go-toast/toast"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	botToken    string = ""                //<--------- WRITE TOKEN HERE!!!
	chatID      int64  = int64(1234567890) //<------------- Write ChatID here!!!
	url         string
	FileName    string
	PreviosName string
	Hash        string
	PreviosHash string
	tick        int
	fileUpload  tgbotapi.DocumentConfig
	rFile       string
)

func main() {
	var CF_err error = nil
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		fmt.Println("Telegram API not avalible!")
	}
	go handleUpdates(bot)

	for {
		CF_err = ChangeFiles(tick, bot, err)
		if CF_err != nil {
			fmt.Println("[Error: " + string(CF_err.Error()) + "] Произошла ошибка! :(")
			CF_err = nil
		}
		tick++
		time.Sleep(30 * time.Second)
	}
}

func handleUpdates(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Println("Проблема с получением обновлений!")
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "alive":
				tickStr := fmt.Sprintf("%d", tick)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tick: "+tickStr)
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println("Ошибка отправки сообщения:", err)
				}
			}
		}
	}
}

func ChangeFiles(tick int, bot *tgbotapi.BotAPI, err error) error {
	var download_err error = nil
	var url_err error = nil
	if tick == 0 { //startup
		url_err = GetEndUrl("https://iti.khsu.ru/student/schedule")
		if url_err != nil {
			return errors.New("1")
		}
		if url == "" {
			fmt.Println("[" + time.Now().Format(time.DateTime) + "] Проблема при получении ссылки на файл!")
			return errors.New("3")
		}
		url = regexp.MustCompile(`/edit#gid=.+`).ReplaceAllString(url, "/export?exportFormat=csv")
		FileName = "output\\" + strings.Replace(strings.Replace(time.Now().Format(time.DateTime)+".csv", ":", "_", -1), " ", "_", -1)
		download_err = downloadFile(url, FileName) //Загрузка файла
		if download_err != nil {
			fmt.Print("[" + time.Now().Format(time.DateTime) + "] Загрузка файла невозможна!")
			return errors.New("1")
		}
		Hash = s1(FileName) //Хеш первого файла
		os.Remove(FileName)
		FileName = strings.Replace(FileName, ".csv", ".xlsx", -1)
		url = strings.Replace(url, "=csv", "=xlsx", -1)
		download_err = downloadFile(url, FileName) //Загрузка файла
		if download_err != nil {
			fmt.Print("[" + time.Now().Format(time.DateTime) + "] Загрузка файла невозможна!")
			return errors.New("1")
		}
		fileUpload = tgbotapi.NewDocumentUpload(chatID, FileName)
		fileUpload.Caption = "Reboot! Hash: " + Hash
		_, err = bot.Send(fileUpload)
		if err != nil {
			fmt.Println("[" + time.Now().Format(time.DateTime) + "] Отправка файла затруднена!")
			return errors.New("2")
		}
		NoteWin()
		//Отстук в тг ...
	} else { //loop
		PreviosHash = Hash
		PreviosName = FileName
		FileName = ""
		Hash = ""
		url_err = GetEndUrl("https://iti.khsu.ru/student/schedule")
		if url_err != nil {
			Hash = PreviosHash
			FileName = PreviosName
			return errors.New("1")
		}
		if url == "" {
			Hash = PreviosHash
			FileName = PreviosName
			fmt.Print("[" + time.Now().Format(time.DateOnly) + "] Проблема при получении ссылки на файл!")
			return errors.New("-1")
		}
		url = regexp.MustCompile(`/edit#gid=.+`).ReplaceAllString(url, "/export?exportFormat=csv")
		FileName = "output\\" + strings.Replace(strings.Replace(time.Now().Format(time.DateTime)+".csv", ":", "_", -1), " ", "_", -1)
		download_err = downloadFile(url, FileName) //Загрузка файла
		if download_err != nil {
			Hash = PreviosHash
			FileName = PreviosName
			fmt.Print("[" + time.Now().Format(time.DateTime) + "] Загрузка файла невозможна!")
			return errors.New("1")
		}
		Hash = s1(FileName) //Хеш файла
		os.Remove(FileName)
		if Hash == PreviosHash {
			fmt.Println("[" + time.Now().Format(time.DateTime) + "] Файл идентичен преведущему! Hash: " + Hash)
		} else {
			fmt.Println("[" + time.Now().Format(time.DateTime) + "] Новое расписание найдено! Hash: " + Hash)
			rFile = strings.ReplaceAll(PreviosName, ".csv", ".xlsx")
			os.Remove(rFile)
			FileName = strings.ReplaceAll(FileName, ".csv", ".xlsx")
			url = strings.Replace(url, "=csv", "=xlsx", -1)
			download_err = downloadFile(url, FileName) //Загрузка файла
			if download_err != nil {
				Hash = PreviosHash
				FileName = PreviosName
				fmt.Print("[" + time.Now().Format(time.DateTime) + "] Загрузка файла невозможна!")
				return errors.New("1")
			}
			fileUpload = tgbotapi.NewDocumentUpload(chatID, FileName)
			fileUpload.Caption = "Hash: " + Hash
			_, err = bot.Send(fileUpload)
			if err != nil {
				Hash = PreviosHash
				FileName = PreviosName
				fmt.Println("[" + time.Now().Format(time.DateTime) + "] Отправка файла затруднена!")
				return errors.New("1")
			}
			NoteWin()
		}
	}
	return nil
}

func s1(FileName string) string {
	file, _ := os.Open(FileName)
	sha1Hash := sha1.New()
	io.Copy(sha1Hash, file)
	hashBytes := sha1Hash.Sum(nil)
	Hash = fmt.Sprintf("%x", hashBytes)
	file.Close()
	return Hash
}
func downloadFile(URL, fileName string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func NoteWin() {
	note := toast.Notification{
		AppID:   "khsu iti",
		Title:   "Найдено новое расписание!",
		Message: "Hash: " + Hash,
	}
	err := note.Push()
	if err != nil {
		panic(err)
	}
}

func GetEndUrl(FirstUrl string) error {
	response, err := soup.Get(FirstUrl)
	if err != nil {
		fmt.Println("[" + time.Now().Format(time.DateTime) + "] Проблемы с доступом к сайту!")
		return errors.New("-2")
	}
	doc := soup.HTMLParse(response)
	list := doc.FindAll("a")
	for _, element := range list {
		if element.Text() == "РАСПИСАНИЕ  (google)" {
			url = element.Attrs()["href"]
			break
		}
	}
	return nil
}
