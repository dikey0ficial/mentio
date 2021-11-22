package main

import (
	"encoding/json"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	regs = []int{}
)

func reply(mID int, txt string, peer int64, bot *tg.BotAPI) {
	msg := tg.NewMessage(peer, txt)
	msg.ParseMode = "markdown"
	msg.DisableWebPagePreview = true
	msg.ReplyToMessageID = mID
	bot.Send(msg)
}

func isInList(s string, list []string) bool {
	ss := strings.ToLower(s)
	for _, val := range list {
		if val == ss {
			return true
		}
	}
	return false
}

func startsList(s string, list []string) bool {
	ss := strings.ToLower(s)
	for _, val := range list {
		if strings.HasPrefix(ss, val) {
			return true
		}
	}
	return false
}

func isInRegs(id int) bool {
	for _, val := range regs {
		if id == val {
			return true
		}
	}
	return false
}

func unreg(id int) {
	for i, val := range regs {
		if val == id {
			regs = append(regs[:i], regs[i+1:]...)
		}
	}
}

func isInChat(uid int, cid int64, bot *tg.BotAPI) bool {
	var is bool = false
	const errnf = "Bad Request: user not found"
	m, err := bot.GetChatMember(tg.ChatConfigWithUser{
		ChatID: cid,
		UserID: uid,
	})
	if err != nil && err.Error() != errnf {
		log.Panicln(err)
	} else if m.Status != "left" && m.Status != "kicked" {
		is = true
	}
	return is
}

type config struct {
	Token string `json:"bot_token"`
}

var conf config = func() config {
	var c config
	file, err := os.Open("config.json")
	defer file.Close()
	if err != nil {
		log.Panicln(err)
		return c
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		log.Panicln(err)
		return c
	}
	return c
}()

func load() {
	file, err := os.Open("regs.json")
	defer file.Close()
	if err != nil {
		log.Panicln(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&regs)
	if err != nil {
		log.Panicln(err)
	}
}

func write() {
	mar, err := json.Marshal(regs)
	if err != nil {
		log.Panicln(err)
	}
	err = ioutil.WriteFile("regs.json", mar, 0644)
	if err != nil {
		log.Panicln(err)
	}
}

func main() {
	bot, err := tg.NewBotAPI(conf.Token)
	if err != nil {
		log.Panicln(err)
	}

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	rand.Seed(time.Now().Unix())

	log.Println("Start!")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		load()
		mid, txt := update.Message.MessageID, strings.TrimSpace(strings.ToLower(update.Message.Text))
		peer, from := update.Message.Chat.ID, update.Message.From.ID
		isGroup := update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()
		go func(update tg.Update) {
			if update.Message.NewChatMembers != nil && len(*update.Message.NewChatMembers) != 0 {
				members := *update.Message.NewChatMembers
				if members[0].ID == bot.Self.ID {
					reply(mid, fmt.Sprintf(
						"Спасибо за то, что добавили бота в группу! Всем участникам надо [зарегистрироваться в боте](t.me/%s?start=reg), чтобы он работал",
						bot.Self.UserName),
						peer, bot,
					)
					return
				}
				var msg string
				msg += "Здравствуйте, "
				for i, user := range members {
					if user.IsBot {
						if len(members) == 1 {
							return
						}
						continue
					}
					msg += fmt.Sprintf("[%s](tg:/user?id=%d)", user.FirstName, user.ID)
					if isInRegs(user.ID) {
						msg += " {зарегистрирован(а)}"
					} else {
						regs = append(regs, user.ID)
						write()
					}
					if i < len(members)-1 {
						msg += ", "
					}
				}
				msg += "!\n"
				msg += "Вы были автоматически зарегистрированы (если ещё не были) в боте [(что это значит?)](https://telegra.ph/CHto-znachit-registraciya-v-Mentio-11-22)."
				msg += fmt.Sprintf("\n[Разрегистрироваться](t.me/%s?start=unreg)", bot.Self.UserName)
				reply(mid, msg, peer, bot)
				return
			}

			if startsList(txt, []string{"/reg", ("/reg@" + bot.Self.UserName), "/start reg",
				("/start@" + bot.Self.UserName + " reg"),
			}) {
				if isInRegs(from) {
					reply(mid, "Вы уже зарегистрировались", peer, bot)
					return
				}
				regs = append(regs, from)
				reply(mid, "Теперь вы зарегистрированны!", peer, bot)
				write()
			} else if startsList(txt, []string{"/unreg", ("/unreg@" + bot.Self.UserName), "/start unreg",
				("/start@" + bot.Self.UserName + " unreg"),
			}) {
				if isInRegs(from) {
					unreg(from)
					reply(mid, "Вы больше не получите уведомлений о призыве", peer, bot)
					write()
				} else {
					reply(mid, "Вы не зарегистрированны...", peer, bot)
				}
			} else if startsList(txt, []string{"/start gr", ("/start@" + bot.Self.UserName + " gr")}) {
				reply(mid, fmt.Sprintf(
					"Спасибо за то, что добавили бота в группу! Всем участникам надо [зарегистрироваться в боте](t.me/%s?start=reg), чтобы он работал",
					bot.Self.UserName),
					peer, bot,
				)
			} else if startsList(txt, []string{"/help", ("/help@" + bot.Self.UserName)}) {
				reply(mid,
					"Достаточно всего лишь написать @all в сообщении, и бот призовёт всех\n\n"+
						"Команды:\n — reg — регистрирует\n — unreg — отрегистрирует\n"+
						" — add — добавляет бота в группу\n — help — отправляет это",
					peer, bot)
			} else if isInList(txt, []string{"/start", ("/start@" + bot.Self.UserName)}) {
				if strings.Contains(txt, " ") {
					reply(mid, fmt.Sprintf("бот не понимает, что означает \"%s\" в ссылке", strings.Fields(txt)[1]), peer, bot)
				}
			} else if startsList(txt, []string{"/add", ("/add@" + bot.Self.UserName)}) {
				reply(mid, fmt.Sprintf(
					"[Нажмите, чтобы добавить бота в группу](t.me/%s?startgroup=gr)", bot.Self.UserName),
					peer, bot)
			} else {
				for _, word := range strings.Fields(txt) {
					var done bool = true
					switch word {
					case "@all":
						if !isGroup {
							reply(mid, fmt.Sprintf(
								"Насчитано %d личностей", rand.Intn(5)+2),
								peer, bot)
							return
						}
						var inChat []int
						for i := 0; i < len(regs); i++ {
							if isInChat(regs[i], peer, bot) {
								inChat = append(inChat, regs[i])
							}
						}
						var msg string = "all:\n"
						for _, id := range inChat {
							i := rand.Intn(42)
							msg += fmt.Sprintf("[%d](tg://user?id=%d) ", i, id)
						}
						if msg == "all:\n" {
							reply(mid, fmt.Sprintf(
								"all: Никто в этом чате не зарегистрирован...\n\n/reg@%s для регистрации",
								bot.Self.UserName), peer, bot)
							return
						}
						reply(mid, msg, peer, bot)
						break
					case "@here":
						if !isGroup {
							i := rand.Intn(5) + 2
							reply(mid, fmt.Sprintf(
								"Насчитано %d личностей, из них %d в сети", i, i-rand.Intn(i-1)),
								peer, bot)
							return
						}
						reply(mid,
							"К сожалению, данная функция недоступна :(",
							peer, bot)
					default:
						done = false
					}
					if done {
						break

					}
				}
			}
		}(update)
	}
}
