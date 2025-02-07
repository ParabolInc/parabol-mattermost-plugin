import React from 'react'

import styled from 'styled-components'

const Panel = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
`

const SidePanel = () => (
  <Panel>
    <div>Failed to connect to Parabol.</div>
    <br/>
    <a
      href='#'
      onClick={() => window.location.reload()}
    >
      Reload page?
    </a>
  </Panel>
)

export default SidePanel
