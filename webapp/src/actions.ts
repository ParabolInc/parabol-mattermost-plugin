import manifest from '@/manifest';
import getClient from './client';

const {id} = manifest;

export const ActionTypes = {
    OPEN_START_ACTIVITY_MODAL: `custom_${id}_open_start_activity_modal`,
    CLOSE_START_ACTIVITY_MODAL: `custom_${id}_close_start_activity_modal`,
    MEETING_TEMPLATES: `custom_${id}_meeting_templates`,
}

export const openStartActivityModal = (data: {title: string, channelId: string}) => {
    return async (dispatch) => {
        dispatch(getMeetingTemplates());

        return {
            type: ActionTypes.OPEN_START_ACTIVITY_MODAL,
            data
        };
    }
}

export const getMeetingTemplates = () => {
    return async (dispatch) => {
        dispatch({
            type: ActionTypes.OPEN_START_ACTIVITY_MODAL,
        })
        const data = await getClient().getTemplates()
        dispatch({
            type: ActionTypes.MEETING_TEMPLATES,
            data
        });
    };
}

export const closeStartActivityModal = () => {
    return {
        type: ActionTypes.CLOSE_START_ACTIVITY_MODAL,
    };
}

