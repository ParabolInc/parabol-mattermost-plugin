import React, { lazy, Suspense } from 'react'
//import useDynamicScript from '../../hooks/use_dynamic_script'
import {importRemote} from './import_remote'
import { loadRemote } from '@module-federation/enhanced/runtime'

type Props = {
  system: {
    module: string
    scope: string
    url: string
  }
}

const System = ({system}: Props) => {
  /*
    const { ready, failed } = useDynamicScript({
        url: system && system.url,
    })

    if (!system) {
        return <h2>Not system specified</h2>
    }

    if (!ready) {
        return <h2>Loading dynamic script: {system.url}</h2>
    }

    if (failed) {
        return <h2>Failed to load dynamic script: {system.url}</h2>
    }
   */

    const Component = lazy(() => loadRemote<any>('parabol/button'))
                           //importRemote(system.url, system.scope, system.module))

    return (
        <Suspense fallback="Loading System">
            <Component />
        </Suspense>
    )
}

export default System
