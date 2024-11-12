import React from 'react'

import {useMeetingSettingsQuery, useSetMeetingSettingsMutation, MeetingSettings as Settings} from '../../api'

interface Props {
  teamId: string;
  meetingType: string;//'retrospective' | 'action' | 'poker';
}

const MeetingSettings = ({teamId, meetingType}: Props) => {
  const {data, isLoading, isError, refetch} = useMeetingSettingsQuery({teamId, meetingType})
  const [setMeetingSettings] = useSetMeetingSettingsMutation()

  const onChange = async (key: keyof Settings, value: boolean) => {
    if (!data) {
      return
    }
    await setMeetingSettings({
      ...data,
      [key]: value,
    })
    refetch()
  }

  if (!data) {
    return null
  }

  const {checkinEnabled, teamHealthEnabled, disableAnonymity} = data

  return (
    <div className='form-group'>
      <div className='checkbox'>
        <label>
          <input
            type='checkbox'
            onChange={(e) => onChange('checkinEnabled', e.target.checked)}
            checked={checkinEnabled}
          />
          <span>Include Icebreaker</span>
        </label>
      </div>
      <div className='checkbox'>
        <label>
          <input
            type='checkbox'
            onChange={(e) => onChange('teamHealthEnabled', e.target.checked)}
            checked={teamHealthEnabled}
          />
          <span>Include Team Health</span>
        </label>
      </div>
      {disableAnonymity !== null && (
        <div className='checkbox'>
          <label>
            <input
              type='checkbox'
              onChange={(e) => onChange('disableAnonymity', !e.target.checked)}
              checked={!disableAnonymity}
            />
            <span>Anonymous Reflections</span>
          </label>
        </div>
      )}
      {isError && <div className='alert alert-danger'>Error updating settings</div>}
      {isLoading && <div className='alert alert-info'>Updating...</div>}
    </div>
  )
}

export default MeetingSettings
