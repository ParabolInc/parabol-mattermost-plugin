import React from 'react';
import {Store, AnyAction} from 'redux';

import {GlobalState} from '@mattermost/types/lib/store';

import manifest from '@/manifest';

import {PluginRegistry} from '@/types/mattermost-webapp';
import StartActivity from './components/start_activity/start_activity';
import rootReducer, {openStartActivityModal}  from './reducers';
import {getAssetsUrl} from './selectors';
import {api} from './api';
import {setupListeners} from '@reduxjs/toolkit/query';

const {id} = manifest;

export default class Plugin {

    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, AnyAction>) {
        //slightly hacky, might not be necessary
        store.dispatch = api.middleware(store as any)(store.dispatch);
        setupListeners(store.dispatch);
        registry.registerReducer(rootReducer);

        registry.registerRootComponent(StartActivity);
        registry.registerWebSocketEventHandler(`custom_${manifest.id}_error`, (message) => {
            console.error(message);
        });

        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerChannelHeaderButtonAction(
            <img src={`${getAssetsUrl(store.getState())}/parabol.png`} />,
            // action - a function called when the button is clicked, passed the channel and channel member as arguments
            () => store.dispatch(openStartActivityModal()),
            // dropdown_text - string or JSX element shown for the dropdown button description
            "Start a Parabol Activity",
        );
        registry.registerWebSocketEventHandler(`custom_${manifest.id}_open_start_activity_modal`, (message) => {
            store.dispatch(openStartActivityModal());
        });

        registry.registerRightHandSidebarComponent("Hello World", 'Hello World');
        console.log(`Initialized plugin ${id}`);
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
