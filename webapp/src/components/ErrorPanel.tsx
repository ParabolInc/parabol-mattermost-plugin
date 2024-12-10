import React from 'react'

import {useDispatch, useSelector} from 'react-redux'

import styled from 'styled-components'

const Panel = styled.div`
  display: flex;
  flex-direction: column;
  align-items: stretch;
  padding: 16px 8px;
`

const SidePanel = () => (
  <Panel>
    <div>Failed to connect to Parabol</div>
  </Panel>
)

export default SidePanel
