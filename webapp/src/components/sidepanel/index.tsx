import React from 'react'

import {useDispatch} from 'react-redux'

import LinkedTeams from './linked_teams'
import ActiveMeetings from './active_meetings'
import styled from 'styled-components'

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
      {/*
      <div className='form-group'>
        <label
          className='control-label'
          htmlFor='team'
        >Choose Parabol Team<span className='error-text'> *</span></label>
        <div className='input-wrapper'>
          <select
            className='form-control'
            id='team'
            value={selectedTab}
            onChange={(e) => setSelectedTab(e.target.value)}
          >
            <option
              key='linked-teams'
              value='linked-teams'
            >Linked Parabol Teams</option>
          </select>
        </div>
      </div>
      */}
      <LinkedTeams />
      <ActiveMeetings />
    </Panel>
  )
}

export default SidePanelRoot

