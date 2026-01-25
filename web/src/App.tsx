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
import { CredentialManager } from './pages/CredentialManager'
import AIWorkflowBuilder from './pages/AIWorkflowBuilder'
import Marketplace from './pages/Marketplace'
import Analytics from './pages/Analytics'
import Documentation from './pages/Documentation'
import { OAuthConnections } from './pages/OAuthConnections'
import { OAuthCallback } from './components/oauth/OAuthCallback'
import { SSOSettings } from './pages/admin/SSOSettings'
import { AuditLogs } from './pages/admin/AuditLogs'
import Monitoring from './pages/Monitoring'
import { UserManagementPage } from './pages/admin/UserManagementPage'
import { TenantManagementPage } from './pages/admin/TenantManagementPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Dashboard />} />

        {/* Workflows - specific routes BEFORE dynamic :id route */}
        <Route path="workflows" element={<WorkflowList />} />
        <Route path="workflows/new" element={<WorkflowEditor />} />
        <Route path="workflows/:id" element={<WorkflowEditor />} />

        {/* Webhooks - specific routes BEFORE dynamic :id route */}
        <Route path="webhooks" element={<WebhookList />} />
        <Route path="webhooks/:id" element={<WebhookDetail />} />

        {/* Executions - specific routes BEFORE dynamic :id route */}
        <Route path="executions" element={<Executions />} />
        <Route path="executions/:id" element={<ExecutionDetail />} />

        {/* Schedules - specific routes BEFORE dynamic :id route */}
        <Route path="schedules" element={<Schedules />} />
        <Route path="schedules/new" element={<CreateSchedule />} />
        <Route path="schedules/:id/edit" element={<EditSchedule />} />

        {/* Other routes */}
        <Route path="credentials" element={<CredentialManager />} />
        <Route path="oauth/connections" element={<OAuthConnections />} />
        <Route path="ai/builder" element={<AIWorkflowBuilder />} />
        <Route path="marketplace" element={<Marketplace />} />
        <Route path="analytics" element={<Analytics />} />
        <Route path="monitoring" element={<Monitoring />} />
        <Route path="docs" element={<Documentation />} />

        {/* Admin routes */}
        <Route path="admin/users" element={<UserManagementPage />} />
        <Route path="admin/tenants" element={<TenantManagementPage />} />
        <Route path="admin/sso" element={<SSOSettings />} />
        <Route path="admin/audit-logs" element={<AuditLogs />} />
      </Route>
      {/* OAuth callback route outside Layout (no navigation) */}
      <Route path="oauth/callback/:provider" element={<OAuthCallback />} />
    </Routes>
  )
}

export default App
