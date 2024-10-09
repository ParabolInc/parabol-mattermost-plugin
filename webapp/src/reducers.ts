import {createSlice} from '@reduxjs/toolkit';

import {api} from './api';

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
    const apiState = api.reducer(state, action);
    const localState = localSlice.reducer(state , action);
    return {
        ...localState,
        ...apiState,
    };
}
export default rootReducer;

