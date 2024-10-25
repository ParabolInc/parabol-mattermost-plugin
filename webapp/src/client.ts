import {Client4} from 'mattermost-redux/client'

export type RetroSettings = {
  id: string
  checkinEnabled: boolean
  teamHealthEnabled: boolean
  disableAnonymity: boolean
}

class Client {
  url: string

  constructor(url: string) {
    this.url = url
  }

  post = async <T>(path: string, body: T): Promise<Response> => {
    const response = await fetch(`${this.url}/${path}`, Client4.getOptions({
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    }))
    if (!response.ok) {
      throw new Error(`Failed to fetch ${path}`)
    }
    return response.json()
  }

  getTemplates = () => {
    return this.post('templates', {})
  }

  setMeetingSettings = (settings: RetroSettings) => {
    return this.post('meeting-settings', settings)
  }
}

let client: Client
export const initClient = (url: string) => {
  client = new Client(url)
}

function getClient() {
  return client!
}

export default getClient

