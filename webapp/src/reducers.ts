import {AnyAction, createSlice, PayloadAction} from '@reduxjs/toolkit'

import {api} from './api'

const localSlice = createSlice({
  name: 'local',
  initialState: {
    isStartActivityModalVisible: false,
    pushPostAsReflection: null as string | null,
  },
  reducers: {
    openStartActivityModal: (state) => {
      state.isStartActivityModalVisible = true
    },
    closeStartActivityModal: (state) => {
      state.isStartActivityModalVisible = false
    },
    openPushPostAsReflection: (state, action: PayloadAction<string>) => {
      state.pushPostAsReflection = action.payload
    },
    closePushPostAsReflection: (state) => {
      state.pushPostAsReflection = null
    },
  },
})

export const {
  openStartActivityModal,
  closeStartActivityModal,
  openPushPostAsReflection,
  closePushPostAsReflection,
} = localSlice.actions

export type PluginState = ReturnType<typeof localSlice.reducer> & ReturnType<typeof api.reducer>

const rootReducer = (state: PluginState, action: AnyAction) => {
  const apiState = api.reducer(state, action)
  const localState = localSlice.reducer(state, action)
  return {
    ...localState,
    ...apiState,
  }
}
export default rootReducer

