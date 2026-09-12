import { createHashRouter, Navigate } from 'react-router-dom'
import App from './App'
import { ThreeDView } from './components/ThreeDView'
import { AdminView } from './components/AdminView'
import { DashboardsRoute } from './components/DashboardsRoute'
import { DashboardEditorRoute } from './components/DashboardEditorRoute'
import { HouseOverviewRoute } from './components/HouseOverviewRoute'
import { SetupShell } from './components/setup/SetupShell'
import { DevicesList } from './components/setup/DevicesList'
import { DashboardsManager } from './components/setup/DashboardsManager'
import { SettingsPanel } from './components/setup/SettingsPanel'

export const router = createHashRouter([
  {
    element: <App />,
    children: [
      { index: true, element: <HouseOverviewRoute /> },
      { path: 'floor/:levelId', element: <ThreeDView /> },
      { path: 'house', element: <HouseOverviewRoute /> },
      { path: 'dashboards', element: <DashboardsRoute /> },
      {
        path: 'setup',
        element: <SetupShell />,
        children: [
          { index: true, element: <Navigate to="/setup/plans" replace /> },
          { path: 'plans', element: <AdminView /> },
          { path: 'plans/:levelId', element: <AdminView /> },
          { path: 'devices', element: <DevicesList /> },
          { path: 'dashboards', element: <DashboardsManager /> },
          { path: 'dashboards/:dashboardId', element: <DashboardEditorRoute /> },
          { path: 'settings', element: <SettingsPanel /> },
        ],
      },
      { path: '*', element: <ThreeDView /> },
    ],
  },
])
