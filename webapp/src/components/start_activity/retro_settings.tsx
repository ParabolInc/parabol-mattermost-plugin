import React from 'react';
import getClient, {RetroSettings as Settings} from '../../client';

const RetroSettings = ({settings}: {settings: Settings}) => {
    const onChange = async (key: keyof Settings, value: boolean) => {
        const client = getClient();
        await client.setMeetingSettings({
            ...settings,
            [key]: value,
        });
    }

    return (
        <div className='form-group'>
            <div className='checkbox'>
                <label>
                    <input type="checkbox" onChange={(e) => onChange('checkinEnabled', e.target.checked)}/>
                    <span>Include Icebreaker</span>
                </label>
            </div>
            <div className='checkbox'>
                <label>
                    <input type="checkbox" onChange={(e) => onChange('teamHealthEnabled', e.target.checked)}/>
                    <span>Include Team Health</span>
                </label>
            </div>
            <div className='checkbox'>
                <label>
                    <input type="checkbox" onChange={(e) => onChange('disableAnonymity', !e.target.checked)}/>
                    <span>Anonymous Reflections</span>
                </label>
            </div>
        </div>
    )
}

export default RetroSettings
