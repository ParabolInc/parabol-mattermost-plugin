import React, {useEffect, useMemo} from 'react'
import {Modal} from 'react-bootstrap'
import {useDispatch, useSelector} from 'react-redux'

import {getPost} from 'mattermost-redux/selectors/entities/posts'

import {GlobalState} from 'mattermost-redux/types/store'

import {useCreateReflectionMutation, useActiveMeetingsQuery} from '../../api'
import {closePushPostAsReflection} from '../../reducers'
import {getAssetsUrl, getPostURL, pushPostAsReflection} from '../../selectors'
import Select from '../select'
import LoadingSpinner from '../loading_spinner'

const PostUtils = (window as any).PostUtils

const PushReflectionModal = () => {
  const postId = useSelector(pushPostAsReflection)
  const post = useSelector((state: GlobalState) => getPost(state, postId!))
  const postUrl = useSelector((state: GlobalState) => getPostURL(state, postId!))

  const {data, isLoading, refetch} = useActiveMeetingsQuery()
  useEffect(() => {
    if (postId) {
      refetch()
    }
  }, [postId, refetch])

  const retroMeetings = useMemo(() => data?.filter(({meetingType}) => meetingType === 'retrospective'), [data])
  const [selectedMeeting, setSelectedMeeting] = React.useState<NonNullable<typeof data>[number] | null>(null)
  const [selectedPrompt, setSelectedPrompt] = React.useState<NonNullable<NonNullable<typeof data>[number]['reflectPrompts']>[number] | null>(null)

  const [comment, setComment] = React.useState('')
  const formattedPost = useMemo(() => {
    if (!post) {
      return null
    }
    const quotedMessage = post.message.split('\n').map((line) => `> ${line}`).join('\n')
    setComment(quotedMessage)
    return `[See comment in Mattermost](${postUrl})`
  }, [post])

  const [createReflection] = useCreateReflectionMutation()

  useEffect(() => {
    setComment('')
  }, [postId])

  useEffect(() => {
    if (!selectedMeeting && retroMeetings && retroMeetings.length > 0) {
      setSelectedMeeting(retroMeetings[0])
    }
  }, [data, selectedMeeting])

  useEffect(() => {
    setSelectedPrompt(selectedMeeting?.reflectPrompts?.[0] ?? null)
  }, [selectedMeeting])

  const dispatch = useDispatch()

  const handleClose = () => {
    dispatch(closePushPostAsReflection())
  }

  const handlePush = async () => {
    if (!selectedMeeting || !selectedPrompt || (!comment && !post.message)) {
      console.log('missing data', selectedPrompt, selectedMeeting, comment, post.message)
      return
    }

    const content = `${comment}\n\n${formattedPost}`

    await createReflection({
      meetingId: selectedMeeting.id,
      promptId: selectedPrompt.id,
      content,
      sortOrder: 0,
    })

    handleClose()
  }

  const assetsPath = useSelector(getAssetsUrl)

  if (!postId) {
    return null
  }

  return (
    <Modal
      dialogClassName='modal--scroll'
      show={true}
      onHide={handleClose}
      onExited={handleClose}
      bsSize='large'
      backdrop='static'
    >
      <Modal.Header closeButton={true}>
        <Modal.Title>
          <img
            width={36}
            height={36}
            src={`${assetsPath}/parabol.png`}
          />
          {' Add Comment to Parabol Activity'}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div>
          <p>Choose an open Retro activity and the Prompt where you want to send the Mattermost comment. A reference link back to Mattermost will be inlcuded in the reflection.</p>
        </div>
        {isLoading &&
        <LoadingSpinner text='Loading...'/>
        }
        {post && (
          <div className='form-group'>
            <label
              className='control-label'
              htmlFor='comment'
            >Add a Comment<span className='error-text'> *</span></label>
            <div
              className='form-control'
              style={{
                resize: 'none',
                height: 'auto',
              }}
            >
              <textarea
                style={{
                  border: 'none',
                  width: '100%',
                }}
                id='comment'
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                placeholder='Add your comment for the retro...'
              />
              <blockquote>
                {PostUtils.messageHtmlToComponent(PostUtils.formatText(post.message))}
              </blockquote>
              <a>See comment in Mattermost</a>
            </div>
          </div>
        )}
        {data && (<>
          <Select
            label='Choose Activity'
            required={true}
            value={selectedMeeting}
            options={retroMeetings ?? []}
            onChange={setSelectedMeeting}
          />
          <Select
            label='Choose Prompt'
            required={true}
            value={selectedPrompt && {id: selectedPrompt.id, name: selectedPrompt.question}}
            options={selectedMeeting?.reflectPrompts?.map(({id, question}) => ({id, name: question})) ?? []}
            onChange={(selected) => selected && setSelectedPrompt(selectedMeeting?.reflectPrompts?.find((prompt) => prompt.id === selected.id) ?? null)}
          />
        </>)}
      </Modal.Body>
      <Modal.Footer>
        <button
          className='btn btn-tertiary cancel-button'
          onClick={handleClose}
        >Cancel</button>
        <button
          className='btn btn-primary save-button'
          onClick={handlePush}
        >Add Comment</button>
      </Modal.Footer>
    </Modal>
  )
}

export default PushReflectionModal
