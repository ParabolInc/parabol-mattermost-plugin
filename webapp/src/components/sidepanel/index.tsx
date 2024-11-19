import React from 'react'

import {useDispatch} from 'react-redux'

import styled from 'styled-components'

import LinkedTeams from './linked_teams'
import ActiveMeetings from './active_meetings'

const Panel = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
  padding: 16px 8px;
`

const SidePanelRoot = () => {
  const dispatch = useDispatch()

  const [selectedTab, setSelectedTab] = React.useState('linked-teams')

  return (
    <Panel>
      <LinkedTeams/>
      <ActiveMeetings/>
    </Panel>
  )
}

export default SidePanelRoot

