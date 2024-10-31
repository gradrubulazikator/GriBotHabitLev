package main

import (
	"fmt"
	"log"
	"time"
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

// Создайте экземпляр бота и начните обработку команд
func main() {
	bot, err := tgbotapi.NewBotAPI("7626498763:AAGD1LsskYHu8_qgyHi48hsH1lVMjz_xP5k")
	if err != nil {
		log.Panic(err)
	}

	// Выводим информацию о боте
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

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

// Команда для добавления привычки
func handleAddHabit(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	if habitName == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите название привычки после команды /add")
		bot.Send(msg)
		return
	}

	newHabit := Habit{Name: habitName}
	habits[message.Chat.ID] = append(habits[message.Chat.ID], newHabit)

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Привычка '%s' добавлена!", habitName))
	bot.Send(msg)
}

// Команда для отметки выполнения привычки
func handleCompleteHabit(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	if habitName == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите название привычки после команды /complete")
		bot.Send(msg)
		return
	}

	// Поиск привычки по названию
	found := false
	for i, habit := range habits[message.Chat.ID] {
		if habit.Name == habitName {
			found = true
			// Проверка, выполнена ли привычка сегодня
			if time.Now().Sub(habit.LastCompleted).Hours() < 24 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Эта привычка уже отмечена на сегодня.")
				bot.Send(msg)
				return
			}

			// Обновление данных привычки
			habits[message.Chat.ID][i].Count++
			habits[message.Chat.ID][i].LastCompleted = time.Now()

			// Обновляем серию (стрик) привычки
			if time.Now().Sub(habit.LastCompleted).Hours() < 48 {
				habits[message.Chat.ID][i].Streak++
			} else {
				habits[message.Chat.ID][i].Streak = 1
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Привычка '%s' отмечена! Счётчик: %d", habitName, habits[message.Chat.ID][i].Count))
			bot.Send(msg)
			break
		}
	}

	if !found {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка не найдена.")
		bot.Send(msg)
	}
}

// Команда для просмотра статистики
func handleStats(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	if habitName == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите название привычки после команды /stats")
		bot.Send(msg)
		return
	}

	// Поиск привычки по названию
	found := false
	for _, habit := range habits[message.Chat.ID] {
		if habit.Name == habitName {
			found = true
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Привычка '%s'\nВсего выполнений: %d\nСерия: %d дней", habit.Name, habit.Count, habit.Streak))
			bot.Send(msg)
			break
		}
	}

	if !found {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка не найдена.")
		bot.Send(msg)
	}
}

// Команда для установки напоминания
func handleReminder(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	habitName := message.CommandArguments()
	if habitName == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите название привычки после команды /reminder")
		bot.Send(msg)
		return
	}

	// Поиск привычки и установка напоминания
	for i, habit := range habits[message.Chat.ID] {
		if habit.Name == habitName {
			habits[message.Chat.ID][i].Reminder = !habits[message.Chat.ID][i].Reminder
			status := "отключены"
			if habits[message.Chat.ID][i].Reminder {
				status = "включены"
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Напоминания для привычки '%s' %s.", habit.Name, status))
			bot.Send(msg)
			return
		}
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Привычка не найдена.")
	bot.Send(msg)
}

// Пример функции, которая могла бы отправлять напоминания (используется в будущем)
func sendReminders(bot *tgbotapi.BotAPI) {
	for chatID, userHabits := range habits {
		for _, habit := range userHabits {
			if habit.Reminder {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Не забудьте выполнить привычку '%s' сегодня!", habit.Name))
				bot.Send(msg)
			}
		}
	}
}

