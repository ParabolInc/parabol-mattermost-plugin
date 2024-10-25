import React, {useCallback, useMemo} from 'react'

import {useStartCheckInMutation, useStartRetrospectiveMutation, useStartSprintPokerMutation, useStartTeamPromptMutation} from './api'

export const useStartMeeting = () => {
  const [startRetrospective, retroStatus] = useStartRetrospectiveMutation()
  const [startCheckIn, checkinStatus] = useStartCheckInMutation()
  const [startSprintPoker, pokerStatus] = useStartSprintPokerMutation()
  const [startTeamPrompt, teamPromptStatus] = useStartTeamPromptMutation()

  const isLoading = useMemo(() => retroStatus.isLoading || checkinStatus.isLoading || pokerStatus.isLoading || teamPromptStatus.isLoading, [retroStatus.isLoading, checkinStatus.isLoading, pokerStatus.isLoading, teamPromptStatus.isLoading])
  const isError = useMemo(() => retroStatus.isError || checkinStatus.isError || pokerStatus.isError || teamPromptStatus.isError, [retroStatus.isError, checkinStatus.isError, pokerStatus.isError, teamPromptStatus.isError])
  const isSuccess = useMemo(() => retroStatus.isSuccess || checkinStatus.isSuccess || pokerStatus.isSuccess || teamPromptStatus.isSuccess, [retroStatus.isSuccess, checkinStatus.isSuccess, pokerStatus.isSuccess, teamPromptStatus.isSuccess])

  const startMeeting = useCallback((teamId: string, meetingType: string, templateId: string) => {
    switch (meetingType) {
      case 'retrospective':
        return startRetrospective({teamId, templateId})
      case 'action':
        return startCheckIn({teamId})
      case 'poker':
        return startSprintPoker({teamId, templateId})
      case 'teamPrompt':
        return startTeamPrompt({teamId})
      default: {
        const error = new Error('Invalid meeting type')
        return {
          error,
          unwrap: (): Promise<void> => Promise.reject(error),
        }
      }
    }
  }, [])

  return [startMeeting, {isLoading, isError, isSuccess}] as const
}
