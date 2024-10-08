import { BaseQueryFn, buildCreateApi, coreModule, createApi, FetchArgs, fetchBaseQuery, FetchBaseQueryError, reactHooksModule } from '@reduxjs/toolkit/query/react'
import {Client4} from 'mattermost-redux/client'
import {useSelector} from 'react-redux'
import {getPluginServerRoute, getPluginState} from './selectors'

export interface Post {
  id: string
  title: string
  author: string
  content: string
  status: (typeof postStatuses)[number]
  created_at: string
  updated_at: string
}

export interface Pagination {
  page: number
  per_page: number
  total: number
  total_pages: number
}

type Template = {
    id: string
    name: string
    type: string
    illustrationUrl: string
    orgId: string
    teamId: string
    scope: string
}

type MeetingSettings = {
    id: string
    phaseTypes: string[]
}

type RetroSettings = MeetingSettings & {
    disableAnonymity: boolean
}

type Team = {
    id: string
    name: string
    orgId: string
    retroSettings: RetroSettings
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
  console.log('GEORG baseQuery', args)
  const baseUrl = getPluginServerRoute(api.getState())
  console.log('GEORG baseUrl', baseUrl)
  const adjustedArgs =
    typeof args === 'string' ? {
      url: joinUrl(baseUrl, args),
    }: {
      ...args,
      url: joinUrl(baseUrl, args.url),
    }
  return rawBaseQuery(Client4.getOptions(adjustedArgs as any) as any, api, extraOptions)
}

const customCreateApi = buildCreateApi(
  coreModule(),
  reactHooksModule({
    useSelector: (state) => {
      return useSelector(getPluginState(state))
    },
  }),
)


//export const initApi = (baseUrl: string) => createApi({
export const api = createApi({
  reducerPath: 'plugins-co.parabol.action',
  baseQuery,
  endpoints: (builder) => ({
    getTemplates: builder.query<MeetingTemplatesResponse, void>({
        query: () => ({
            url: '/templates2',
            method: 'POST',
        }),
        transformResponse: (response) => {
          console.log('GEORG transformResponse', response)
          return response as any
        },
        transformErrorResponse: (response) => {
          console.log('GEORG transformErrorResponse', response)
          return response
        },
    }),
  }),
})

export const { useGetTemplatesQuery } = api
