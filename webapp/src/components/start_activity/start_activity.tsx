import React, {useEffect, useMemo} from 'react';
import {Modal} from 'react-bootstrap';
import Spinner from 'react-bootstrap/Spinner';
import {useDispatch, useSelector} from 'react-redux';
import {isError, useGetTemplatesQuery} from '../../api';
import {useStartMeeting} from '../../hooks';
import {closeStartActivityModal} from '../../reducers';
import {getAssetsUrl, isStartActivityModalVisible} from '../../selectors';
import MeetingSettings from './meeting_settings';

const StartActivity = () => {
  const isVisible = useSelector(isStartActivityModalVisible);

  const {data, isLoading} = useGetTemplatesQuery();
  const {availableTemplates, teams} = data ?? {};
  const [selectedTeam, setSelectedTeam] = React.useState<NonNullable<typeof teams>[number]>();
  const [selectedTemplate, setSelectedTemplate] = React.useState<NonNullable<typeof availableTemplates>[number]>();

  const filteredTemplates = useMemo(() => availableTemplates?.filter((template) =>
          template.scope === 'PUBLIC'
          || template.scope === 'TEAM' && template.teamId === selectedTeam?.id
          || template.scope === 'ORGANIZATION' && template.orgId === selectedTeam?.orgId
        ), [availableTemplates, selectedTeam]);

  useEffect(() => {
    if (!selectedTeam && teams && teams.length > 0) {
      setSelectedTeam(teams[0]);
    }
  }, [teams, selectedTeam]);
  useEffect(() => {
    if (!selectedTemplate && filteredTemplates && filteredTemplates.length > 0) {
      setSelectedTemplate(filteredTemplates[0]);
    }
  }, [filteredTemplates, selectedTemplate]);

  const onChangeTeam = (teamId: string) => {
    setSelectedTeam(teams?.find((team) => team.id === teamId));
  }

  const onChangeTemplate = (templateId: string) => {
    setSelectedTemplate(availableTemplates?.find((template) => template.id === templateId));
  }

  const dispatch = useDispatch();

  const handleClose = () => {
    dispatch(closeStartActivityModal());
  }

  const [startMeeting, {isLoading: isStartActivityLoading, isError: isStartActivityError}] = useStartMeeting();

  const handleStart = async () => {
    if (!selectedTeam || !selectedTemplate) {
      return;
    }
    const res = await startMeeting(selectedTeam.id, selectedTemplate.type, selectedTemplate.id);

    if (isError(res)) {
      console.error('Failed to start activity', res.error);
      return;
    }
    handleClose();
  }

  const assetsPath = useSelector(getAssetsUrl);

  if (!isVisible) {
    return null;
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
              <img width={36} height={36} src={`${assetsPath}/parabol.png`} />
                {'Start a Parabol Activity'}
            </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div>
            <p>To see the full details for any activity, visit <a href='https://mattermost.com'>Parabol's Activity Library</a></p>
          </div>
          {isLoading &&
             <Spinner animation="border" role="status">
               <span className="visually-hidden">Loading...</span>
             </Spinner>
          }
          {data && (<>
            <div className='form-group'>
              <label className='control-label' htmlFor='team'>Choose Parabol Team<span className='error-text'> *</span></label>
              <div className='input-wrapper'>
                <select
                  className='form-control'
                  id='team'
                  value={selectedTeam?.id}
                  onChange={(e) => onChangeTeam(e.target.value)}
                >
                  {teams?.map((team) => (
                    <option key={team.id} value={team.id}>{team.name}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className='form-group'>
              <label htmlFor='activity'>Choose Activity<span className='error-text'> *</span></label>
              <select
                className='form-control'
                id='activity'
                value={selectedTemplate?.id}
                onChange={(e) => onChangeTemplate(e.target.value)}
              >
                {filteredTemplates?.map((template) => (
                  <option key={template.id} value={template.id}>{template.name}</option>
                ))}
              </select>
            </div>
            {selectedTeam && selectedTemplate && ['retrospective', 'action', 'poker'].includes(selectedTemplate.type) && (
                <MeetingSettings teamId={selectedTeam.id} meetingType={selectedTemplate.type} />
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

export default StartActivity;
