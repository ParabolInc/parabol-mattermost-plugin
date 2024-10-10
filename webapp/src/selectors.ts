import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getCurrentTeam} from 'mattermost-redux/selectors/entities/teams';
import manifest from '@/manifest';

const {id} = manifest;


export const getSiteURL = (state) => {
  const config = getConfig(state);
  return config?.SiteURL ?? '';
};

export const getCurrentTeamURL = (state) => {
  const siteURL = getSiteURL(state);
  const team = getCurrentTeam(state);
  return `${siteURL}/${team?.name ?? ''}`;
}

export const getPostURL = (state, postId) => {
  const teamURL = getCurrentTeamURL(state);
  return `${teamURL}/pl/${postId}`;
}

export const getPluginRoot = (state) => {
    const config = getConfig(state);
    const siteURL = config?.SiteURL ?? '';
    return `${siteURL}/plugins/${id}`;
};


export const getPluginServerRoute = (state) => {
    let basePath = '';
    const config = getConfig(state);
    const siteURL = config?.SiteURL ?? '';
    if (siteURL) {
        basePath = new URL(siteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return `${basePath}/plugins/${id}`;
};

export const getAssetsUrl = (state) => {
    const siteURL = getPluginRoot(state);
    return `${siteURL}/public`;
};

export const getPluginState = (state) => state[`plugins-${id}`] ?? {};

export const meetingTemplates = (state) => getPluginState(state).meetingTemplates;

export const isStartActivityModalVisible = (state) => getPluginState(state).isStartActivityModalVisible;

export const pushPostAsReflection = (state) => getPluginState(state).pushPostAsReflection;

//export const get
