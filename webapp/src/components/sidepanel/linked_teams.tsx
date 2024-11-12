import React from 'react'

import {useDispatch} from 'react-redux'

import {openLinkTeamModal} from '../../reducers'

import useLinkedTeams from '../../hooks/use_linked_teams'

import TeamRow from './team_row'

const LinkedTeams = () => {
  const {linkedTeams, refetch} = useLinkedTeams()

  const dispatch = useDispatch()

  const handleLink = () => {
    dispatch(openLinkTeamModal())
  }

  return (
    <div>
      <h2>Linked Parabol Teams</h2>
      <button onClick={handleLink}>Link Team</button>
      {linkedTeams?.map((team) => (
        <TeamRow
          key={team.id}
          team={team}
        />
      ))}
    </div>
  )
}

export default LinkedTeams

