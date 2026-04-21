import { check, sleep } from 'k6'
import http from 'k6/http'

const BASE_URL = 'http://localhost:8081'

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

const accountIds = [uuidv4(), uuidv4(), uuidv4()]

const transactionTypes = ['income', 'expense']

// Настройки нагрузки
export const options = {
  stages: [
    { duration: '5s', target: 20 }, // Разгоняемся до 20 виртуальных юзеров за 5 сек
    { duration: '30s', target: 20 }, // Держим нагрузку 30 секунд
    { duration: '5s', target: 0 }, // Плавно гасим до 0
  ],
}

export default function () {
  const externalId = uuidv4()
  const accountId = accountIds[Math.floor(Math.random() * accountIds.length)]

  const amount = parseFloat((Math.random() * 500 + 10).toFixed(2))

  const txType = transactionTypes[Math.floor(Math.random() * transactionTypes.length)]

  const payload = JSON.stringify({
    external_id: externalId,
    account_id: accountId,
    amount: amount,
    transaction_type: txType,
  })

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  }

  const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, params)

  check(res, {
    'status is 201 (Created)': (r) => r.status === 201,
    'status is 400 (Insufficient Balance)': (r) => r.status === 400,
    'status is NOT 500': (r) => r.status !== 500,
  })

  // Маленькая пауза, чтобы не повесить себе систему на 100% CPU
  sleep(0.05)
}
