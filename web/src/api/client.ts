import axios from 'axios'

export interface ApiErrorBody {
  error: {
    code: string
    message: string
    details?: any
  }
}

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

client.interceptors.response.use(
  (resp) => resp,
  (error) => {
    if (error.response?.data?.error) {
      return Promise.reject(new Error(error.response.data.error.message))
    }
    return Promise.reject(error)
  },
)

export default client
