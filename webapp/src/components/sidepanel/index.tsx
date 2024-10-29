import React from 'react'

import {useDispatch, useSelector} from 'react-redux'
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'

import {useGetActiveMeetingsQuery, useLinkedTeamsQuery} from '../../api'
import {openLinkTeamModal, openStartActivityModal} from '../../reducers'

const SidePanelRoot = () => {
  const {data: meetings, isLoading} = useGetActiveMeetingsQuery()
  const channelId = useSelector(getCurrentChannelId)
  const {data: teams} = useLinkedTeamsQuery({channelId})
  const dispatch = useDispatch()

  const [selectedTab, setSelectedTab] = React.useState('linked-teams')

  const handleLink = () => {
    dispatch(openLinkTeamModal())
  }

  const handleStartActivity = () => {
    dispatch(openStartActivityModal())
  }

  return (
    <div>
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
        Foo
      </div>

      <h2>Linked Parabol Teams</h2>
      <button onClick={handleLink}>Add Team</button>
      <button onClick={handleStartActivity}>Start Activity</button>
      {teams}
      <h2>Channel: {channelId}</h2>
      {meetings?.map((meeting) => (
        <div key={meeting.id}>
          <h2>{meeting.name}</h2>
        </div>
      ))}
    </div>
  )
}

export default SidePanelRoot

