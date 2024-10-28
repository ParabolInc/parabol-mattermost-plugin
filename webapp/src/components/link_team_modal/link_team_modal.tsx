import React, {useEffect, useMemo} from 'react'
import {Modal} from 'react-bootstrap'
import {useDispatch, useSelector} from 'react-redux'

import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common'

import {isError, useGetTemplatesQuery, useLinkedTeamsQuery, useLinkTeamMutation} from '../../api'
import {closeLinkTeamModal} from '../../reducers'
import {getAssetsUrl, isLinkTeamModalVisible} from '../../selectors'

const LinkTeamModal = () => {
  const isVisible = useSelector(isLinkTeamModalVisible)
  const {data: teamData, refetch} = useGetTemplatesQuery()
  useEffect(() => {
    if (isVisible) {
      refetch()
    }
  }, [isVisible, refetch])

  const [linkTeam] = useLinkTeamMutation()
  const channelId = useSelector(getCurrentChannelId)
  const {data: linkedTeamIds} = useLinkedTeamsQuery({channelId})

  const {teams} = teamData ?? {}
  const [selectedTeam, setSelectedTeam] = React.useState<NonNullable<typeof teams>[number]>()

  useEffect(() => {
    if (!selectedTeam && teams && teams.length > 0) {
      setSelectedTeam(teams[0])
    }
  }, [teams, selectedTeam])
  const dispatch = useDispatch()

  const onChangeTeam = (teamId: string) => {
    setSelectedTeam(teams?.find((team) => team.id === teamId))
  }

  const handleClose = () => {
    dispatch(closeLinkTeamModal())
  }

  const handleLink = async () => {
    if (!selectedTeam) {
      return
    }
    const res = await linkTeam({channelId, teamId: selectedTeam.id})

    if (isError(res)) {
      console.error('Failed to link team', res.error)
      return
    }
    handleClose()
  }

  const assetsPath = useSelector(getAssetsUrl)

  if (!isVisible) {
    return null
  }

  return (
    <Modal
      dialogClassName='modal--scroll'
      show={true}
      onHide={handleClose}
      onExited={handleClose}
      bsSize='large'
      backdrop='static'
    >
      <Modal.Header closeButton={true}>
        <Modal.Title>
          <img
            width={36}
            height={36}
            src={`${assetsPath}/parabol.png`}
          />
          {'Link a Parabol Team to this Channel'}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {teamData && (<>
          <div className='form-group'>
            <label
              className='control-label'
              htmlFor='team'
            >Choose Parabol Team<span className='error-text'> *</span></label>
            <div className='input-wrapper'>
              <select
                className='form-control'
                id='team'
                value={selectedTeam?.id}
                onChange={(e) => onChangeTeam(e.target.value)}
              >
                {teams?.map((team) => (
                  <option
                    key={team.id}
                    value={team.id}
                  >{team.name}</option>
                ))}
              </select>
            </div>
          </div>
        </>)}
      </Modal.Body>
      <Modal.Footer>
        <button
          className='btn btn-tertiary cancel-button'
          onClick={handleClose}
        >Cancel</button>
        <button
          className='btn btn-primary save-button'
          onClick={handleLink}
        >Link Team</button>
      </Modal.Footer>
    </Modal>
  )
}

export default LinkTeamModal
