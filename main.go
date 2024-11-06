package main

import (
	"log"
	"strconv"
	"time"
	"GriBotHabitLev/internal"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура для хранения информации о привычке
type Habit struct {
	Name          string
	Count         int       // Счётчик выполнений привычки
	LastCompleted time.Time // Последний день выполнения
	Streak        int       // Количество дней подряд выполнения
	Reminder      bool      // Флаг, нужно ли напоминание
}

// Карта для хранения привычек по идентификатору пользователя
var habits = make(map[int64][]Habit)

func main() {
	botToken := internal.BotToken // Получаем токен из config.go
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		internal.LogError(err)
		log.Panic(err)
	}

	bot.Debug = true
	internal.LogInfo("Авторизация выполнена на аккаунте: " + bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Основной цикл обработки обновлений от Telegram
	for update := range updates {
		if update.Message != nil {
			// Обрабатываем команды пользователя
			switch update.Message.Command() {
			case "add":
				handleAddHabit(bot, update.Message)
			case "complete":
				handleCompleteHabit(bot, update.Message)
			case "stats":
				handleStats(bot, update.Message)
			case "reminder":
				handleReminder(bot, update.Message)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте команды: /add, /complete, /stats, /reminder")
				bot.Send(msg)
			}
		}
	}
}

// handleAddHabit добавляет привычку для пользователя
func handleAddHabit(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	if habitName == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста, укажите название привычки.")
		bot.Send(msg)
		return
	}
	habits[message.Chat.ID] = append(habits[message.Chat.ID], Habit{Name: habitName})
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка \""+habitName+"\" добавлена.")
	bot.Send(msg)
}

// handleCompleteHabit отмечает привычку как выполненную
func handleCompleteHabit(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	for i, habit := range habits[message.Chat.ID] {
		if habit.Name == habitName {
			habits[message.Chat.ID][i].Count++
			habits[message.Chat.ID][i].LastCompleted = time.Now()
			habits[message.Chat.ID][i].Streak++
			msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка \""+habitName+"\" выполнена!")
			bot.Send(msg)
			return
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка \""+habitName+"\" не найдена.")
	bot.Send(msg)
}

// handleStats выводит статистику по привычкам пользователя
func handleStats(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	stats := "Ваши привычки:\n"
	for _, habit := range habits[message.Chat.ID] {
		stats += "- " + habit.Name + ": выполнено " + strconv.Itoa(habit.Count) + " раз, подряд " + strconv.Itoa(habit.Streak) + " дней\n"
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, stats)
	bot.Send(msg)
}

// handleReminder устанавливает напоминание для привычки
func handleReminder(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	for i, habit := range habits[message.Chat.ID] {
		if habit.Name == habitName {
			habits[message.Chat.ID][i].Reminder = !habits[message.Chat.ID][i].Reminder
			reminderStatus := "включено"
			if !habits[message.Chat.ID][i].Reminder {
				reminderStatus = "выключено"
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "Напоминание для привычки \""+habitName+"\" "+reminderStatus+".")
			bot.Send(msg)
			return
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка \""+habitName+"\" не найдена.")
	bot.Send(msg)
}

