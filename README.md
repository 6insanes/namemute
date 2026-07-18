# NameMute

Мод для Sid Meier's Civilization V (Brave New World), который скрывает реальный
никнейм игрока во внутриигровых экранах мультиплеера и показывает вместо него
короткое название цивилизации, за которую он играет. В комплекте — установщик.

## Требования

- Civilization V с дополнением **Brave New World**.

## Что делает мод

Заменяет вывод `Player:GetNickName()` на короткое название цивилизации (или на
имя лидера — там, где рядом уже отдельно выводится цивилизация, чтобы не
дублировать) в следующих местах:

- дипломатический попап и диалог с лидером;
- дипломатический угол и **текстовый чат**;
- панель очереди ходов и список подключений (F8/MP);
- экран сделок и список дипломатии;
- голосование Всемирного конгресса и результаты голосований Лиги;
- обзоры отношений;
- экран шпионажа;
- F8 «Демография»/«Прогресс побед»;
- тултипы юнитов и владения клетками;
- флажки юнитов;
- меню паузы.

## Как работает установщик

По умолчанию установщик копирует файлы мода внутрь `Assets/DLC/Expansion2`:
файлы, которые в оригинале уже лежат внутри `Expansion2`, заменяются на месте
(с бэкапом оригинала); файлы, которых там изначально нет, добавляются как
новые, не трогая никакие оригинальные файлы игры.

Есть второй режим — `-direct-replace`: подменяет все 17 файлов мода прямо в
их родных путях (`Assets/UI/...`, `Assets/Gameplay/...` и т.д.), в обход
Expansion2. Включается флагом `-direct-replace`.

В обоих режимах есть `-uninstall` для отката (восстанавливает оригиналы,
удаляет добавленные файлы) и `-game-dir` для ручного указания пути к игре,
если установщик не нашёл её сам.

## Установка

