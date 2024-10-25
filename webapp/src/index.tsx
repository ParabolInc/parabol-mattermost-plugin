import React from 'react'
import {Store, AnyAction} from 'redux'

import {setupListeners} from '@reduxjs/toolkit/query'

import {GlobalState} from 'mattermost-redux/types/store'

import manifest from '@/manifest'

import {PluginRegistry} from '@/types/mattermost-webapp'

import StartActivityModal from './components/start_activity'
import rootReducer, {openPushPostAsReflection, openStartActivityModal} from './reducers'
import {getAssetsUrl} from './selectors'
import {api} from './api'
import SidePanelRoot from './components/sidepanel'
import PushReflectionModal from './components/push_reflection/push_reflection_modal'

const {id} = manifest

export default class Plugin {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
  public async initialize(registry: PluginRegistry, store: Store<GlobalState, AnyAction>) {
    //slightly hacky, might not be necessary
    store.dispatch = api.middleware(store as any)(store.dispatch)
    setupListeners(store.dispatch)
    registry.registerReducer(rootReducer)

    // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
    registry.registerRootComponent(StartActivityModal)
    registry.registerWebSocketEventHandler(`custom_${manifest.id}_open_start_activity_modal`, (message) => {
      store.dispatch(openStartActivityModal())
    })
    registry.registerRootComponent(PushReflectionModal)

    const {toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
      SidePanelRoot,
      <div>
        <img
          width={24}
          height={24}
          src={`${getAssetsUrl(store.getState())}/parabol.png`}
        />Parabol
      </div>,
    )
    registry.registerChannelHeaderButtonAction(
      <img src={`${getAssetsUrl(store.getState())}/parabol.png`}/>,

      // In the future we want to toggle the side panel
      //() => store.dispatch(toggleRHSPlugin),
      () => store.dispatch(openStartActivityModal()),
      'Start a Parabol Activity',
    )

    registry.registerPostDropdownMenuAction(
      <div><span className='MenuItem__icon'><img src={`${getAssetsUrl(store.getState())}/parabol.png`}/></span>Push reflection to Parabol</div>,
      (postId) => store.dispatch(openPushPostAsReflection(postId)),
    )

    console.log(`Initialized plugin ${id}`)
  }
}

declare global {
  interface Window {
    registerPlugin(pluginId: string, plugin: Plugin): void
  }
}

window.registerPlugin(manifest.id, new Plugin())
