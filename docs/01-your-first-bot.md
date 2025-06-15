# `1` Ваш первый бот

Для начала вам нужно получить токен бота. Для этого откройте диалог с [MasterBot](https://max.ru/primebot) и создайте нового бота, следуя инструкциям. После этого PrimeBot отправит вам токен.

Создайте новый проект `my-first-bot` и установите `github.com/xenon007/max-bot-api-client-go`. Для этого откройте терминал и выполните следующие команды:
```sh
# Создайте новую папку для исходного кода вашего модуля Go и перейдите в неё
mkdir my-first-bot
cd my-first-bot 
# Запустите свой модуль с помощью команды go mod init.
go mod init first-max-bot

# Установите библиотеку для работы с MACX API на golang
go get github.com/xenon007/max-bot-api-client-go
```

Команда go mod init создает файл go.mod для отслеживания зависимостей вашего кода. Пока что файл включает только имя вашего модуля и версию Go, которую поддерживает ваш код.

Теперь создайте файл, например, `bot.go`. 

Весь код в языке Go организуется в пакеты. Пакеты представляют удобную организацию разделения кода на отдельные части или модули. Модульность позволяет определять один раз пакет с нужной функциональностью и потом использовать его многкратно в различных программах.
Код пакета располагается в одном или нескольких файлах с расширением go. Для определения пакета применяется ключевое слово package. Поэтому наш файл `bot.go` будет иметь следующую структуру. 

```go 
package main
import "fmt"
 
func main() {
     
    fmt.Println("Hello Max Bot Go")
}
```

В данном случае пакет называется main. Определение пакета должно идти в начале файла.

Есть два типа пакетов: исполняемые (executable) и библиотеки (reusable). Для создания исполняемых файлов пакет должен иметь имя main. Все остальные пакеты не являются исполняемыми. При этом пакет main должен содержать функцию main, которая является входной точкой в приложение.

Импортируем в наш пакет main установленный модуль `github.com/xenon007/max-bot-api-client-go`  


```go 
package main
import (
    "fmt"
	maxbot "github.com/xenon007/max-bot-api-client-go"
)

func main() {
	api := maxbot.New(os.Getenv("TOKEN"))
	// Some methods demo:
	info, err := api.Bots.GetBot()
	fmt.Printf("Get me: %#v %#v", info, err)
}
```

Код выше, создаюет объект `api`, передавая токен в его конструктор New. Мы рекомендуем передавать токен через переменные окружения, т.к. использовать токен в коде - плохая практика.

Данная программы выведет только информацию о вашем боте и закончит работу. 
Чтобы бот заработал необходимы обработчик событий из калана с обновлениями

```go 
package main
import (
    "fmt"
	maxbot "github.com/xenon007/max-bot-api-client-go"
)

func main() {
	api := maxbot.New(os.Getenv("TOKEN"))
	// Some methods demo:
	info, err := api.Bots.GetBot()
	fmt.Printf("Get me: %#v %#v", info, err)

	ctx, cancel := context.WithCancel(context.Background()) // создам 
	go func() {
		exit := make(chan os.Signal)
		signal.Notify(exit, os.Kill, os.Interrupt)
		<-exit
		cancel()
	}()

	for upd := range api.GetUpdates(ctx) { // Чтение из канала с обновлениями
		switch upd := upd.(type) { // Определение типа пришедшего обновления
		case *schemes.MessageCreatedUpdate:
			// Отправка сообщения 
			_, err := api.Messages.Send(maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Hello from Bot"))
        }
    }
}
```
Поздравляем, вы написали первого бота! 🎉
