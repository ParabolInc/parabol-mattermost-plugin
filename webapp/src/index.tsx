import React from 'react'
import {Store, AnyAction} from 'redux'

import {setupListeners} from '@reduxjs/toolkit/query'

import {GlobalState} from 'mattermost-redux/types/store'

import manifest from '@/manifest'
import {PluginRegistry} from '@/types/mattermost-webapp'

import StartActivityModal from './components/start_activity'
import LinkTeamModal from './components/link_team_modal'
import rootReducer, {openPushPostAsReflection, openStartActivityModal} from './reducers'
import {getAssetsUrl, getPluginServerRoute} from './selectors'

//import {api} from './api'
import SidePanelRoot from './components/sidepanel'
import PushReflectionModal from './components/push_reflection/push_reflection_modal'
import PanelTitle from './components/sidepanel/panel_title'

const {id} = manifest
import { init } from '@module-federation/enhanced/runtime'

export default class Plugin {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
  public async initialize(registry: PluginRegistry, store: Store<GlobalState, AnyAction>) {
    //slightly hacky, might not be necessary
    //store.dispatch = api.middleware(store as any)(store.dispatch)
    setupListeners(store.dispatch)
    registry.registerReducer(rootReducer)

    // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
    registry.registerRootComponent(StartActivityModal)
    registry.registerRootComponent(LinkTeamModal)
    registry.registerWebSocketEventHandler(`custom_${manifest.id}_open_start_activity_modal`, () => {
      store.dispatch(openStartActivityModal())
    })
    registry.registerRootComponent(PushReflectionModal)

    const {toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
      SidePanelRoot,
      <PanelTitle/>,
    )
    registry.registerChannelHeaderButtonAction(
      <img src={`${getAssetsUrl(store.getState())}/parabol.png`}/>,
      () => store.dispatch(toggleRHSPlugin),
      'Open Parabol Panel',
    )

    registry.registerPostDropdownMenuAction(
      <div><span className='MenuItem__icon'><img src={`${getAssetsUrl(store.getState())}/parabol.png`}/></span>Push reflection to Parabol</div>,
      (postId) => store.dispatch(openPushPostAsReflection(postId)),
    )

    const pluginServerRoute = getPluginServerRoute(store.getState())

    init({
      name: 'parabol-main',
      remotes: [{
        name: 'parabol',
        entry: `${pluginServerRoute}/components/remoteEntry.js`,
      }]
    })

    console.log(`Initialized plugin ${id}`)
  }
}

declare global {
  interface Window {
    registerPlugin(pluginId: string, plugin: Plugin): void
  }
}

window.registerPlugin(manifest.id, new Plugin())
