import getSSE from '../api/api.sse'
import { useState } from 'react'
import { Typography } from '@mui/material'
import { Mutex } from "async-mutex"

const participantStyle = {
  color: '#ff8359',
};

const roundStyle = {
  color: '#4786ff',
};

const sessionStyle = {
  color: '#00c89c',
};

const decorateMessage = (data) => {
  return (
    <p style={{
      margin: "2px"
    }} key={Math.random()}>
      <b>[{new Date(data.timestamp * 1000).toLocaleString() }]: </b>
      <span>{data.message} </span>
      {data.participant ? (
        <>
        <span>[</span>
        <span style={participantStyle}>ID: </span><span>{data.participant}</span>
        {data.ip ? (<><span style={participantStyle}> IP: </span><span>{data.ip}</span></>) : ''}
        {data.protocol ? (<><span style={roundStyle}> P: </span><span>{data.protocol}</span></>) : ''}
        {data.round ? (<><span style={roundStyle}> R: </span><span>{data.round}</span></>) : ''}
        {data.sessionID ? (<><span style={sessionStyle}> SID: </span><span>{data.sessionID}</span></>) : ''}
        <span>]</span>
        </>)
        : ''}
    </p>
  )
}

export default function Logs() {
  const initialMessage = decorateMessage({
    timestamp: Date.now() / 1000,
    message: "console initialized, scheme - {message} [ID: {participant ID} IP: {participant IP} P: {protocol} R: {round} SID: {session ID}]"
  })
  const [logs, setLogs] = useState([initialMessage])
  const mutex = new Mutex()
  const sseClient = getSSE()
  
  sseClient.onmessage = async (event) => {
    await mutex.acquire().then((release) => {
      const data = JSON.parse(event.data)
      setLogs((prevState => {
        return prevState.concat(decorateMessage(data));
      }));
      release();
    });
  }
  
  return (
    <Typography variant="body" whiteSpace="pre-line" color="white" fontSize="12px">
      {logs}
    </Typography>
  )
}
