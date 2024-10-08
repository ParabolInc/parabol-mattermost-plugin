import manifest from '@/manifest';

const {id} = manifest;

export const ActionTypes = {
    OPEN_START_ACTIVITY_MODAL: `custom_${id}_open_start_activity_modal`,
    CLOSE_START_ACTIVITY_MODAL: `custom_${id}_close_start_activity_modal`,
}

export const openStartActivityModal = () => {
    return {
        type: ActionTypes.OPEN_START_ACTIVITY_MODAL,
    };
}

export const closeStartActivityModal = () => {
    return {
        type: ActionTypes.CLOSE_START_ACTIVITY_MODAL,
    };
}

