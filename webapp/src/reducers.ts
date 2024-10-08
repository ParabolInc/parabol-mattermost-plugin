import {createReducer, createSlice} from '@reduxjs/toolkit';

import {ActionTypes} from 'src/actions';
import {api} from './api';

/*
const isStartActivityModalVisible = (state = false, action) => {
    switch (action?.type) {
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

const reducer = createReducer({isStartActivityModalVisible: false}, (builder) => {
    builder
        .addCase(ActionTypes.OPEN_START_ACTIVITY_MODAL, state => {
            console.log('GEORG Open Start Activity Modal');
            return {isStartActivityModalVisible: true}})
        .addCase(ActionTypes.CLOSE_START_ACTIVITY_MODAL, state => {return {isStartActivityModalVisible: false}})
});
*/

const localSlice = createSlice({
    name: 'local',
    initialState: {isStartActivityModalVisible: false},
    reducers: {
        openStartActivityModal: (state) => {
            console.log('GEORG Open Start Activity Modal');
            state.isStartActivityModalVisible = true;
        },
        closeStartActivityModal: (state) => {
            console.log('GEORG Close Start Activity Modal');
            state.isStartActivityModalVisible = false;
        },
    },
});

export const {openStartActivityModal, closeStartActivityModal} = localSlice.actions;

const rootReducer = (state, action) => {
    //console.log('GEORG rootReducer', state, action);
    const apiState = api.reducer(state, action);
    //console.log('GEORG intermediateState', intermediateState);
    const localState = localSlice.reducer(state , action);
    //console.log('GEORG finalState', finalState);
    Object.keys(localState).forEach((key) => {
        if (apiState[key] !== undefined) {
            console.log('GEORG duplicate key', key, apiState[key]);
        }
    });
    return {
        ...localState,
        ...apiState,
    };
}
export default rootReducer;

