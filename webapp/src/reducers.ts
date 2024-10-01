import {combineReducers} from 'redux';

import {ActionTypes} from 'src/actions';

const isStartActivityModalVisible = (state = false, action) => {
    switch (action.type) {
    case ActionTypes.OPEN_START_ACTIVITY_MODAL:
        console.log('GEORG Open Start Activity Modal');
        return true;
    case ActionTypes.CLOSE_START_ACTIVITY_MODAL:
        console.log('GEORG Close Start Activity Modal');
        return false;
    default:
        return state;
    }
};

const meetingTemplates = (state = {}, action) => {
    switch (action.type) {
    case ActionTypes.MEETING_TEMPLATES:
        console.log('GEORG reduce Meeting Templates', action.data);
        return action.data
    default:
        return state;
    }
}

export default combineReducers({
    isStartActivityModalVisible,
    meetingTemplates,
});
