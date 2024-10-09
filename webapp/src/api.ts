import {BaseQueryFn, createApi, FetchArgs, fetchBaseQuery, FetchBaseQueryError} from '@reduxjs/toolkit/query/react'
import {Client4} from 'mattermost-redux/client'
import {getPluginServerRoute} from './selectors'
import manifest from '@/manifest';

const {id} = manifest;

type Template = {
    id: string
    name: string
    type: string
    illustrationUrl: string
    orgId: string
    teamId: string
    scope: string
}

export type MeetingSettings = {
    id: string
    checkinEnabled: boolean
    teamHealthEnabled: boolean
    disableAnonymity?: boolean
}

type Team = {
    id: string
    name: string
    orgId: string
    retroSettings: MeetingSettings
    pokerSettings: MeetingSettings
    actionSettings: MeetingSettings
}

type MeetingTemplatesResponse = {
    availableTemplates: Template[]
    teams: Team[]
}

const joinUrl = (baseUrl: string, url: string) => {
  if (url.startsWith('/')) {
    return `${baseUrl}${url}`
  }
  return url
}

const rawBaseQuery = fetchBaseQuery({

})

const baseQuery: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  const baseUrl = getPluginServerRoute(api.getState())
  const adjustedArgs =
    typeof args === 'string' ? {
      url: joinUrl(baseUrl, args),
    }: {
      ...args,
      url: joinUrl(baseUrl, args.url),
    }
  return rawBaseQuery(Client4.getOptions(adjustedArgs as any) as any, api, extraOptions)
}

export const api = createApi({
  reducerPath: `plugins-${id}`,
  baseQuery,
  tagTypes: ['MeetingTemplates', 'MeetingSettings'],
  endpoints: (builder) => ({
    getTemplates: builder.query<MeetingTemplatesResponse, void>({
      query: () => ({
        url: '/query/meetingTemplates',
        method: 'POST',
      }),
    }),
    // teamId and meetingType are required for convenient cache updates
    setMeetingSettings: builder.mutation<MeetingSettings, MeetingSettings>({
      query: (variables) => ({
        url: '/query/setMeetingSettings',
        method: 'POST',
        body: variables
      }),
      invalidatesTags: () => {
        console.log('invalidating tags')
        return ['MeetingSettings']
      },
      onQueryStarted: (variables, {dispatch, queryFulfilled}) => {
        console.log('onQueryStarted')
      },
      onCacheEntryAdded: (cache, action) => {
        console.log('onCacheEntryAdded')
      },

        /*(result) => {
        return result ? [{type: 'MeetingSettings', id: result.id}] : ['MeetingSettings']
      }
      */
    }),
    getMeetingSettings: builder.query<MeetingSettings, {teamId: string, meetingType: string}>({
      query: (variables) => ({
        url: '/query/getMeetingSettings',
        method: 'POST',
        body: variables
      }),
      providesTags: () => {
        console.log('providing tags')
        return ['MeetingSettings']
      }
        /*(result) => {
        return result ? ['MeetingSettings', {type: 'MeetingSettings', id: result.id}] : ['MeetingSettings']
      }
      */
    }),
    startRetrospective: builder.mutation<void, {teamId: string, templateId: string}>({
      query: (variables) => ({
        url: '/query/startRetrospective',
        method: 'POST',
        body: variables
      }),
    }),
    startCheckIn: builder.mutation<void, {teamId: string}>({
      query: (variables) => ({
        url: '/query/startCheckIn',
        method: 'POST',
        body: variables
      }),
    }),
    startSprintPoker: builder.mutation<void, {teamId: string, templateId: string}>({
      query: (variables) => ({
        url: '/query/startSprintPoker',
        method: 'POST',
        body: variables
      }),
    }),
    startTeamPrompt: builder.mutation<void, {teamId: string}>({
      query: (variables) => ({
        url: '/query/startTeamPrompt',
        method: 'POST',
        body: variables
      }),
    }),
  }),
})


export const isError = (result: any): result is {error: Error} => {
  return 'error' in result && result.error instanceof Object
}

export const { useGetTemplatesQuery, useSetMeetingSettingsMutation, useGetMeetingSettingsQuery, useStartRetrospectiveMutation, useStartTeamPromptMutation, useStartCheckInMutation, useStartSprintPokerMutation } = api
