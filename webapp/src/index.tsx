import {Store, Action} from 'redux';

import {GlobalState} from '@mattermost/types/lib/store';

import manifest from '@/manifest';

import {PluginRegistry} from '@/types/mattermost-webapp';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerChannelHeaderButtonAction(
            // icon - JSX element to use as the button's icon
            'Should be an icon',//<Icon />,
            // action - a function called when the button is clicked, passed the channel and channel member as arguments
            // null,
            () => {
                alert("Hello World!");
            },
            // dropdown_text - string or JSX element shown for the dropdown button description
            "Hello World",
        );
        registry.registerRightHandSidebarComponent("Hello World", 'Hello World');
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
