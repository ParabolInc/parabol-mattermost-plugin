import React, {useMemo} from 'react'
import {isError, useConfigQuery, useLinkedTeamsQuery, useTeamsQuery, useUnlinkTeamMutation} from '../../api'
import {useSelector} from 'react-redux'
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'
import MoreMenu from '../menu'
import styled from 'styled-components'
import plural from '../../utils'

const Card = styled.div`
  display: flex;
  flex-direction: column;
  padding: 8px;
  margin: 8px 0;
  border: 1px solid #ccc;
  border-radius: 5px;
`

const Col = styled.div`
  display: flex;
  flex-direction: column;
`

const Row = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: start;
  padding: 4px 0;
`

const Name = styled.div`
  font-size: 1.5rem;
  font-weight: bold;
`
const MemberCount = styled.div`
  font-size: 1.5rem;
`

type Props = {
  meeting: {
    id: string
    name: string
    teamId: string
  }
}

const MeetingRow = ({meeting}: Props) => {
  const {id, name, teamId} = meeting
  const {data: teams} = useTeamsQuery()
  const team = useMemo(() => teams?.find((t) => t.id === teamId), [teams, teamId])
  const {data: config} = useConfigQuery()

  return (
    <Card>
      <Row>
        <Col>
          <Name>{name}</Name>
          <MemberCount>{team?.name}</MemberCount>
        </Col>
      </Row>
      <Row>
        <a href={`${config?.parabolURL}/meet/${id}`} target='_blank'>{"Join Meeting"}</a>
      </Row>
    </Card>
  )
}

export default MeetingRow
