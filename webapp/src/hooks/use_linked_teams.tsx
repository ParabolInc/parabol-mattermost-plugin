import React, {useCallback, useMemo} from 'react'
import {useSelector} from 'react-redux'
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'

import {useLinkedTeamsQuery, useTeamsQuery} from '../api'

export const useLinkedTeams = () => {
  const channelId = useSelector(getCurrentChannelId)
  const {data: teams, isLoading: isLoadingTeams, error: teamsError, refetch: refetchTeams} = useTeamsQuery()
  const {data: linkedTeamIds, isLoading: isLoadingLinkedTeamIds, error: linkedTeamIdsError, refetch: refetchLinkedTeams} = useLinkedTeamsQuery({channelId})

  const refetch = useCallback(() => {
    refetchTeams()
    refetchLinkedTeams()
  }, [refetchTeams, refetchLinkedTeams])

  const [linkedTeams, unlinkedTeams] = useMemo(() => {
    if (!teams) {
      return [null, null]
    }
    return [
      teams.filter((team) => linkedTeamIds?.includes(team.id)),
      teams.filter((team) => !linkedTeamIds?.includes(team.id)),
    ]
  }, [teams, linkedTeamIds])

  const isLoading = isLoadingTeams || isLoadingLinkedTeamIds
  const error = teamsError || linkedTeamIdsError

  return {
    linkedTeams,
    unlinkedTeams,
    isLoading,
    error,
    refetch,
  }
}

export default useLinkedTeams
