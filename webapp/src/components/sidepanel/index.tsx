import React from 'react'

import {useSelector} from 'react-redux'
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'

import {useGetActiveMeetingsQuery} from '../../api'

const SidePanelRoot = () => {
  const {data, isLoading} = useGetActiveMeetingsQuery()
  const channel = useSelector(getCurrentChannelId)
  return (
    <div>
      <h1>SidePanelRoot</h1>
      <h2>Channel: {channel}</h2>
      {data?.map((meeting) => (
        <div key={meeting.id}>
          <h2>{meeting.name}</h2>
        </div>
      ))}
    </div>
  )
}

export default SidePanelRoot

