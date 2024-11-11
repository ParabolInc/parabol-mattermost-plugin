# Parabol Mattermost Plugin

## WIP ##

This plugin is very much work in progress and not yet ready for testing.

## Requirements

- Mattermost version 10.1 or later
- `SiteURL` If the `SiteURL` is not set correctly, some functions like notifications will not work.
- SSO identity provider for both Mattermost and Parabol

## Development

Start up a Mattermost development server
```bash
make start-server
```
go to http://localhost:8065 and do the initial setup. Create a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html)
and set the environment variable `MM_ADMIN_TOKEN` to the token value.


Run 
```
make watch
```
to start the plugin in development mode. This will watch for webapp changes and automatically rebuild the plugin.

Configure the plugin in mattermost, go to System Console -> Plugins -> Parabol and enter
- Parabol URL: http://host.docker.internal:3001/mattermost
- Parabol API Token: get this from MATTERMOST_SECRET environment from your Parabol instance

### Releasing new versions

The version of a plugin is determined at compile time, automatically populating a `version` field in the [plugin manifest](plugin.json):
* If the current commit matches a tag, the version will match after stripping any leading `v`, e.g. `1.3.1`.
* Otherwise, the version will combine the nearest tag with `git rev-parse --short HEAD`, e.g. `1.3.1+d06e53e1`.
* If there is no version tag, an empty version will be combined with the short hash, e.g. `0.0.0+76081421`.

To disable this behaviour, manually populate and maintain the `version` field.

## How to Release

To trigger a release, follow these steps:

1. **For Patch Release:** Run the following command:
    ```
    make patch
    ```
   This will release a patch change.

2. **For Minor Release:** Run the following command:
    ```
    make minor
    ```
   This will release a minor change.

3. **For Major Release:** Run the following command:
    ```
    make major
    ```
   This will release a major change.

4. **For Patch Release Candidate (RC):** Run the following command:
    ```
    make patch-rc
    ```
   This will release a patch release candidate.

5. **For Minor Release Candidate (RC):** Run the following command:
    ```
    make minor-rc
    ```
   This will release a minor release candidate.

6. **For Major Release Candidate (RC):** Run the following command:
    ```
    make major-rc
    ```
   This will release a major release candidate.

