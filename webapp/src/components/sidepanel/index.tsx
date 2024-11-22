import React from 'react'

import {useDispatch, useSelector} from 'react-redux'

import styled from 'styled-components'

import LinkedTeams from './linked_teams'
import ActiveMeetings from './active_meetings'
import System from './system'
import {getPluginServerRoute} from '../../selectors'

const Panel = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
  padding: 16px 8px;
`

const SidePanelRoot = () => {
  const dispatch = useDispatch()
  const pluginServerRoute = useSelector(getPluginServerRoute)

  const [selectedTab, setSelectedTab] = React.useState('linked-teams')

  return (
    <Panel>
      <System system={{
        module: 'button',
        scope: 'parabol',
        url: `${pluginServerRoute}/components/remoteEntry.js`
      }}/>
      <LinkedTeams/>
      <ActiveMeetings/>
    </Panel>
  )
}

export default SidePanelRoot

