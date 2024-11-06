import React from 'react'
import ReactSelect from 'react-select'

interface IdName {
  id: string
  name: string
}
export type SelectProps<T extends IdName> = {
  label?: string
  required?: boolean
  labelProp?: keyof T
  options: T[]
  selected: T | null
  onChange: (value: T | null) => void
}

const LinkTeamModal = <T extends IdName>(props: SelectProps<T>) => {
  const {label, required, options, selected, onChange} = props
  return (
    <div className='form-group'>
      {label &&
        <label
          className='control-label'
          htmlFor='team'
        >{label}{required && <span className='error-text'> *</span>}</label>
      }
      <div className='Input_Wrapper'>
        <ReactSelect
          id='team'
          value={selected && {value: selected.id, label: selected.name}}
          options={options.map(({id, name}) => ({value: id, label: name}))}
          onChange={(value) => onChange(options.find(({id}) => id === value?.value) ?? null)}
          styles={{ menuPortal: (base) => ({ ...base, zIndex: 9999 }) }}
          menuPortalTarget={document.body}
          isSearchable={true}
          menuPosition='fixed'
        />
      </div>
    </div>
  )
}

export default LinkTeamModal
