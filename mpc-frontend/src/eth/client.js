import { INFURA_PROJECTID } from '../constants/application'
import Web3 from 'web3'

const getEthProvider = (network) => {
  return new Web3.providers.HttpProvider(`https://${network}.infura.io/v3/${INFURA_PROJECTID}`);
}

export const getEthClient = (network) => {
  return new Web3(getEthProvider(network));
}
