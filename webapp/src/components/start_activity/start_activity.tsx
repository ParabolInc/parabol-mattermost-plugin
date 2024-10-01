import React, {useMemo} from 'react';
import {Modal, Form} from 'react-bootstrap';
import {useSelector} from 'react-redux';
import {getAssetsUrl} from '../../selectors';
import RetroSettings from './retro_settings';

type Props = {
  visible: boolean;
  meetingTemplates: any;
  close: () => void;
}

const StartActivity = ({visible, meetingTemplates, close}: Props) => {
  const {availableTemplates, teams} = meetingTemplates;

  const [selectedTeam, setSelectedTeam] = React.useState(teams?.[0]);
  const filteredTemplates = useMemo(() => availableTemplates?.filter((template) =>
          template.scope === 'PUBLIC'
          || template.scope === 'TEAM' && template.teamId === selectedTeam.id
          || template.scope === 'ORGANIZATION' && template.orgId === selectedTeam.orgId
        ), [availableTemplates, selectedTeam]);
  const [selectedTemplate, setSelectedTemplate] = React.useState(filteredTemplates?.[0]);

  const onChangeTeam = (teamId: string) => {
    setSelectedTeam(teams.find((team) => team.id === teamId));
  }

  const onChangeTemplate = (templateId: string) => {
    setSelectedTemplate(availableTemplates.find((template) => template.id === templateId));
  }

  const assetsPath = useSelector(getAssetsUrl);

  if (!visible) {
    return null;
  }

  return (
    <Modal
        dialogClassName='modal--scroll'
        show={true}
        onHide={close}
        onExited={close}
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
          <div>To see the full details for any activity, visit <a href='https://mattermost.com'>Parabol's Activity Library</a></div>
          <div className='form-group'>
            <label className='control-label' htmlFor='team'>Choose Parabol Team<span className='error-text'> *</span></label>
            <div className='input-wrapper'>
              <select
                className='form-control'
                id='team'
                value={selectedTeam.id}
                onChange={(e) => onChangeTeam(e.target.value)}
              >
                {teams.map((team) => (
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
              value={selectedTemplate.id}
              onChange={(e) => onChangeTemplate(e.target.value)}
            >
              {filteredTemplates.map((template) => (
                <option key={template.id} value={template.id}>{template.name}</option>
              ))}
            </select>
          </div>

          {selectedTemplate?.type === 'retrospective' && (
              <RetroSettings settings={selectedTeam.retroSettings} />
          )}
        </Modal.Body>
        <Modal.Footer>
          <button
            className='btn btn-tertiary cancel-button'
            onClick={close}
          >Cancel</button>
          <button
            className='btn btn-primary save-button'
            onClick={close}
          >Start Activity</button>
        </Modal.Footer>
      </Modal>
  )


}

export default StartActivity;
