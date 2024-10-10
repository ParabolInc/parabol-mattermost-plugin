import React from 'react';
import {useSelector} from 'react-redux';
import {getAssetsUrl} from '../../selectors';

const PushReflectionButton = (props) => {
  console.log('PushReflectionButton', props);
  const assetsPath = useSelector(getAssetsUrl);

  const handleClick = () => {
    console.log('Push Reflection Button clicked');
  }

  return (
    <button
      className='post-menu__item'
      onClick={handleClick}
    >
      <img width={16} height={16} src={`${assetsPath}/parabol.png`} />
    </button>
  );
}

export default PushReflectionButton;
