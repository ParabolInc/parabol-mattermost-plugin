import React from 'react'
import {Store, AnyAction} from 'redux'

import {GlobalState} from 'mattermost-redux/types/store'

import {init, loadRemote} from '@module-federation/enhanced/runtime'

import manifest from '@/manifest'
import {PluginRegistry} from '@/types/mattermost-webapp'

import ErrorPanel from './components/ErrorPanel'
import PanelTitle from './components/PanelTitle'
import {getAssetsUrl, getPluginServerRoute} from './selectors'

const {id} = manifest

export default class Plugin {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
  public async initialize(registry: PluginRegistry, store: Store<GlobalState, AnyAction>) {
    const pluginServerRoute = getPluginServerRoute(store.getState())

    try {
      init({
        name: 'parabol-main',
        remotes: [{
          name: 'parabol',
          entry: `${pluginServerRoute}/components/mattermost-plugin-entry.js`,
        }],
      })

      const plugin = await loadRemote<any>('parabol/plugin')
      plugin?.init(registry, store)
      console.log(`Initialized plugin ${id}`)
    } catch (e) {
      const iconUrl = `${getAssetsUrl(store.getState())}/parabol.png`

      const {toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
        ErrorPanel,
        <PanelTitle iconUrl={iconUrl}/>,
      )

      registry.registerChannelHeaderButtonAction(
        <img src={iconUrl}/>,
        () => store.dispatch(toggleRHSPlugin),
        'Open Parabol Panel',
      )
      console.error(`Failed to load content from Parabol to initialize plugin ${id}`, e)
    }
  }
}

declare global {
  interface Window {
    registerPlugin(pluginId: string, plugin: Plugin): void
  }
}

window.registerPlugin(manifest.id, new Plugin())
