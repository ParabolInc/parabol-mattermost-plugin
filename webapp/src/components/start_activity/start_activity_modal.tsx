import React, {useEffect, useMemo} from 'react'
import {Modal} from 'react-bootstrap'
import {useDispatch, useSelector} from 'react-redux'

import {isError, useConfigQuery, useTemplatesQuery, useTeamsQuery} from '../../api'
import useStartMeeting from '../../hooks/use_start_meeting'
import {closeStartActivityModal} from '../../reducers'
import {getAssetsUrl, isStartActivityModalVisible} from '../../selectors'

import Select from '../select'

import LoadingSpinner from '../loading_spinner'

import MeetingSettings from './meeting_settings'

const StartActivityModal = () => {
  const {data: availableTemplates, isLoading: isTemplatesLoading, refetch: refetchTemplates, error: templatesError} = useTemplatesQuery()
  const {data: teams, isLoading: isTeamsLoading, refetch: refetchTeams, error: teamsError} = useTeamsQuery()
  const {data: config} = useConfigQuery()
  const isVisible = useSelector(isStartActivityModalVisible)
  useEffect(() => {
    if (isVisible) {
      refetchTemplates()
      refetchTeams()
    }
  }, [isVisible, refetchTemplates, refetchTeams])

  const [selectedTeam, setSelectedTeam] = React.useState<NonNullable<typeof teams>[number] | null>(null)
  const [selectedTemplate, setSelectedTemplate] = React.useState<NonNullable<typeof availableTemplates>[number] | null>(null)

  const filteredTemplates = useMemo(() => availableTemplates?.filter((template) =>
    template.scope === 'PUBLIC' ||
          (template.scope === 'TEAM' && template.teamId === selectedTeam?.id) ||
          (template.scope === 'ORGANIZATION' && template.orgId === selectedTeam?.orgId),
  ), [availableTemplates, selectedTeam])

  useEffect(() => {
    if (!selectedTeam && teams && teams.length > 0) {
      setSelectedTeam(teams[0])
    }
  }, [teams, selectedTeam])
  useEffect(() => {
    if (!selectedTemplate && filteredTemplates && filteredTemplates.length > 0) {
      setSelectedTemplate(filteredTemplates[0])
    }
  }, [filteredTemplates, selectedTemplate])

  const dispatch = useDispatch()

  const handleClose = () => {
    dispatch(closeStartActivityModal())
  }

  const [startMeeting, {isLoading: isStartActivityLoading, isError: isStartActivityError}] = useStartMeeting()

  const handleStart = async () => {
    if (!selectedTeam || !selectedTemplate) {
      return
    }
    if (isStartActivityLoading) {
      return
    }

    const res = await startMeeting(selectedTeam.id, selectedTemplate.type, selectedTemplate.id)

    if (isError(res)) {
      console.error('Failed to start activity', res.error)
      return
    }
    handleClose()
  }

  const assetsPath = useSelector(getAssetsUrl)

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
          {' Start a Parabol Activity'}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div>
          <p>To see the full details for any activity, visit
            <a
              href={`${config?.parabolURL}/activity-library/`}
              target='_blank'
              rel='noreferrer'
            >
              {"Parabol's Activity Library"}
            </a>
          </p>
        </div>
        {isTeamsLoading && <LoadingSpinner text='Loading...'/>}
        {teamsError && <div className='error-text'>Loading teams failed, try refreshing the page</div>}
        {teams &&
          <Select
            label='Choose Parabol Team'
            required={true}
            options={teams ?? []}
            value={selectedTeam}
            onChange={setSelectedTeam}
          />
        }
        {isTemplatesLoading && <LoadingSpinner text='Loading...'/> }
        {templatesError && <div className='error-text'>Loading templates failed, try refreshing the page</div>}
        {selectedTeam && availableTemplates && (<>
          <Select
            label='Choose Activity'
            required={true}
            options={filteredTemplates ?? []}
            value={selectedTemplate}
            onChange={setSelectedTemplate}
          />
          {selectedTeam && selectedTemplate && ['retrospective', 'action', 'poker'].includes(selectedTemplate.type) && (
            <MeetingSettings
              teamId={selectedTeam.id}
              meetingType={selectedTemplate.type}
            />
          )}
        </>)}
      </Modal.Body>
      <Modal.Footer>
        <button
          className='btn btn-tertiary cancel-button'
          onClick={handleClose}
        >Cancel</button>
        <button
          className='btn btn-primary save-button'
          onClick={handleStart}
        >Start Activity</button>
      </Modal.Footer>
    </Modal>
  )
}

export default StartActivityModal
