import { OverviewsView } from './OverviewsView'
import { ViewModeMenu } from './ViewModeMenu'

export function OverviewsRoute() {
  return (
    <OverviewsView
      initialIsAdmin={false}
      viewModeMenu={<ViewModeMenu />}
    />
  )
}