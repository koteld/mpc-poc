import {API_URL} from '../constants/application'

const sse = new EventSource(`${API_URL}/sse`)

const getSSE = () => {
  return sse
}

export default getSSE
