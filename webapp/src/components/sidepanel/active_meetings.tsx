import React from 'react'

import {useDispatch} from 'react-redux'

import {useActiveMeetingsQuery} from '../../api'
import {openStartActivityModal} from '../../reducers'
import MeetingRow from './meeting_row'
import LoadingSpinner from '../loading_spinner'

const ActiveMeetings = () => {
  const {data: meetings, isLoading, error, refetch} = useActiveMeetingsQuery()

  const dispatch = useDispatch()

  const handleStartActivity = () => {
    dispatch(openStartActivityModal())
  }

  return (
    <div>
      <h2>Active Meetings</h2>
      <button onClick={handleStartActivity}>Start Activity</button>
      {isLoading && <LoadingSpinner text='Loading...'/>}
      {error && <div className='error-text'>Loading meetings failed, try refreshing the page</div>}
      {meetings?.map((meeting) =>
        <MeetingRow key={meeting.id} meeting={meeting} />
      )}
    </div>
  )
}

export default ActiveMeetings

