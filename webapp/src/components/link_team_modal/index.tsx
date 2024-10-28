import React, {lazy, Suspense} from 'react'
import {useSelector} from 'react-redux'

import {isLinkTeamModalVisible} from '../../selectors'

const LinkTeamModal = lazy(() => import(/* webpackChunkName: 'LinkTeamModal' */ './link_team_modal'))

const LinkTeamModalRoot = () => {
  const isVisible = useSelector(isLinkTeamModalVisible)
  if (!isVisible) {
    return null
  }

  return (
    <Suspense fallback={null}>
      <LinkTeamModal/>
    </Suspense>
  )
}

export default LinkTeamModalRoot
