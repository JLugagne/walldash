import { createHashRouter } from 'react-router-dom'
import App from './App'
import { ThreeDView } from './components/ThreeDView'
import { AdminView } from './components/AdminView'
import { OverviewsRoute } from './components/OverviewsRoute'

export const router = createHashRouter([
  {
    element: <App />,
    children: [
      { index: true, element: <ThreeDView /> },
      { path: 'floor/:levelId', element: <ThreeDView /> },
      { path: 'admin', element: <AdminView /> },
      { path: 'admin/:levelId', element: <AdminView /> },
      { path: 'overviews', element: <OverviewsRoute /> },
      { path: '*', element: <ThreeDView /> },
    ],
  },
])