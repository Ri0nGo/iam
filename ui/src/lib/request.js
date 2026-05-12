const API_BASE = import.meta.env.VITE_API_BASE || '/api/v1'
export const AUTH_INVALID_EVENT = 'iam:auth-invalid'

function buildHeaders(token, extraHeaders = {}) {
  const headers = { ...extraHeaders }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  return headers
}

export async function request(path, options = {}) {
  const { token, isOAuthTokenEndpoint = false, ...rest } = options
  const response = await fetch(`${API_BASE}${path}`, {
    ...rest,
    headers: buildHeaders(token, rest.headers),
  })

  const contentType = response.headers.get('content-type') || ''
  const body = contentType.includes('application/json') ? await response.json() : await response.text()

  if (!response.ok) {
    const message = typeof body === 'string'
      ? body
      : body?.message || body?.error_description || body?.error || 'Request failed'
    if (response.status === 401 || message.toLowerCase().includes('invalid token')) {
      window.dispatchEvent(new Event(AUTH_INVALID_EVENT))
    }
    throw new Error(message)
  }

  if (isOAuthTokenEndpoint) {
    return body
  }

  return body.data
}

export const api = {
  login(payload) {
    return request('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  logout(token) {
    return request('/auth/logout', { method: 'POST', token })
  },
  me(token) {
    return request('/auth/me', { token })
  },
  oauthToken(payload) {
    return request('/oauth/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      isOAuthTokenEndpoint: true,
    })
  },
  oauthUserinfo(token) {
    return request('/oauth/userinfo', { token })
  },
  listAuthApplications(token, params = {}) {
    const query = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        query.set(key, String(value))
      }
    })
    return request(`/auth-applications${query.toString() ? `?${query.toString()}` : ''}`, { token })
  },
  getAuthApplication(token, id) {
    return request(`/auth-applications/${id}`, { token })
  },
  createAuthApplication(token, payload) {
    return request('/auth-applications', {
      method: 'POST',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  updateAuthApplication(token, id, payload) {
    return request(`/auth-applications/${id}`, {
      method: 'PUT',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  deleteAuthApplication(token, id) {
    return request(`/auth-applications/${id}`, { method: 'DELETE', token })
  },
  listUsers(token, params = {}) {
    const query = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        query.set(key, String(value))
      }
    })
    return request(`/users${query.toString() ? `?${query.toString()}` : ''}`, { token })
  },
  getUser(token, id) {
    return request(`/users/${id}`, { token })
  },
  createUser(token, payload) {
    return request('/users', {
      method: 'POST',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  updateUserStatus(token, id, payload) {
    return request(`/users/${id}/status`, {
      method: 'PUT',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  deleteUser(token, id) {
    return request(`/users/${id}`, { method: 'DELETE', token })
  },
  resetPassword(token, id, payload) {
    return request(`/users/${id}/password`, {
      method: 'PUT',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  listRoles(token) {
    return request('/roles', { token })
  },
  createRole(token, payload) {
    return request('/roles', {
      method: 'POST',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  bindUserRoles(token, userId, payload) {
    return request(`/users/${userId}/roles`, {
      method: 'PUT',
      token,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
  },
  getUserRoles(token, userId) {
    return request(`/users/${userId}/roles`, { token })
  },
}
