import React, {useEffect, useState} from 'react'

import {useSelector} from 'react-redux'
import styled from 'styled-components'
import {Client4} from 'mattermost-redux/client'

import {getPluginServerRoute} from '../selectors'

const Panel = styled.div`
  display: flex;
  align-items: center;
`

const TitleLink = styled.a`
  font-size: 1.5rem;
  font-weight: bold;
  margin-left: 8px;
  color: #000;
  text-decoration: none;
`

type Props = {
  iconUrl: string
}

const PanelTitle = ({iconUrl}: Props) => {
  const pluginServerRoute = useSelector(getPluginServerRoute)
  const [parabolURL, setParabolURL] = useState<string>()

  useEffect(() => {
    if (!pluginServerRoute) {
      return
    }
    const fetchConfig = async () => {
      try {
        const response = await fetch(`${pluginServerRoute}/config`, Client4.getOptions({method: 'GET'}))
        const data = await response.json()
        setParabolURL(data.parabolURL)
      } catch (error) {
        console.log('Failed to fetch config', error)
      }
    }
    fetchConfig()
  }, [pluginServerRoute])

  return (
    <Panel>
      <img
        width={24}
        height={24}
        src={iconUrl}
      />
      <TitleLink
        href={parabolURL}
        target='_blank'
      >Parabol</TitleLink>
    </Panel>
  )
}

export default PanelTitle

