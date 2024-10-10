import React from 'react';
import {useGetActiveMeetingsQuery} from '../../api';

const SidePanelRoot = () => {
  const {data, isLoading} = useGetActiveMeetingsQuery()
  return (
    <div>
      <h1>SidePanelRoot</h1>
      {data?.map((meeting) => (
        <div key={meeting.id}>
          <h2>{meeting.name}</h2>
        </div>
      ))}
    </div>
  );
}

export default SidePanelRoot;

