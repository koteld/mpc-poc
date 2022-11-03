import axios from 'axios';
import { API_URL } from '../constants/application';

axios.defaults.baseURL = API_URL;
axios.defaults.timeout = 60000;
// axios.defaults.headers.common['X-Requested-With'] = 'XMLHttpRequest'
// axios.defaults.headers.common['Access-Control-Allow-Origin'] = '*'

export default axios;
export { API_URL };
