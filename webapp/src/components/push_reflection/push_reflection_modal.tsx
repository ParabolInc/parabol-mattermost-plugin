import React, {useEffect, useMemo} from 'react';
import {Modal} from 'react-bootstrap';
import Spinner from 'react-bootstrap/Spinner';
import {useDispatch, useSelector} from 'react-redux';
import {useCreateReflectionMutation, useGetActiveMeetingsQuery} from '../../api';
import {closePushPostAsReflection} from '../../reducers';
import {getAssetsUrl, getPostURL, pushPostAsReflection} from '../../selectors';
import {getPost} from 'mattermost-redux/selectors/entities/posts';

const PostUtils = window.PostUtils;

const PushReflectionModal = () => {
  const postId = useSelector(pushPostAsReflection);
  const post = useSelector(state => getPost(state, postId));
  const postUrl = useSelector(state => getPostURL(state, postId));
  console.log('postUrl', postUrl);

  const {data, isLoading} = useGetActiveMeetingsQuery();
  const retroMeetings = useMemo(() => data?.filter(({meetingType}) => meetingType === 'retrospective'), [data]);
  const [selectedMeeting, setSelectedMeeting] = React.useState<NonNullable<typeof data>[number]>();
  const [selectedPrompt, setSelectedPrompt] = React.useState<NonNullable<NonNullable<typeof data>[number]['reflectPrompts']>[number]>();

  const [comment, setComment] = React.useState('');
  const formattedPost = useMemo(() => {
    if (!post) {
      return undefined;
    }
    const quotedMessage = post.message.split('\n').map((line) => `> ${line}`).join('\n');
    return `${quotedMessage}\n\n[See comment in Mattermost](${postUrl})`;
    //http://localhost:8065/parabol/pl/7nkyzed7qbgyideowpy6uows3o
  }, [post]);

  const [createReflection] = useCreateReflectionMutation();

  useEffect(() => {
    setComment('');
  }, [postId]);

  useEffect(() => {
    if (!selectedMeeting && retroMeetings && retroMeetings.length > 0) {
      setSelectedMeeting(retroMeetings[0]);
    }
  }, [data, selectedMeeting]);

  useEffect(() => {
    setSelectedPrompt(selectedMeeting?.reflectPrompts?.[0]);
  }, [selectedMeeting]);

  const onChangeMeeting = (meetingId: string) => {
    setSelectedMeeting(retroMeetings?.find((meeting) => meeting.id === meetingId));
  }

  const onChangePrompt = (promptId: string) => {
    setSelectedPrompt(selectedMeeting?.reflectPrompts?.find((prompt) => prompt.id === promptId));
  }

  const dispatch = useDispatch();

  const handleClose = () => {
    dispatch(closePushPostAsReflection());
  }

  const handlePush = async () => {
    if (!selectedMeeting || !selectedPrompt || (!comment && !post.message)) {
      console.log('missing data', selectedPrompt, selectedMeeting, comment, post.message);
      return;
    }

    const content = `${comment}\n\n${formattedPost}`;

    await createReflection({
      meetingId: selectedMeeting.id,
      promptId: selectedPrompt.id,
      content,
      sortOrder: 0,
    });

    handleClose();
  }

  const assetsPath = useSelector(getAssetsUrl);

  if (!postId) {
    return null;
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
              <img width={36} height={36} src={`${assetsPath}/parabol.png`} />
                {'Start a Parabol Activity'}
            </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div>
            <p>Choose an open Retro activity and the Prompt where you want to send the Mattermost comment. A reference link back to Mattermost will be inlcuded in the reflection.</p>
          </div>
          {isLoading &&
             <Spinner animation="border" role="status">
               <span className="visually-hidden">Loading...</span>
             </Spinner>
          }
          {post && (
            <div className='form-group'>
              <label className='control-label' htmlFor='comment'>Add a Comment<span className='error-text'> *</span></label>
              <div
                className='form-control'
                style={{
                  resize: 'none',
                  height: 'auto',
                }}>
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
            <div className='form-group'>
              <label className='control-label' htmlFor='meeting'>Choose Retro<span className='error-text'> *</span></label>
              <div className='input-wrapper'>
                <select
                  className='form-control'
                  id='meeting'
                  value={selectedMeeting?.id}
                  onChange={(e) => onChangeMeeting(e.target.value)}
                >
                  {retroMeetings?.map((retro) => (
                    <option key={retro.id} value={retro.id}>{retro.name}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className='form-group'>
              <label htmlFor='prompt'>Choose Prompt<span className='error-text'> *</span></label>
              <select
                className='form-control'
                id='prompt'
                value={selectedPrompt?.id}
                onChange={(e) => onChangePrompt(e.target.value)}
              >
                {selectedMeeting?.reflectPrompts?.map((prompt) => (
                  <option key={prompt.id} value={prompt.id}>{prompt.question}</option>
                ))}
              </select>
            </div>
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

export default PushReflectionModal;
