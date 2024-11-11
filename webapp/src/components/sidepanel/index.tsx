import React from 'react'

import {useDispatch, useSelector} from 'react-redux'
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'

import {useGetActiveMeetingsQuery, useGetTemplatesQuery, useLinkedTeamsQuery} from '../../api'
import {openLinkTeamModal, openStartActivityModal} from '../../reducers'

const SidePanelRoot = () => {
  const {data: meetings, isLoading} = useGetActiveMeetingsQuery()
  const channelId = useSelector(getCurrentChannelId)
  const {data: linkedTeams} = useLinkedTeamsQuery({channelId})
  const {data} = useGetTemplatesQuery()
  const {teams} = data ?? {}

  const dispatch = useDispatch()

  const [selectedTab, setSelectedTab] = React.useState('linked-teams')

  const handleLink = () => {
    dispatch(openLinkTeamModal())
  }

  const handleStartActivity = () => {
    dispatch(openStartActivityModal())
  }

  console.log('linkedTeams', linkedTeams)
  console.log('teams', teams)

  return (
    <div>
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

      <h2>Linked Parabol Teams</h2>
      <button onClick={handleLink}>Add Team</button>
      {teams?.map((team) => (linkedTeams?.includes(team.id) ? (
        <div key={team.id}>
          <h3>{team.name}</h3>
        </div>
      ) : <div key={team.id}>Unlinked {team.name}</div>),
      )}
      <h2>Active Meetings</h2>
      <button onClick={handleStartActivity}>Start Activity</button>
      {meetings?.map((meeting) => (
        <div key={meeting.id}>
          <h3>{meeting.name}</h3>
        </div>
      ))}
    </div>
  )
}

export default SidePanelRoot

