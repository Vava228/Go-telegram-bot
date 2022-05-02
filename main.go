package main

import (
  "fmt"
	"log"
  "strings"
  "strconv"
  "net/http"
  "encoding/json"
  "errors"
  "math/rand"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const token string = "5345348417:AAFFeKLIp_F0kwv1NEWC0IueYWYCdGbJ3tI"

type bnResp struct {
  Price float64 `json:"price,string"`
  Code int64 `json:"code"`
}

type wallet map[string]float64
var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
    errormsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не распознанная команда!")
    command := strings.Split(update.Message.Text, " ")

    switch command[0] {
    case "ROLL":
      max, err := strconv.Atoi(command[2])
      min, err := strconv.Atoi(command[1])
      if max < min {
        max, min = min, max
      }
      if len(command) != 3 || err != nil {
        bot.Send(errormsg)
      }
      result := strconv.Itoa(rand.Intn(max - min) + min)
      bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, ("В результате броска от " + command[1] + " до " + command[2] + " выпало число: " + result)))
    case "/start":
      if len(command) != 1 {
        bot.Send(errormsg)
      }
      bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Вот список доступных команд:\nADD(добавить валюту)\nSUB(продать валюту)\nDEL\nSHOW(показать список валют)\nROLL(получить случайное число)"))
    case "ADD":
      amount, err := strconv.ParseFloat(command[2], 64)
      if len(command) != 3 || err != nil {
        bot.Send(errormsg)
      }

      if _, ok := db[update.Message.Chat.ID]; !ok {
        db[update.Message.Chat.ID] = wallet{}
      }

      db[update.Message.Chat.ID][command[1]] += amount
      successmsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Счёт " + command[1] + " успешно пополнен на " + command[2])
      bot.Send(successmsg)
    case "SUB":
      amount, err := strconv.ParseFloat(command[2], 64)
      if len(command) != 3 || err != nil {
        bot.Send(errormsg)
      }

      if _, ok := db[update.Message.Chat.ID]; !ok {
        db[update.Message.Chat.ID] = wallet{}
      }

      db[update.Message.Chat.ID][command[1]] -= amount
      successmsg := tgbotapi.NewMessage(update.Message.Chat.ID, "Со счета " + command[1] + " было списано " + command[2])
    bot.Send(successmsg)
    case "DEL":
        if len(command) != 2 {
          bot.Send(errormsg)
        }

        delete(db[update.Message.Chat.ID], command[1])
        bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Было успешно удалено " + command[1]))
    case "SHOW":
      if len(command) != 1 {
        bot.Send(errormsg)
      }
      msg := "Текущий счёт:\n"
      for key, value := range db[update.Message.Chat.ID] {
        price, _ := getPrice(key)

        msg += fmt.Sprintf("%s: %f [%f $]\n", key, value, value*price)
      }
      bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
    default:
      bot.Send(errormsg)
    }
	}
}

func getPrice(symbol string) (price float64, err error) {
  resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
  if err != nil {
    return
  }
  defer resp.Body.Close()

  var jsonResp bnResp
  err = json.NewDecoder(resp.Body).Decode(&jsonResp)
  if err != nil {
    return
  }
  if jsonResp.Code != 0 {
    err = errors.New("Wrong Symbol")
  }

  price = jsonResp.Price

  return
}
