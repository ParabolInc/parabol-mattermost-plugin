import React, {lazy, Suspense} from 'react'
import {useSelector} from 'react-redux'

import {pushPostAsReflection} from '../../selectors'

const PushReflectionModal = lazy(() => import(/* webpackChunkName: 'PushReflectionModal' */ './push_reflection_modal'))

const PushReflectionModalRoot = () => {
  const postId = useSelector(pushPostAsReflection)
  if (!postId) {
    return null
  }

  return (
    <Suspense fallback={null}>
      <PushReflectionModal/>
    </Suspense>
  )
}

export default PushReflectionModalRoot
