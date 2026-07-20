# Apex — Variant A: Freemium + Pro Subscription

Этот документ фиксирует выбранный вариант коммерциализации, чтобы не терять контекст между чатами.

## 1) Продуктовое решение

**Модель:** freemium-подписка.

- **Free**: базовые функции, достаточные для регулярного использования.
- **Pro**: расширенные функции для power users, которым важны скорость подготовки к гонкам и глубина планирования.

Цель: сделать платный апгрейд естественным продолжением существующего user flow, а не отдельным сложным продуктом.

## 2) Позиционирование тарифа Pro

Pro продается как ускоритель результата:

- меньше ручной рутины перед гонкой;
- выше качество подготовки сетапов/плана;
- больше контроля над сезоном.

## 3) Границы Free и Pro (первая версия)

### Free

- Fuel planner (базовый).
- Goals.
- Базовый season planner.
- Setups showroom (сохранение/просмотр).
- Генерация **baseline** сетапа.

### Pro

- Генерация **setup pack** (skill × session).
- Расширенные сценарии planner (полный режим/доп. ограничения, поэтапно).
- Расширенные облачные возможности и будущие pro-модули.

## 4) Техническая целевая архитектура подписки

### 4.1 Backend

- Слой `subscription`/`billing` с проверкой entitlement (`free`/`pro`).
- API:
  - `GET /api/billing/plans` — публичный список планов.
  - `GET /api/billing/subscription` — текущий статус пользователя.
- Middleware/guard для pro-фич (`requirePro`).
- Feature gating в чувствительных endpoint (первый кандидат: setup pack generator).

### 4.2 Data model

Минимально необходимые сущности:

- Текущий tier у пользователя (`users.plan_tier`).
- История/состояние подписки (`billing_subscriptions`).
- Журнал провайдер-событий (`billing_events`) для идемпотентной обработки webhook.

### 4.3 Frontend

- Upgrade page (`/upgrade`) с:
  - сравнением Free vs Pro;
  - текущим статусом тарифа;
  - CTA на апгрейд.
- UI-gating Pro-фич:
  - понятный disabled-state;
  - объяснение, что функция доступна в Pro;
  - ссылка на `/upgrade`.

### 4.4 Billing provider (целевая интеграция)

Stripe:

- Checkout Session (покупка).
- Customer Portal (управление подпиской).
- Webhooks (истина о статусе подписки в системе).

## 5) События и метрики

Минимальный набор событий:

- `upgrade_page_viewed`
- `upgrade_cta_clicked`
- `checkout_started`
- `checkout_completed`
- `subscription_activated`
- `subscription_canceled`
- `pro_feature_blocked`

Цель: измерять воронку и точки потерь.

## 6) Правила доступа (entitlements)

- `free` — доступ только к free-функциям.
- `pro` + активная подписка — доступ к pro-функциям.
- При спорных состояниях (`past_due`, `incomplete`) поведение определяется на backend (явный deny + user-friendly объяснение).

## 7) Этапы внедрения

### Phase 1 — Foundation

- Схема БД для подписок.
- Backend entitlement service.
- API планов и текущей подписки.
- Первое pro-gating (setup pack).
- Базовый upgrade UI.

### Phase 2 — Stripe

- Checkout + Portal.
- Webhook ingestion и синхронизация статусов.
- Статусы подписки в UI.

### Phase 3 — Growth

- A/B эксперименты paywall/packaging.
- Расширение pro-набора в planner и других модулях.
- Улучшение retention через lifecycle-цепочки.

## 8) Критерии готовности (Definition of Done для Variant A MVP)

- Пользователь видит свой план (`free`/`pro`) в UI.
- Pro-фича технически защищена на backend, не только в UI.
- Есть рабочий путь апгрейда (сначала foundation, затем Stripe).
- Все изменения покрыты автопроверками проекта.

---

**Текущий статус:** принята стратегия Variant A, начинается реализация Phase 1 (foundation).
