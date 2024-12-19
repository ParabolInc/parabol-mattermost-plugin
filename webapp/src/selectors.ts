import {getConfig} from 'mattermost-redux/selectors/entities/general'

import {GlobalState} from 'mattermost-redux/types/store'

import manifest from '@/manifest'

const {id} = manifest

export const getSiteURL = (state: GlobalState) => {
  const config = getConfig(state)
  return config?.SiteURL ?? ''
}

export const getPluginRoot = (state: GlobalState) => {
  const config = getConfig(state)
  const siteURL = config?.SiteURL ?? ''
  return `${siteURL}/plugins/${id}`
}

export const getPluginServerRoute = (state: GlobalState) => {
  let basePath = ''
  const config = getConfig(state)
  const siteURL = config?.SiteURL ?? ''
  if (siteURL) {
    basePath = new URL(siteURL).pathname

    if (basePath && basePath[basePath.length - 1] === '/') {
      basePath = basePath.substr(0, basePath.length - 1)
    }
  }

  return `${basePath}/plugins/${id}`
}

export const getAssetsUrl = (state: GlobalState) => {
  const siteURL = getPluginRoot(state)
  return `${siteURL}/public`
}

