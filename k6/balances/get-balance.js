import { check, sleep } from 'k6'
import http from 'k6/http'

// Порт баланс-вьювера, согласно твоему .env файлу
const BASE_URL = 'http://localhost:8084'

// Функция для генерации случайного UUID, если нам понадобятся несуществующие аккаунты
function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

// ID аккаунтов. В идеале сюда стоит вставить UUID, для которых УЖЕ были сгенерированы
// транзакции предыдущим скриптом create-transaction.js.
// Либо можно просто генерировать случайные (тогда API будет отдавать 404).
const accountIds = [
  'e997ae0d-f763-4036-91a3-0ac82f700ffa',
  '877d5c2f-2646-4046-8c35-650322bc56a0',
  'f1c02c03-6265-49e3-b23a-10a6231eefe0',
  uuidv4(),
]

// Настройки нагрузки
export const options = {
  stages: [
    { duration: '5s', target: 50 }, // Разгоняемся до 50 виртуальных юзеров за 5 сек (для чтения можно поставить больше)
    { duration: '30s', target: 50 }, // Держим нагрузку 30 секунд
    { duration: '5s', target: 0 }, // Плавно гасим до 0
  ],
}

export default function () {
  // Выбираем случайный аккаунт из списка
  const accountId = accountIds[Math.floor(Math.random() * accountIds.length)]

  // Отправляем GET запрос
  const res = http.get(`${BASE_URL}/api/v1/account/${accountId}/balance`)

  // Проверяем ответы
  check(res, {
    'status is 200 (OK)': (r) => r.status === 200,
    'status is 404 (Not Found)': (r) => r.status === 404,
    'status is NOT 500': (r) => r.status !== 500,
    'response has current_balance (if 200)': (r) => {
      if (r.status === 200) {
        try {
          const body = JSON.parse(r.body)
          return body.hasOwnProperty('current_balance')
        } catch (e) {
          return false
        }
      }
      return true // Пропускаем проверку тела, если статус 404
    },
  })

  // Маленькая пауза
  sleep(0.05)
}
