import {
  Button,
  Box,
  Container,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  TextField,
  Card,
  Typography,
  CardContent,
  Divider,
  Alert,
  Snackbar,
  Link,
  IconButton,
  LinearProgress
} from '@mui/material'
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CachedIcon from '@mui/icons-material/Cached';
import {useEffect, useState} from 'react'
import Logs from './components/Logs'
import web3 from 'web3'
import BN from 'bn.js'
import {SEVERITIES, NETWORKS} from './constants/application'
import {getEthClient} from './eth/client'
import {getConfigs} from './api/api.info'
import {sendETH, startDKF, startDKG} from './api/api.protocol'

function App() {
  const [isLoading, setIsLoading] = useState(false)
  const [configAddress, setConfigAddress] = useState("")
  const [configs, setConfigs] = useState([])
  const [eth, setEth] = useState("")
  const [address, setAddress] = useState("")
  const [isAddressValid, setIsAddressValid] = useState(false)
  const [openSnackbar, setOpenSnackbar] = useState(false)
  const [snackbarMessage, setSnackbarMessage] = useState("")
  const [snackbarSeverity, setSnackbarSeverity] = useState(SEVERITIES.SUCCESS)
  const [balance, setBalance] = useState("0")
  const [ethClient, setEthClient] = useState(null)
  const [txLink, setTxLink] = useState("")
  
  useEffect(() => {
    setIsLoading(true)
    setEthClient(getEthClient(NETWORKS.GOERLI))
    getConfigs().then((res) => {
      if (!res.error) {
        const c = res.data.reduce((agg, val) => {
          agg[val.address] = {
            ...val,
          }
          return agg
        }, {})
        setConfigs(c)
      } else {
        setSnackbarSeverity(SEVERITIES.ERROR)
        setSnackbarMessage(`Error loading configs: ${res.data}`)
        setOpenSnackbar(true)
      }
      setIsLoading(false)
    })
  }, [])
  
  const handleCloseSnackbar = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    setSnackbarSeverity(SEVERITIES.SUCCESS)
    setSnackbarMessage('')
    setOpenSnackbar(false)
  }
  
  const handleChangeConfig = (event) => {
    setConfigAddress(event.target.value)
    setTxLink('')
    void getBalance(event.target.value)
  }
  
  const handleChangeAddress = (event) => {
    setIsAddressValid(web3.utils.isAddress(event.target.value))
    setAddress(event.target.value)
  }
  
  const handleChangeEth = (event) => {
    const regex = /^[+-]?([0-9]+([.][0-9]*)?|[.][0-9]+)$/
    if (!event.target.value.length || regex.test(event.target.value)) {
      setEth(event.target.value)
    }
  }
  
  const getBalance = async (address) => {
    setBalance(await ethClient.eth.getBalance(address))
  }
  
  const copyAddress = () => {
    navigator.clipboard.writeText(configAddress).then(() => {
      setSnackbarSeverity(SEVERITIES.SUCCESS)
      setSnackbarMessage(`Address successfully copied`)
      setOpenSnackbar(true)
    })
  }
  
  const generateConfig = () => {
    setIsLoading(true)
    startDKG().then((res) => {
      if (!res.error) {
        const nconfig = res.data
        const cconfigs = configs
        cconfigs[nconfig.address] = {
          ...nconfig
        }
        setConfigs(cconfigs)
        
        setSnackbarSeverity(SEVERITIES.SUCCESS)
        setSnackbarMessage("New keys configuration generated successfully")
        setOpenSnackbar(true)
      } else {
        setSnackbarSeverity(SEVERITIES.ERROR)
        setSnackbarMessage(`Error generating config: ${res.data}`)
        setOpenSnackbar(true)
      }
      setIsLoading(false)
    })
  }
  
  const refreshConfig = () => {
    if (!configAddress) {
      return
    }
    setIsLoading(true)
    startDKF(configAddress).then((res) => {
      if (!res.error) {
        const nconfig = res.data
        const cconfigs = configs
        cconfigs[nconfig.address] = {
          ...nconfig
        }
        setConfigs(cconfigs)
        
        setSnackbarSeverity(SEVERITIES.SUCCESS)
        setSnackbarMessage("Keys configuration refreshed successfully")
        setOpenSnackbar(true)
      } else {
        setSnackbarSeverity(SEVERITIES.ERROR)
        setSnackbarMessage(`Error refreshing config: ${res.data}`)
        setOpenSnackbar(true)
      }
      setIsLoading(false)
    })
  }
  
  const sendEth = () => {
    if (!eth) {
      setSnackbarSeverity(SEVERITIES.WARNING)
      setSnackbarMessage('ETH value should be greater than 0')
      setOpenSnackbar(true)
      return
    }
    if (!isAddressValid) {
      setSnackbarSeverity(SEVERITIES.WARNING)
      setSnackbarMessage('Address should be valid ETH address')
      setOpenSnackbar(true)
      return
    }
  
    const amount = web3.utils.toWei(eth, 'ether')
    const amountBN = new BN(amount)
    const balanceBN = new BN(balance)
    if (amountBN.gt(balanceBN)) {
      setSnackbarSeverity(SEVERITIES.WARNING)
      setSnackbarMessage('Balance should be greater than amount to send')
      setOpenSnackbar(true)
      return
    }

    setIsLoading(true)
    setTxLink('')
    setSnackbarMessage('')
    setOpenSnackbar(false)
    
    sendETH(configAddress, address, amount).then((res) => {
      if (!res.error) {
        setTxLink(`https://goerli.etherscan.io/tx/${res.data}`)
        
        setSnackbarSeverity(SEVERITIES.SUCCESS)
        setSnackbarMessage("ETH amount was successfully sent")
        setOpenSnackbar(true)
      } else {
        setSnackbarSeverity(SEVERITIES.ERROR)
        setSnackbarMessage(`Error sending ETH: ${res.data}`)
        setOpenSnackbar(true)
      }
      setIsLoading(false)
    })
  }
  
  return (
    <Box>
      <Container maxWidth="md" sx={{
        p: 2,
        height: "100vh",
        display: "flex",
        flexDirection: "column",
      }}>
        <Paper
          sx={{
            p: 2,
          }}
        >
          <Box>
            <Button
              variant="text"
              onClick={generateConfig}
              disabled={isLoading}
              sx={{
                borderRadius: "10px"
            }}>Generate new wallet</Button>
          </Box>
          <Box sx={{
            mb:2,
          }}>
            <Grid container>
              <Grid item xs={24} md={8} sx={{
                mr: 2,
                maxWidth: '100%'
              }}>
                <FormControl fullWidth variant="standard" sx={{
                  height: "55px",
                  mb: 2,
                  mt: 2
                }}>
                  <InputLabel id="select-label">Wallet</InputLabel>
                  <Select
                    labelId="select-label"
                    id="select"
                    label="wallet"
                    onChange={handleChangeConfig}
                    defaultValue = ""
                  >
                    {Object.values(configs).map((value) =>
                      (<MenuItem key={value.address} value={value.address}>{value.address}</MenuItem>)
                    )}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs="auto" sx={{
                display: "flex"
              }}>
                <Box sx={{
                  display: "flex",
                  alignItems: "center"
                }}>
                  <Box sx={{
                    position: 'relative',
                  }}>
                    <Button
                      variant="contained"
                      onClick={refreshConfig}
                      disabled={isLoading || !configAddress}
                      sx={{
                        height: "40px",
                        borderRadius: "10px"
                      }}>Refresh keys</Button>
                  </Box>
                  <IconButton
                    variant="outlined"
                    onClick={copyAddress}
                    aria-label="copy address"
                    sx={{
                      ml: 1,
                      borderRadius: "10px"
                    }}>
                    <ContentCopyIcon/>
                  </IconButton>
                </Box>
              </Grid>
            </Grid>
          </Box>
          {configAddress && (<Card variant="outlined" >
            <CardContent>
              <Typography variant="body1" display="block">
                <b>Balance</b>: {Number(web3.utils.fromWei(balance)).toFixed(6)} ETH (
                <Link href="https://goerlifaucet.com/" rel="noopener noreferrer" target="_blank" underline="hover">
                  Goerli faucet
                </Link>)
                
                <IconButton
                  onClick={() => getBalance(configAddress)}
                  sx={{
                    ml: 1,
                    height: "30px",
                    width: "30px"
                  }}
                >
                  <CachedIcon
                    sx={{
                      height: "20px",
                      width: "20px"
                  }}/>
                </IconButton>
              </Typography>
              <Divider textAlign="left" sx={{
                fontSize: 12,
                m:1
              }}>Keys configuration</Divider>
              <Typography variant="body1" display="block">
                <b>Session ID</b>: {configs[configAddress].sessionId}
              </Typography>
              <Typography variant="body1" display="block">
                <b>Participants</b>: {configs[configAddress].participants.join(', ')}
              </Typography>
            </CardContent>
          </Card>)}
          {isLoading && <LinearProgress />}
          <Divider sx={{
            mt: 4
          }}></Divider>
          <Box sx={{
            mb:2,
          }}>
            <Grid container>
              <Grid item xs={24} md={8} sx={{
                display: "flex",
                height: "55px",
                mb: 2,
                mt: 2,
                mr: 2
              }}>
                <TextField
                  id="eth"
                  label="ETH"
                  value={eth}
                  onChange={handleChangeEth}
                  placeholder="0"
                  variant="standard"
                  inputProps={{ type: 'text', shrink: "true" }}
                  sx={{
                    maxWidth: '16ch',
                  }}
                />
                <TextField
                  id="address"
                  label="Address"
                  value={address}
                  onChange={handleChangeAddress}
                  variant="standard"
                  placeholder="0x"
                  inputProps={{
                    type: 'text',
                    shrink: "true",
                  }}
                  sx={{
                    ml: 2,
                    flexGrow: 2
                  }}
                />
              </Grid>
              <Grid item xs="auto" sx={{
                display: "flex"
              }}>
                <Box sx={{
                  display: "flex",
                  alignItems: "center"
                }}>
                  <Button
                    variant="contained"
                    onClick={sendEth}
                    disabled={isLoading || !configAddress}
                    sx={{
                      height: "40px",
                      borderRadius: "10px",
                    }}>Send ETH</Button>
                </Box>
              </Grid>
              {
                txLink &&
                (<Typography sx={{
                  pl:2,
                  pr:2,
                  fontSize: "12px"
                }}> Check your transaction: <Link href={txLink} rel="noopener noreferrer" target="_blank" underline="hover">
                  {txLink}
                </Link></Typography>)
              }
            </Grid>

          </Box>
        </Paper>
        <Paper sx={{
          borderRadius: "5px",
          flexGrow: 1,
          mt: 2,
          p:2,
          background: "#302f45",
          overflowY: "scroll"
        }}>
          <Logs/>
        </Paper>
      </Container>
      <Snackbar
        open={openSnackbar}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        sx={{
          display: "block"
        }}
      >
        <Alert onClose={() => handleCloseSnackbar()} severity={snackbarSeverity} sx={{ width: '100%' }}>
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default App;
