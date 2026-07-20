# Content Operations Blueprint (YouTube + TikTok + Telegram + Boosty)

Статус: draft v1  
Владелец: you  
Цель документа: не потерять стратегию и перейти от хаотичного продакшна к системной контент-операционке, которую можно сначала использовать самому, а потом потенциально продавать.

---

## 1) Контекст и цель

Ты строишь мультиканальный контент-движок:

- YouTube (long + shorts),
- TikTok,
- Telegram,
- Boosty.

Текущий приоритет: сделать систему, которая **в первую очередь полезна тебе как автору**, даже если не будет сразу коммерчески успешной.

Главный принцип:  
**Product for self-use first -> product for market second.**

---

## 2) Текущий процесс (as-is)

Ниже то, как у тебя уже устроен pipeline:

1. В Claude Projects формируется контекст канала:
   - тематика,
   - название,
   - описание.
2. Готовятся промпты для изображений.
3. В GPT валидируются/докручиваются изображения.
4. Создается банк тем.
5. Выбирается тема.
6. Пишется сценарий, дальше несколько итераций до "готово".
7. Озвучка и монтаж вручную.
8. Генерация title + thumbnail (частично через GPT + Pinterest).
9. Создание описания к видео.
10. Резка long-видео в shorts/tiktoks.
11. Часть материала уходит в Telegram.
12. Boosty пока не определен (идея: расширенный контент / доп-видео).

---

## 3) Основные боли (что реально тормозит)

### 3.1 Разрыв между этапами
Сценарий, превью, публикация, шорты, посты живут в разных местах.  
Нет единого "root object", который связывает всё вокруг одного эпизода.

### 3.2 Слабый feedback loop
Метрики от публикаций не возвращаются обратно в процесс выбора новых тем/хуков.

### 3.3 Нет repeatable templates
Много ручных и повторяющихся действий на каждом эпизоде.

### 3.4 Неопределенность по Boosty
Нет четкой product ladder: "что дается бесплатно, что платно, зачем человеку подписка".

### 3.5 Высокая cognitive load
Много решений "с нуля" на каждом цикле вместо работы по шаблону.

---

## 4) Целевой operating model (to-be)

### 4.1 Основная сущность: `Content Episode`

Каждый эпизод (long-видео) - корневая сущность, к которой прикрепляется всё:

- идея/гипотеза,
- сценарий (версии),
- assets (превью, графика),
- публикации по каналам,
- производные units (shorts/tiktoks/telegram/boosty),
- метрики.

### 4.2 Каналы = дистрибуция одной content unit
Не "делать 5 разных контентов", а:

**1 long-video -> N derivatives -> публикация по каналам по плану.**

### 4.3 Процесс как pipeline с Definition of Done

Для каждого этапа заранее фиксируется "когда считается готовым":

- Ideation
- Script
- Production
- Packaging
- Publishing
- Repurposing
- Analysis

---

## 5) Рекомендуемый workflow (новый)

## 5.1 Stage 0: Channel Strategy
Что фиксируем:

- ICP (кто зритель),
- обещание канала (value promise),
- pillar topics (3-5 опорных тем),
- tone/style guide,
- KPI на 90 дней.

Выход этапа:
- channel brief v1.

## 5.2 Stage 1: Idea Intake & Prioritization
Любая идея попадает в backlog с полями:

- формат (long/short/post),
- цель (охват/подписки/продажа),
- "боль аудитории",
- confidence score.

Отбор в продакшн по простому score:
`Impact x Confidence x Speed`.

## 5.3 Stage 2: Script System
Для long-видео сценарий строится по шаблону:

- Hook (0-15 сек),
- Problem framing,
- Core value blocks,
- CTA,
- Repurpose markers (таймкоды потенциальных short-моментов).

Важно: сразу в сценарии отмечать сегменты, которые потом уйдут в short-form.

## 5.4 Stage 3: Production
Чеклист:

- запись аудио,
- монтаж long,
- экспорт master,
- сохранение project files/исходников.

## 5.5 Stage 4: Packaging
Параллельно:

- 3-5 вариантов title,
- 2-3 концепта thumbnail,
- description + chapter marks,
- pinned comment/callout plan.

## 5.6 Stage 5: Distribution
Для каждого канала отдельный publish card:

- YouTube long,
- YouTube Shorts,
- TikTok,
- Telegram post,
- Boosty post.

С таймингом (дата/время) и статусом публикации.

## 5.7 Stage 6: Repurpose Factory
Из long-видео формируется пакет:

- 3-7 коротких клипов,
- 1 Telegram пост "основная мысль + CTA",
- 1 Boosty-расширение (см. раздел Boosty).

## 5.8 Stage 7: Review & Learn
На D+1 / D+3 / D+7:

- собираем KPI,
- фиксируем "что сработало / что нет",
- создаем follow-up ideas.

