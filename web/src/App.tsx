import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import WorkflowList from './pages/WorkflowList'
import WorkflowEditor from './pages/WorkflowEditor'
import Executions from './pages/Executions'
import ExecutionDetail from './pages/ExecutionDetail'
import Schedules from './pages/Schedules'
import CreateSchedule from './pages/CreateSchedule'
import EditSchedule from './pages/EditSchedule'
import WebhookList from './pages/WebhookList'
import WebhookDetail from './pages/WebhookDetail'
import CredentialManager from './pages/CredentialManager'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="workflows" element={<WorkflowList />} />
        <Route path="workflows/new" element={<WorkflowEditor />} />
        <Route path="workflows/:id" element={<WorkflowEditor />} />
        <Route path="webhooks" element={<WebhookList />} />
        <Route path="webhooks/:id" element={<WebhookDetail />} />
        <Route path="executions" element={<Executions />} />
        <Route path="executions/:id" element={<ExecutionDetail />} />
        <Route path="schedules" element={<Schedules />} />
        <Route path="schedules/new" element={<CreateSchedule />} />
        <Route path="schedules/:id/edit" element={<EditSchedule />} />
        <Route path="credentials" element={<CredentialManager />} />
      </Route>
    </Routes>
  )
}

export default App
