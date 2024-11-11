import classNames from 'classnames'
import React from 'react'

type Props = {
  text?: React.ReactNode
  style?: React.CSSProperties
}
const LoadingSpinner = ({text, style}: Props) => {
  return (
    <span
      id='loadingSpinner'
      className={classNames('LoadingSpinner', {'with-text': Boolean(text)})}
      style={style}
      data-testid='loadingSpinner'
    >
      <span
        className='fa fa-spinner fa-fw fa-pulse spinner'
      />
      {text}
    </span>
  )
}

export default LoadingSpinner