import { useParams } from 'react-router-dom'
import { DashboardsView } from './DashboardsView'

export function DashboardEditorRoute() {
  const { dashboardId } = useParams<{ dashboardId: string }>()
  return <DashboardsView mode="edit" dashboardId={dashboardId} />
}