Готовые самодостаточные исполняемые файлы (~2.4 МБ, все файлы мода зашиты
внутрь на этапе сборки, никакие папки рядом не нужны) — на странице
[Releases](https://github.com/6insanes/namemute/releases/latest):

- `namemute-linux` — для Linux (нативный порт Aspyr и/или Proton)
- `namemute-windows.exe` — для Windows

Скачайте нужный файл (сам файл, ничего больше) и запустите:

- **Linux:** `chmod +x namemute-linux && ./namemute-linux`
- **Windows:** двойной клик по `namemute-windows.exe`, либо из консоли (см. ниже)

Установщик сам находит установку Civilization V (перебирая все библиотеки
Steam, включая Proton). Подтвердите установку (`y`), полностью закройте игру
(если запущена) и запустите заново — готово, ничего включать в меню Mods не
нужно.

### Запуск из консоли на Windows

Двойной клик достаточен для установки по умолчанию, но чтобы передать флаги
(`-uninstall`, `-direct-replace`, `-game-dir`) или увидеть вывод после
завершения, установщик нужно запускать из консоли:

1. Откройте папку, куда скачали `namemute-windows.exe`, в Проводнике.
2. Откройте консоль прямо в этой папке одним из способов:
   - зажмите **Shift**, кликните правой кнопкой мыши по пустому месту в папке
     и выберите **«Открыть окно PowerShell здесь»** (или «Открыть окно команд
     здесь» — в зависимости от версии Windows);
   - либо кликните в адресной строке Проводника (где написан путь к папке),
     введите `powershell` или `cmd` и нажмите **Enter** — откроется консоль
     уже в этой папке;
   - либо нажмите **Win**, наберите `powershell` (или `cmd`), откройте
     обычным способом и перейдите в папку командой `cd`, например:
     ```
     cd "$env:USERPROFILE\Downloads"
     ```
     (путь зависит от того, куда браузер сохранил файл — обычно это папка
     «Загрузки»).
3. Запустите установщик. В PowerShell перед именем файла нужно `.\`:
   ```
   .\namemute-windows.exe
   ```
   В `cmd.exe` (командной строке) `.\` не нужен:
   ```
   namemute-windows.exe
   ```

Если Windows Defender SmartScreen блокирует запуск («Windows защитила ваш
компьютер») — это стандартная реакция на неподписанный exe-файл. Нажмите
**«Подробнее»**, затем **«Выполнить в любом случае»**.

Если установщик не нашёл игру автоматически:

```
./namemute-linux -game-dir "/путь/к/Sid Meier's Civilization V"
```

```
.\namemute-windows.exe -game-dir "C:\Program Files (x86)\Steam\steamapps\common\Sid Meier's Civilization V"
```

Прямая замена файлов вместо инъекции в Expansion2:

```
./namemute-linux -direct-replace
```

```
.\namemute-windows.exe -direct-replace
```

Откатить изменения (восстановить все оригинальные файлы, удалить добавленные):

```
./namemute-linux -uninstall
```

```
.\namemute-windows.exe -uninstall
```

Установщик идемпотентен — повторный запуск не подменяет файлы дважды.
Учтите: обновление Civilization V через Steam может перезаписать
подменённые/добавленные файлы — после крупного обновления игры запустите
установщик ещё раз.

## Ручная установка (без установщика)

Если не хотите запускать исполняемый файл, можно установить мод руками —
просто скопировав 17 `.lua`-файлов из папки [`assets/`](assets) поверх
соответствующих файлов игры. Так как в GitHub Releases лежит только сам
установщик, для ручной установки нужны файлы мода — склонируйте репозиторий
или скачайте [`assets/`](assets) с GitHub (Code → Download ZIP).

1. Найдите папку установки Civilization V — там, где лежит подпапка `Assets`
   (обычно `.../steamapps/common/Sid Meier's Civilization V`).
2. Полностью закройте игру, если она запущена.
3. Для каждого файла из таблицы ниже: сделайте резервную копию оригинала
   (например, добавив к имени `.bak`), затем скопируйте на его место
   одноимённый файл из `assets/`, заменив содержимое.
4. Запустите игру — ничего включать в меню Mods не нужно.

| Файл мода (`assets/…`)        | Заменяет файл игры (относительно папки установки)                    |
|--------------------------------|------------------------------------------------------------------------|
| `GameplayUtilities.lua`        | `Assets/Gameplay/Lua/GameplayUtilities.lua`                            |
| `DiscussionDialog.lua`         | `Assets/DLC/Expansion2/UI/InGame/LeaderHead/DiscussionDialog.lua`      |
| `DiploCorner.lua`              | `Assets/DLC/Expansion2/UI/InGame/WorldView/DiploCorner.lua`            |
| `MPTurnPanel.lua`              | `Assets/UI/InGame/WorldView/MPTurnPanel.lua`                           |
| `MPList.lua`                   | `Assets/UI/InGame/WorldView/MPList.lua`                                |
| `TradeLogic.lua`               | `Assets/DLC/Expansion2/UI/InGame/WorldView/TradeLogic.lua`             |
| `DiploList.lua`                | `Assets/DLC/Expansion2/UI/InGame/DiploList.lua`                        |
| `DiploCurrentDeals.lua`        | `Assets/UI/InGame/Popups/DiploCurrentDeals.lua`                        |
| `DiploVotePopup.lua`           | `Assets/DLC/Expansion2/UI/InGame/Popups/DiploVotePopup.lua`            |
| `DiploRelationships.lua`       | `Assets/DLC/Expansion2/UI/InGame/Popups/DiploRelationships.lua`        |
| `DiploGlobalRelationships.lua` | `Assets/DLC/Expansion2/UI/InGame/Popups/DiploGlobalRelationships.lua`  |
| `VoteResultsPopup.lua`         | `Assets/DLC/Expansion2/UI/InGame/Popups/VoteResultsPopup.lua`          |
| `EspionageOverview.lua`        | `Assets/DLC/Expansion2/UI/InGame/Popups/EspionageOverview.lua`         |
| `VictoryProgress.lua`          | `Assets/DLC/Expansion2/UI/InGame/Popups/VictoryProgress.lua`           |
| `PlotMouseOverInclude.lua`     | `Assets/DLC/Expansion2/UI/InGame/PlotMouseoverInclude.lua`             |
| `UnitFlagManager.lua`          | `Assets/DLC/Expansion2/UI/InGame/UnitFlagManager.lua`                  |
| `GameMenu.lua`                 | `Assets/UI/InGame/Menus/GameMenu.lua`                                  |

Это соответствует режиму `-direct-replace` установщика — файлы подменяются
прямо в родных путях игры, в обход Expansion2.

На Linux (нативный порт Aspyr, не Proton) те же пути живут в нижнем регистре
под `steamassets/assets/...` (например,
`steamassets/assets/gameplay/lua/gameplayutilities.lua`) — используйте их,
если папки `Assets/...` в верхнем регистре не существует.

Мод не трогает сохранения (`AffectsSavedGames=0`) и не меняет игровую базу
данных — это чистый UI-мод, поэтому замена файлов безопасна. Чтобы откатить
установку, верните сохранённые на шаге 3 резервные копии на место.

## Структура проекта

```
main.go     — установщик (Go, кросс-компилируется под Linux/Windows)
go.mod
assets/     — файлы мода (.lua), встраиваются в установщик через go:embed
dist/       — собранные бинарники установщика
```

Пересборка бинарников:

```
GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/namemute-linux .
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/namemute-windows.exe .
```