---

## 6) Boosty: как убрать неопределенность

## 6.1 Лестница ценности (free -> paid)

Free (публично):
- основной long-контент,
- часть shorts/тизеров.

Paid (Boosty):
- extended cut / доп.пример,
- backstage (процесс и рабочие заметки),
- шаблоны/чеклисты/ресурсы из выпуска,
- ранний доступ к выпускам,
- monthly Q&A / разборы.

## 6.2 Минимальная Boosty-модель для старта

Уровень 1 (base support):
- early access + закрытый пост.

Уровень 2 (value tier):
- extended notes + resources + behind-the-scenes.

Уровень 3 (premium):
- periodic mini-consult / персональные ответы / воркшоп.

---

## 7) Data model (MVP)

Ниже сущности, которые стоит держать в системе:

- `Workspace`
- `Channel`
- `TopicPillar`
- `Idea`
- `Episode`
- `ScriptVersion`
- `Asset` (image/video/audio/project)
- `Derivative` (short/tiktok/post/boosty unit)
- `PublishTarget` (канал + слот + статус)
- `Task`
- `MetricSnapshot`
- `Experiment`
- `RetrospectiveNote`

### Ключевые связи

- `Episode` <-many- `ScriptVersion`
- `Episode` <-many- `Asset`
- `Episode` <-many- `Derivative`
- `Derivative` <-one/many- `PublishTarget`
- `Episode` <-many- `MetricSnapshot`

---

## 8) Automation roadmap

## Phase A - Manual-first system (быстро)
Автоматизируем только структуру:

- автосоздание чеклистов по эпизоду,
- автосоздание derivative tasks,
- календарный board по publish targets.

## Phase B - Assisted automation

- AI-помощь для title/description/hooks,
- AI-подсказки коротких клипов на базе сценария/таймкодов,
- auto-brief для Telegram/Boosty из long script.

## Phase C - API integrations

- YouTube Data API (метрики, статусы публикации),
- (опционально) TikTok API/ручной импорт,
- Telegram Bot/API для semi-auto публикации.

---

## 9) MVP scope (для себя, 4-6 недель)

### MVP-1 (обязательно)

1. Episode как root entity  
2. Script versions  
3. Pipeline statuses  
4. Derivative planner  
5. Multi-channel publish calendar  
6. Weekly review board  

### MVP-2 (после MVP-1)

1. YouTube metrics import  
2. Performance dashboard  
3. Template library (hooks, episode formats)  
4. Boosty content pack templates  

---

## 10) Метрики успеха (self-use first)

### Operating metrics

- `idea_to_publish_days` (медиана)
- `% episodes with full derivative pack`
- `% on-time publishes`
- `batch efficiency` (сколько единиц за сессию)

### Outcome metrics

- views/subscribers growth
- watch-time
- conversion to Boosty (если включен)

### Personal sustainability

- часов в неделю на продакшн
- субъективная нагрузка/выгорание (self-score)

---

## 11) Режим работы (рекомендуемый)

### Еженедельный цикл

- Пн: backlog triage + план недели
- Вт-Ср: сценарии + запись
- Чт: монтаж + упаковка
- Пт: публикации + repurpose pack
- Вс: review (метрики + выводы)

### Ежемесячный цикл

- аудит pillar topics,
- обновление шаблонов,
- пересмотр Boosty-оффера.

---

## 12) Риски и как снижать

1. **Слишком широкий MVP**  
   -> Держать фокус: long -> derivatives -> publish -> review.

2. **Переавтоматизация до PMF**  
   -> Сначала manual-first + шаблоны.

3. **Платформа ради платформы**  
   -> Каждая фича должна экономить время или повышать output.

4. **Неясный Boosty value prop**  
   -> Ввести контент-лестницу и минимум 1 paid deliverable на эпизод.

---

## 13) Concrete next steps (что делать дальше)

1. Зафиксировать 1 канал как pilot (не все сразу).  
2. Завести первые 10 `Episode` карточек по новой модели.  
3. На каждом эпизоде делать минимум 3 derivative unit.  
4. Через 2 недели ретро: что мешает, что автоматизировать первым.  
5. После 4 недель решить: оставаться self-use или открывать beta-доступ.

---

## 14) Productization decision gate (через 4-6 недель)

Запускать в рынок имеет смысл, если:

- ты пользуешься системой стабильно каждую неделю,
- time-to-publish заметно сократился,
- output и consistency выросли,
- есть хотя бы несколько повторяемых сценариев ценности.

Если нет - оставляем как внутренний toolchain без давления на монетизацию.

---

## 15) Важный принцип

Не "строить ещё один контент-планировщик", а строить  
**Content Operating System для автора с мультиканальной дистрибуцией**.

Это и будет твоим углом отличия.
