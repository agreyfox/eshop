import store from '@/store'
import router from '@/router'
import { Base64 } from 'js-base64'
//import { baseURL } from '@/utils/constants'
import { apiServer } from './constants'


export function parseToken(token) {

  if (!token) {
    return
  }
  const parts = token.split('.')

  if (parts.length !== 3) {
    throw new Error('token malformed')
  }

  const data = JSON.parse(Base64.decode(parts[1]))

  if (Math.round(new Date().getTime() / 1000) > data.exp) {
    throw new Error('token expired')
  }
  console.log(data);

  localStorage.setItem('jwt', token)
  store.commit('setJWT', token)
  store.commit('setUser', data.user)
}

export async function validateLogin() {
  try {
    if (localStorage.getItem('jwt')) {
      await renew(localStorage.getItem('jwt'))
    }
  } catch (_) {
    console.warn('Invalid JWT token in storage') // eslint-disable-line
  }
}

export async function login(email, password) {
  if (!email) {
    return
  }
  const res = await fetch(`${apiServer}/api/user/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ "email": email, "password": password })
  })

  const body = await res.json();
  // console.log(body)
  if (res.status === 200 && body.retCode >= 0) {
    /* localStorage.setItem('jwt', body.data)
    store.commit('setJWT', body.data)
    store.commit('setUser', user) */
    parseToken(body.data);
  } else {
    throw new Error(body)
  }
}

export async function renew(jwt) {
  const res = await fetch(`${apiServer}/api/renew`, {
    method: 'POST',
    headers: {
      'X-Auth': jwt,
    }
  })

  const body = await res.text()

  if (res.status === 200) {
    parseToken(body)
  } else {
    throw new Error(body)
  }
}

export async function signup(username, password) {
  const data = { username, password }

  const res = await fetch(`${apiServer}/api/signup`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  })

  if (res.status !== 200) {
    throw new Error(res.status)
  }
}

export function logout() {
  store.commit('setJWT', '')
  store.commit('setUser', null)
  localStorage.setItem('jwt', null)
  router.push({ path: '/login' })
}
