import {API_URL} from '../constants/application'

const sse = new EventSource(`${API_URL}/sse`)

export default () => {
  return sse
}
