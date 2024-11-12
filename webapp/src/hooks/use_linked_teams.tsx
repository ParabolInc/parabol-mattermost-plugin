import React, {useCallback, useEffect, useMemo} from 'react'
import {useSelector} from 'react-redux'
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'
import {useLinkedTeamsQuery, useTeamsQuery} from '../api'

export const useLinkedTeams = () => {
  const channelId = useSelector(getCurrentChannelId)
  const {data: teams, refetch: refetchTeams} = useTeamsQuery()
  const {data: linkedTeamIds, refetch: refetchLinkedTeams} = useLinkedTeamsQuery({channelId})

  const refetch = useCallback(() => {
    refetchTeams()
    refetchLinkedTeams()
  }, [refetchTeams, refetchLinkedTeams])

  useEffect(() => {
    refetch()
  }, [refetch])

  const [linkedTeams, unlinkedTeams] = useMemo(() => {
    if (!teams) {
      return [null, null]
    }
    return [
      teams.filter((team) => linkedTeamIds?.includes(team.id)),
      teams.filter((team) => !linkedTeamIds?.includes(team.id))
    ]
  }, [teams, linkedTeamIds])

  return {
    linkedTeams,
    unlinkedTeams,
    refetch,
  }
}

export default useLinkedTeams
