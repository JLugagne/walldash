import { createHashRouter, Navigate, useParams } from 'react-router-dom'
import App from './App'
import { ThreeDView } from './components/ThreeDView'
import { AdminView } from './components/AdminView'
import { OverviewsRoute } from './components/OverviewsRoute'
import { HouseOverviewRoute } from './components/HouseOverviewRoute'
import { SetupShell } from './components/setup/SetupShell'
import { DevicesList } from './components/setup/DevicesList'
import { DashboardsManager } from './components/setup/DashboardsManager'
import { SettingsPanel } from './components/setup/SettingsPanel'

function LegacyAdminRedirect() {
  const { levelId } = useParams<{ levelId?: string }>()
  return <Navigate to={levelId ? `/setup/plans/${levelId}` : '/setup/plans'} replace />
}

export const router = createHashRouter([
  {
    element: <App />,
    children: [
      { index: true, element: <HouseOverviewRoute /> },
      { path: 'floor/:levelId', element: <ThreeDView /> },
      { path: 'house', element: <HouseOverviewRoute /> },
      { path: 'admin', element: <LegacyAdminRedirect /> },
      { path: 'admin/:levelId', element: <LegacyAdminRedirect /> },
      { path: 'overviews', element: <OverviewsRoute /> },
      {
        path: 'setup',
        element: <SetupShell />,
        children: [
          { index: true, element: <Navigate to="/setup/plans" replace /> },
          { path: 'plans', element: <AdminView /> },
          { path: 'plans/:levelId', element: <AdminView /> },
          { path: 'devices', element: <DevicesList /> },
          { path: 'dashboards', element: <DashboardsManager /> },
          { path: 'settings', element: <SettingsPanel /> },
        ],
      },
      { path: '*', element: <ThreeDView /> },
    ],
  },
])
