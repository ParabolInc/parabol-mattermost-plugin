import React from 'react';

interface Props {
    settings: any;
}

const RetroSettings = (props: Props) => {
    console.log('GEORG MeetingSettings!', props);
    return (
        <div className='form-group'>
            <div className='checkbox'>
                <label>
                    <input type="checkbox"/>
                    <span>Include Icebreaker</span>
                </label>
            </div>
            <div className='checkbox'>
                <label>
                    <input type="checkbox"/>
                    <span>Include Team Health</span>
                </label>
            </div>
            <div className='checkbox'>
                <label>
                    <input type="checkbox"/>
                    <span>Anonymous Reflections</span>
                </label>
            </div>
        </div>
    )
}

export default RetroSettings
