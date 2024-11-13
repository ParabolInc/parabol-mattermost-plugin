import {BaseQueryFn, createApi, FetchArgs, fetchBaseQuery, FetchBaseQueryError} from '@reduxjs/toolkit/query/react'
import {Client4} from 'mattermost-redux/client'

import manifest from '@/manifest'

import {getPluginServerRoute} from './selectors'

const {id} = manifest

type MeetingTemplate = {
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

type TeamMember = {
  id: string
  email: string
}

type Team = {
  id: string
  name: string
  orgId: string
  teamMembers: TeamMember[]
}

type ReflectPrompt = {
  id: string
  question: string
  description: string
}

export type Meeting = {
  id: string
  teamId: string
  name: string
  meetingType: string
  templateId: string
  reflectPrompts?: ReflectPrompt[]
  isComplete: boolean
}

export type CreateReflectionInput = {
  content: string
  meetingId: string
  promptId: string
  sortOrder: number
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
  const baseUrl = getPluginServerRoute(api.getState() as any)
  const adjustedArgs =
    typeof args === 'string' ? {
      url: joinUrl(baseUrl, args),
    } : {
      ...args,
      url: joinUrl(baseUrl, args.url),
    }
  return rawBaseQuery(Client4.getOptions(adjustedArgs as any) as any, api, extraOptions)
}

export const api = createApi({
  reducerPath: `plugins-${id}`,
  baseQuery,
  tagTypes: ['Teams', 'MeetingTemplates', 'MeetingSettings', 'Meetings'],
  endpoints: (builder) => ({
    templates: builder.query<MeetingTemplate[], void>({
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
        body: variables,
      }),
      invalidatesTags: ['MeetingSettings'],
    }),
    meetingSettings: builder.query<MeetingSettings, {teamId: string, meetingType: string}>({
      query: (variables) => ({
        url: '/query/meetingSettings',
        method: 'POST',
        body: variables,
      }),
      providesTags: ['MeetingSettings'],
    }),
    startRetrospective: builder.mutation<void, {teamId: string, templateId: string}>({
      query: (variables) => ({
        url: '/query/startRetrospective',
        method: 'POST',
        body: variables,
      }),
      invalidatesTags: ['Meetings'],
    }),
    startCheckIn: builder.mutation<void, {teamId: string}>({
      query: (variables) => ({
        url: '/query/startCheckIn',
        method: 'POST',
        body: variables,
      }),
      invalidatesTags: ['Meetings'],
    }),
    startSprintPoker: builder.mutation<void, {teamId: string, templateId: string}>({
      query: (variables) => ({
        url: '/query/startSprintPoker',
        method: 'POST',
        body: variables,
      }),
      invalidatesTags: ['Meetings'],
    }),
    startTeamPrompt: builder.mutation<void, {teamId: string}>({
      query: (variables) => ({
        url: '/query/startTeamPrompt',
        method: 'POST',
        body: variables,
      }),
      invalidatesTags: ['Meetings'],
    }),
    activeMeetings: builder.query<Meeting[], void>({
      query: () => ({
        url: '/query/activeMeetings',
        method: 'POST',
      }),
      providesTags: ['Meetings'],
    }),
    createReflection: builder.mutation<void, CreateReflectionInput>({
      query: (variables) => ({
        url: '/query/createReflection',
        method: 'POST',
        body: variables,
      }),
    }),
    teams: builder.query<Team[], void>({
      query: () => ({
        url: '/query/teams',
        method: 'POST',
      }),
      providesTags: ['Teams'],
    }),
    linkedTeams: builder.query<string[], {channelId: string}>({
      query: ({channelId}) => ({
        url: `/linkedTeams/${channelId}`,
        method: 'GET',
      }),
      providesTags: ['Teams'],
    }),
    linkTeam: builder.mutation<void, {channelId: string, teamId: string}>({
      query: ({channelId, teamId}) => ({
        url: `/linkTeam/${channelId}/${teamId}`,
        method: 'POST',
      }),
      invalidatesTags: ['Teams'],
    }),
    unlinkTeam: builder.mutation<void, {channelId: string, teamId: string}>({
      query: ({channelId, teamId}) => ({
        url: `/unlinkTeam/${channelId}/${teamId}`,
        method: 'POST',
      }),
      invalidatesTags: ['Teams'],
    }),
    config: builder.query<{parabolURL: string}, void>({
      query: () => ({
        url: '/config',
        method: 'GET',
      }),
    }),
  }),
})

export const isError = (result: any): result is {error: Error} => {
  return 'error' in result && result.error instanceof Object
}

export const {
  useTemplatesQuery,
  useSetMeetingSettingsMutation,
  useMeetingSettingsQuery,
  useStartRetrospectiveMutation,
  useStartTeamPromptMutation,
  useStartCheckInMutation,
  useStartSprintPokerMutation,
  useActiveMeetingsQuery,
  useCreateReflectionMutation,
  useLinkedTeamsQuery,
  useTeamsQuery,
  useLinkTeamMutation,
  useUnlinkTeamMutation,
  useConfigQuery,
} = api
